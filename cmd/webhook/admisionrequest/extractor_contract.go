package admisionrequest

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ObjectMetadata represents the metadata of WorkloadResource object.
type ObjectMetadata struct{
	Namespace string
	Annotation map[string]string
	OwnerReferences []metav1.OwnerReference
}

// Container represents container object.
type Container struct{
	Name string
	Image string
}

// PodSpec represents a specification of the desired behavior of the WorkloadResource.
type PodSpec struct{
	Containers []Container
	InitContainers []Container
	ImagePullSecrets []corev1.LocalObjectReference
	ServiceAccountName string
}

// WorkloadResource represents an abstraction of a kubernetes workload resources such as:
// Pod, Deployments, ReplicaSet, StatefulSets, DaemonSet, Jobs, CronJob and ReplicationController.
type WorkloadResource struct{
	Metadata ObjectMetadata
	Spec     PodSpec
}
