package main

import (
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/cmd/webhook"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/azdsecinfo"
	config "github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/config"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation/tivan"
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/infra/instrumentation/trace"
	"log"
	"os"
)

const (
	_configFileKey = "CONFIG_FILE"
)

// main is the entrypoint to AzureDefenderInClusterDefense .
func main() {
	configFile := os.Getenv(_configFileKey)
	if len(configFile) == 0{
		log.Fatalf("%v env variable is not defined.", _configFileKey)
	}
	// Load configuration
	AppConfig, err := config.LoadConfig(configFile)
	if err != nil {
		log.Fatal(err)
	}

	// Create Configuration objects for each factory
	managerConfiguration := new(webhook.ManagerConfiguration)
	certRotatorConfiguration := new(webhook.CertRotatorConfiguration)
	serverConfiguration := new(webhook.ServerConfiguration)
	handlerConfiguration := new(webhook.HandlerConfiguration)
	tivanInstrumentationConfiguration := new(tivan.TivanInstrumentationConfiguration)
	metricSubmitterConfiguration := new(tivan.MetricSubmitterConfiguration)
	tracerConfiguration := new(trace.TracerConfiguration)
	instrumentationConfiguration := new(instrumentation.InstrumentationProviderConfiguration)


	// Create a map between configuration object and key in main config file
	keyConfigMap := map[string]interface{} {
		"webhook.ManagerConfiguration": managerConfiguration,
		"webhook.CertRotatorConfiguration": certRotatorConfiguration,
		"webhook.ServerConfiguration": serverConfiguration,
		"webhook.HandlerConfiguration": handlerConfiguration,
		"instrumentation.tivan.TivanInstrumentationConfiguration": tivanInstrumentationConfiguration,
		"instrumentation.trace.TracerConfiguration": tracerConfiguration}

	for key, configObject := range keyConfigMap{
		// Unmarshal the relevant parts of appConfig's data to each of the configuration objects
		err = config.CreateSubConfiguration(AppConfig, key, configObject)
		if err != nil {
			log.Fatal("failed to load specific configuration data.", err)
		}
	}

	// Create Tivan's instrumentation
	tivanInstrumentationResult, err := tivan.NewTivanInstrumentationResult(tivanInstrumentationConfiguration)
	if err != nil {
		log.Fatal("main.NewTivanInstrumentationResult", err)
	}

	// Create factories
	tracerFactory := tivan.NewTracerFactory(tracerConfiguration, tivanInstrumentationResult.Tracer)
	metricSubmitterFactory := tivan.NewMetricSubmitterFactory(metricSubmitterConfiguration, &tivanInstrumentationResult.MetricSubmitter)
	instrumentationProviderFactory := instrumentation.NewInstrumentationProviderFactory(instrumentationConfiguration, tracerFactory, metricSubmitterFactory)
	instrumentationProvider, err := instrumentationProviderFactory.CreateInstrumentationProvider()
	if err != nil {
		log.Fatal("main.instrumentationProviderFactory.CreateInstrumentationProvider", err)
	}

	managerFactory := webhook.NewManagerFactory(managerConfiguration, instrumentationProvider)
	certRotatorFactory := webhook.NewCertRotatorFactory(certRotatorConfiguration)
	azdSecInfoProvider := azdsecinfo.NewAzdSecInfoProvider()
	handler := webhook.NewHandler(azdSecInfoProvider, handlerConfiguration, instrumentationProvider)
	serverFactory := webhook.NewServerFactory(serverConfiguration, managerFactory, certRotatorFactory, handler, instrumentationProvider)

	// Create Server
	server, err := serverFactory.CreateServer()
	if err != nil {
		log.Fatal("main.serverFactory.CreateServer", err)
	}
	// Run server
	if err = server.Run(); err != nil {
		log.Fatal("main.server.Run", err)
	}
}

