package admisionrequest

import (
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

const (
	_errMsgObjectNotFound           = "admisionrequest.extractor: request did not include object"
	_errMsgInvalidAdmission         = "admisionrequest.extractor: admission request was nil"
	_errMsgJsonToYamlConversionFail = "admisionrequest.extractor: failed to convert json to yaml node"
	_errMsgInvalidPath              = "admisionrequest.extractor: failed to access the given path"
	_errMsgUnexpectedResource       = "admisionrequest.extractor: expected workload resource"
	imagePullSecretsConst           = "imagePullSecrets"
	metadataConst                   = "metadata"
	ownerReferencesConst            = "ownerReferences"
	containersConst                 = "containers"
	initContainersConst             = "initContainers"
	serviceAccountNameConst         = "serviceAccountName"
	NameConst                       = "name"
	KindConst                       = "kind"
    ApiVersionConst                 = "apiVersion"
)

var (
	//from yaml.ConventionalContainersPaths, without containers(the last string in paths).
	conventionalPodSpecPaths = [][]string{
		{"spec", "jobTemplate", "spec", "template", "spec"},
		{"spec", "template", "spec"},
		{"template", "spec"},
		{"spec"}}
	kubernetesWorkloadResources = []string{"Pod", "Deployment", "ReplicaSet", "StatefulSet", "DaemonSet",
		"Job", "CronJob", "ReplicationController"} //https://kubernetes.io/docs/concepts/workloads/
	_errInvalidAdmission   = errors.New(_errMsgInvalidAdmission)
	_errObjectNotFound     = errors.New(_errMsgObjectNotFound)
	_errUnexpectedResource = errors.New(_errMsgUnexpectedResource)
)


func getContainers(specNode *yaml.RNode) (containers []Container, initContainers []Container, err error) {
	allContainers := make([][]Container,2)
	for pathIndex, containerPath := range []string{containersConst, initContainersConst} {
		containersInterface, err := specNode.GetSlice(containerPath)
		// if err != nil it means that containerType is an empty field in admission request
		if err != nil {
			allContainers[pathIndex] = nil
			break
		}
		allContainers[pathIndex] = make([]Container, len(containersInterface))
		for i, containerObj := range containersInterface {
			v, ok := containerObj.(map[string]interface{})
			if ok == false {
				return nil, nil, err
			}
			allContainers[pathIndex][i].Image, ok = (v["image"]).(string)
			if ok == false {
				return nil, nil, err
			}
			allContainers[pathIndex][i].Name, ok = (v["name"]).(string)
			if ok == false {
				return nil, nil, err
			}
		}
	}
	return allContainers[0], allContainers[1], nil
}

// getImagePullSecrets returns workload kubernetes resource's image pull secrets.
func getImagePullSecrets(specRoot *yaml.RNode) (secrets []corev1.LocalObjectReference, err error) {
	imagePullSecretsInterface, pathErr := specRoot.GetSlice(imagePullSecretsConst)
	// if pathErr != nil it means that "imagePullSecrets" is an empty field in admission request
	if pathErr != nil {
		return nil, nil
	}

	secrets = make([]corev1.LocalObjectReference, len(imagePullSecretsInterface))
	for i, secret := range imagePullSecretsInterface {
		v, ok := secret.(map[string]interface{})
		if ok == false {
			return nil, err
		}
		secrets[i].Name, ok = (v[NameConst]).(string)
		if ok == false {
			return nil, err
		}
	}
	return secrets, nil
}

// GetOwnerReference returns workload kubernetes resource's owner reference.
func (extractor *Extractor) getOwnerReference(root *yaml.RNode) (ownerReferences []OwnerReference, err error) {
	metaNode, pathErr := goToDestNode(root, metadataConst)
	// if err != nil it means that "ownerReferences" is an empty field in admission request
	if pathErr != nil {
		return nil, nil
	}

	sliceOwnerReferences, err := metaNode.GetSlice(ownerReferencesConst)
	if err != nil {
		return nil, nil
	}
	ownerReferences = make([]OwnerReference,len(sliceOwnerReferences))
	for i,reference := range sliceOwnerReferences{
		mapReference,ok := reference.(map[string]interface{})
		if ok == false {
			return nil, err
		}
		ownerReferences[i].APIVersion,ok = mapReference[ApiVersionConst].(string)
		if ok == false {
			return nil, err
		}
		ownerReferences[i].Kind,ok = mapReference[KindConst].(string)
		if ok == false {
			return nil, err
		}
		ownerReferences[i].Name,ok = mapReference[NameConst].(string)
		if ok == false {
			return nil, err
		}

	}
	return ownerReferences, nil
}


//Basics checks of the application admission request.
func reqBasicChecks(req *admission.Request) (err error) {
	if req == nil {
		return _errInvalidAdmission
	}
	if len(req.Object.Raw) == 0 {
		return _errObjectNotFound
	}
	if !stringInSlice(req.Kind.Kind, kubernetesWorkloadResources) {
		return _errUnexpectedResource
	}
	return nil
}

// ExtractMetadataFromAdmissionRequest return *ObjectMetadata object according
//// to the information in yamlFile.
func (extractor *Extractor) ExtractMetadataFromAdmissionRequest(root *yaml.RNode) (metadata *ObjectMetadata,err error) {
	tracer := extractor.tracerProvider.GetTracer("ExtractMetadataFromAdmissionRequest")
	name := root.GetName()
	namespace := root.GetNamespace()
	annotation := root.GetAnnotations()
	if len(annotation)==0{
		annotation = nil
	}
	ownerReferences, err := extractor.getOwnerReference(root)
	if err != nil {
		tracer.Error(err, "")
		return nil, err
	}
	meta := newObjectMetadata(name, namespace, annotation, ownerReferences)
	return &meta, nil
}

// ExtractSpecFromAdmissionRequest return *PodSpec object according
//// to the information in yamlFile.
func (extractor *Extractor) ExtractSpecFromAdmissionRequest(root *yaml.RNode) (spec *PodSpec,err error) {
	tracer := extractor.tracerProvider.GetTracer("ExtractSpecFromAdmissionRequest")
	// return podspec yaml rNode.
	specNode, err := yaml.LookupFirstMatch(conventionalPodSpecPaths).Filter(root)
	if err != nil{
		// spec is optional
		podSpec := newSpec(nil, nil, nil, "")
		return &podSpec,nil
	}
	containerList, initContainerList, err := getContainers(specNode)
	if err != nil {
		tracer.Error(err, "")
		return nil, err
	}
	imagePullSecrets, err := getImagePullSecrets(specNode)
	if err != nil {
		tracer.Error(err, "")
		return nil, err
	}

	// err ignore because serviceAccountName may not exist. in that case ot will
	// be assigned to empty string.
	serviceAccountName, err := specNode.GetString(serviceAccountNameConst)

	podSpec := newSpec(containerList, initContainerList, imagePullSecrets, serviceAccountName)
	return &podSpec, nil
}

// ExtractWorkloadResourceFromAdmissionRequest return WorkloadResource object according
// to the information in admission.Request.
func (extractor *Extractor) ExtractWorkloadResourceFromAdmissionRequest(req *admission.Request) (resource *WorkloadResource, err error) {
	tracer := extractor.tracerProvider.GetTracer("ExtractWorkloadResourceFromAdmissionRequest")
	tracer.Info("ExtractWorkloadResourceFromAdmissionRequest Enter", "admission request", req)

	err = reqBasicChecks(req)
	if err != nil {
		tracer.Error(err, "")
		return nil, err
	}
	yamlFile, err := yaml.ConvertJSONToYamlNode(string(req.Object.Raw))
	if err != nil {
		tracer.Error(errors.Wrap(err, _errMsgJsonToYamlConversionFail), "")
		return nil, errors.Wrap(err, _errMsgJsonToYamlConversionFail)
	}

	metadata, err := extractor.ExtractMetadataFromAdmissionRequest(yamlFile)
	if err != nil {
		return nil, err
	}

	spec, err := extractor.ExtractSpecFromAdmissionRequest(yamlFile)
	if err != nil {
		return nil, err
	}

	return newWorkLoadResource(*metadata, *spec), nil
}
