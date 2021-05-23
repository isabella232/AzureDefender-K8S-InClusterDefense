package main

import (
	"encoding/json"
	"fmt"

	// "io/ioutil"
	//"net/http"

	// "path"
	// "regexp"
	// "strings"

	log "github.com/sirupsen/logrus"
	"golang.org/x/net/context"

	// yaml "gopkg.in/yaml.v2"
	// azsecurity "github.com/Azure/azure-sdk-for-go/services/preview/security/mgmt/v3.0/security"
	azresourcegraph "github.com/Azure/azure-sdk-for-go/services/resourcegraph/mgmt/2019-04-01/resourcegraph"

	//acrmgmt "github.com/Azure/azure-sdk-for-go/services/preview/containerregistry/mgmt/2018-02-01/containerregistry"
	//acr "github.com/Azure/azure-sdk-for-go/services/preview/containerregistry/runtime/2019-08-15-preview/containerregistry"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/adal"
	"github.com/Azure/go-autorest/autorest/azure"

	"github.com/pkg/errors"

	"flag"
	"net/http"
	"strings"
	"time"
)

// auth
const (
	// OAuthGrantTypeServicePrincipal for client credentials flow
	OAuthGrantTypeServicePrincipal OAuthGrantType = iota
	cloudName                      string         = ""
)

const (
	timeout = 80 * time.Second
)

var (
	unscannedImage string = "Unscanned"
)

var (
	period           time.Duration
	resourceName     string
	subscriptionID   string
	resourceGroup    string
	identityClientID string
	isInitialized    bool
)

// OAuthGrantType specifies which grant type to use.
type OAuthGrantType int

// AuthGrantType ...
func AuthGrantType() OAuthGrantType {
	return OAuthGrantTypeServicePrincipal
}

// Server
type Server struct {
	// subscriptionId to azure
	SubscriptionID string
	// tenantID in AAD
	TenantID string
	// AAD app client secret (if not using POD AAD Identity)
	AADClientSecret string
	// AAD app client secret id (if not using POD AAD Identity)
	AADClientID string
	// Location of security center
	Location string
	// Scope of assessment
	Scope string
}

// ScanInfo
type ScanInfo struct {
	//ScanStatus
	ImageDigest *string `json:"imageDigest,omitempty"`
	//ScanStatus
	ScanStatus *string `json:"scanStatus,omitempty"`
	// SeveritySummary
	SeveritySummary map[string]float64 `json:"severitySummary,omitempty"`
}

// NewServer creates a new server instance.
func NewServer() (*Server, error) {
	log.Debugf("NewServer")
	var s Server
	s.SubscriptionID = "409111bf-3097-421c-ad68-a44e716edf58" // os.Getenv("SUBSCRIPTION_ID")

	return &s, nil
}

// ParseAzureEnvironment returns azure environment by name
func ParseAzureEnvironment(cloudName string) (*azure.Environment, error) {
	var env azure.Environment
	var err error
	if cloudName == "" {
		env = azure.PublicCloud
	} else {
		env, err = azure.EnvironmentFromName(cloudName)
	}
	return &env, err
}

func (s *Server) Process(ctx context.Context, image Image) (resps []ScanInfo, err error) {
	if image.digest == "" {
		return nil, fmt.Errorf("Failed to provide digest to query")
	}
	// Connect to ARG:
	myClient := azresourcegraph.New()
	token, tokenErr := getTokenMSI()
	if tokenErr != nil {
		return nil, errors.Wrapf(tokenErr, "failed to get management token")
	}
	myClient.Authorizer = autorest.NewBearerAuthorizer(token)

	// Generate Query
	query := s.generateQuery(image.digest)
	// Execute Query
	results, err := myClient.Resources(ctx, query)
	if err != nil {
		return nil, err
	}
	//Parse query response:
	resps, err2 := s.parseQueryResponse(results)
	if err2 != nil {
		log.Debug(err2.Error())
		return nil, err2
	}

	return resps, nil
}

func (s *Server) parseQueryResponse(results azresourcegraph.QueryResponse) (scanInfoList []ScanInfo, err error) {
	log.Debugf("results: %d", results)

	var data []ScanInfo
	count := *results.Count
	log.Debugf("total unhealthy images: %d", count)

	scanInfoList = make([]ScanInfo, 0)
	// In case that scan info returned from ARG.
	if count > 0 {
		raw, err := json.Marshal(results.Data)
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal(raw, &data)
		if err != nil {
			return nil, err
		}
		log.Debugf("Data: %d", data)
		for _, v := range data {
			oneScanInfo := ScanInfo{
				ScanStatus:      v.ScanStatus,
				SeveritySummary: v.SeveritySummary,
			}
			scanInfoList = append(scanInfoList, oneScanInfo)
		}
		// In Caste that there are no results.
	} else {
		oneScanInfo := ScanInfo{
			ScanStatus:      &unscannedImage,
			SeveritySummary: nil,
		}
		scanInfoList = append(scanInfoList, oneScanInfo)
	}

	log.Debugf("scanInfoList: %s", scanInfoList)

	return scanInfoList, nil
}

