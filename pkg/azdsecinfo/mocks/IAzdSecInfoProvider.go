// Code generated by mockery (devel). DO NOT EDIT.

package mocks

import (
	"github.com/Azure/AzureDefender-K8S-InClusterDefense/cmd/webhook/admisionrequest"
	contracts "github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/azdsecinfo/contracts"
	mock "github.com/stretchr/testify/mock"
)

// IAzdSecInfoProvider is an autogenerated mock type for the IAzdSecInfoProvider type
type IAzdSecInfoProvider struct {
	mock.Mock
}

// GetContainersVulnerabilityScanInfo provides a mock function with given fields: podSpec, resourceMetadata, resourceKind
func (_m *IAzdSecInfoProvider) GetContainersVulnerabilityScanInfo(workloadResource *admisionrequest.WorkloadResource) ([]*contracts.ContainerVulnerabilityScanInfo, error) {
	ret := _m.Called(workloadResource)

	var r0 []*contracts.ContainerVulnerabilityScanInfo
	if rf, ok := ret.Get(0).(func(workloadResource *admisionrequest.WorkloadResource) []*contracts.ContainerVulnerabilityScanInfo); ok {
		r0 = rf(workloadResource)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*contracts.ContainerVulnerabilityScanInfo)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(workloadResource *admisionrequest.WorkloadResource) error); ok {
		r1 = rf(workloadResource)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