func (s *Server) generateQuery(digest string) azresourcegraph.QueryRequest {
	// Prepare query:
	subs := []string{s.SubscriptionID}
	rawQuery := `
		securityresources
		| where type == 'microsoft.security/assessments/subassessments'
		| where id matches regex '(.+?)/providers/Microsoft.Security/assessments/dbd0cb49-b563-45e7-9724-889e799fa648/'
		//| parse id with registryResourceId '/providers/Microsoft.Security/assessments/' *
		//| parse registryResourceId with * "/providers/Microsoft.ContainerRegistry/registries/" registryName
		| extend imageDigest = tostring(properties.additionalData.imageDigest)
		| where imageDigest == "` + digest + `"
		| extend repository = tostring(properties.additionalData.repositoryName)
		| extend scanFindingSeverity = tostring(properties.status.severity), scanStatus = tostring(properties.status.code)
		| summarize scanFindingSeverityCount = count() by scanFindingSeverity, scanStatus, repository, imageDigest
		| summarize severitySummary = make_bag(pack(scanFindingSeverity, scanFindingSeverityCount)) by  imageDigest, scanStatus`

	log.Debugf("Query: %s", rawQuery)
	options := azresourcegraph.QueryRequestOptions{ResultFormat: azresourcegraph.ResultFormatObjectArray}
	query := azresourcegraph.QueryRequest{
		Subscriptions: &subs,
		Query:         &rawQuery,
		Options:       &options,
	}
	return query
}

func getTokenMSI() (msg *adal.Token, err error) {
	if !isInitialized {
		flag.DurationVar(&period, "period", 100*time.Second, "The period that the demo is being executed")
		flag.StringVar(&resourceName, "resource-name", "https://management.azure.com/", "The resource name to grant the access token")
		flag.StringVar(&subscriptionID, "subscription-id", "", "The Azure subscription ID")
		flag.StringVar(&resourceGroup, "resource-group", "", "The resource group name which the user-assigned identity read access to")
		flag.StringVar(&identityClientID, "identity-client-id", "", "The user-assigned identity client ID")
		flag.Parse()
		isInitialized = true
	}

	imdsTokenEndpoint, err := adal.GetMSIVMEndpoint()
	if err != nil {
		log.Errorf("failed to get IMDS token endpoint, error: %+v", err)
	}

	ticker := time.NewTicker(period)
	defer ticker.Stop()
	for ; true; <-ticker.C {
		curlIMDSMetadataInstanceEndpoint()
		t1 := getTokenFromIMDS(imdsTokenEndpoint)
		t2 := getTokenFromIMDSWithUserAssignedID(imdsTokenEndpoint)
		if t1 == nil || t2 == nil || !strings.EqualFold(t1.AccessToken, t2.AccessToken) {
			log.Error("Tokens acquired from IMDS with and without identity client ID do not match")
		} else {
			// log.Infof("Try decoding your token %s at https://jwt.io", t1.AccessToken)
			return t1, nil
		}
	}
	return nil, nil
}

func getTokenFromIMDS(imdsTokenEndpoint string) *adal.Token {
	spt, err := adal.NewServicePrincipalTokenFromMSIWithUserAssignedID(imdsTokenEndpoint, resourceName, identityClientID)
	if err != nil {
		log.Errorf("failed to acquire a token from IMDS using user-assigned identity, error: %+v", err)
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	if err := spt.RefreshWithContext(ctx); err != nil {
		log.Errorf("failed to refresh the service principal token, error: %+v", err)
		return nil
	}

	token := spt.Token()
	if token.IsZero() {
		log.Errorf("%+v is a zero token", token)
		return nil
	}

	log.Infof("successfully acquired a service principal token from %s", imdsTokenEndpoint)
	return &token
}

func getTokenFromIMDSWithUserAssignedID(imdsTokenEndpoint string) *adal.Token {
	spt, err := adal.NewServicePrincipalTokenFromMSIWithUserAssignedID(imdsTokenEndpoint, resourceName, identityClientID)
	if err != nil {
		log.Errorf("failed to acquire a token from IMDS using user-assigned identity, error: %+v", err)
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	if err := spt.RefreshWithContext(ctx); err != nil {
		log.Errorf("failed to refresh the service principal token, error: %+v", err)
		return nil
	}

	token := spt.Token()
	if token.IsZero() {
		log.Errorf("%+v is a zero token", token)
		return nil
	}

	log.Info("successfully acquired a service principal token from %s using a user-assigned identity (%s)", imdsTokenEndpoint, identityClientID)
	return &token
}

func curlIMDSMetadataInstanceEndpoint() {
	client := &http.Client{
		Timeout: timeout,
	}
	req, err := http.NewRequest("GET", "http://169.254.169.254/metadata/instance?api-version=2017-08-01", nil)
	if err != nil {
		log.Errorf("failed to create a new HTTP request, error: %+v", err)
		return
	}
	req.Header.Add("Metadata", "true")

	resp, err := client.Do(req)
	if err != nil {
		log.Errorf("%s", err)
		return
	}
	defer resp.Body.Close()
}
