// Code generated by mockery v2.9.4. DO NOT EDIT.

package mocks

import (
	admisionrequest "github.com/Azure/AzureDefender-K8S-InClusterDefense/cmd/webhook/admisionrequest"

	contracts "github.com/Azure/AzureDefender-K8S-InClusterDefense/pkg/azdsecinfo/contracts"

	mock "github.com/stretchr/testify/mock"
)

// IAzdSecInfoProviderCacheClient is an autogenerated mock type for the IAzdSecInfoProviderCacheClient type
type IAzdSecInfoProviderCacheClient struct {
	mock.Mock
}

// GetContainerVulnerabilityScanInfofromCache provides a mock function with given fields: podSpecCacheKey
func (_m *IAzdSecInfoProviderCacheClient) GetContainerVulnerabilityScanInfofromCache(podSpecCacheKey string) ([]*contracts.ContainerVulnerabilityScanInfo, error, error) {
	ret := _m.Called(podSpecCacheKey)

	var r0 []*contracts.ContainerVulnerabilityScanInfo
	if rf, ok := ret.Get(0).(func(string) []*contracts.ContainerVulnerabilityScanInfo); ok {
		r0 = rf(podSpecCacheKey)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*contracts.ContainerVulnerabilityScanInfo)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(podSpecCacheKey)
	} else {
		r1 = ret.Error(1)
	}

	var r2 error
	if rf, ok := ret.Get(2).(func(string) error); ok {
		r2 = rf(podSpecCacheKey)
	} else {
		r2 = ret.Error(2)
	}

	return r0, r1, r2
}

// GetPodSpecCacheKey provides a mock function with given fields: podSpec
func (_m *IAzdSecInfoProviderCacheClient) GetPodSpecCacheKey(podSpec *admisionrequest.PodSpec) string {
	ret := _m.Called(podSpec)

	var r0 string
	if rf, ok := ret.Get(0).(func(*admisionrequest.PodSpec) string); ok {
		r0 = rf(podSpec)
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// GetTimeOutStatus provides a mock function with given fields: podSpecCacheKey
func (_m *IAzdSecInfoProviderCacheClient) GetTimeOutStatus(podSpecCacheKey string) (int, error) {
	ret := _m.Called(podSpecCacheKey)

	var r0 int
	if rf, ok := ret.Get(0).(func(string) int); ok {
		r0 = rf(podSpecCacheKey)
	} else {
		r0 = ret.Get(0).(int)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(podSpecCacheKey)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ResetTimeOutInCacheAfterGettingScanResults provides a mock function with given fields: podSpecCacheKey
func (_m *IAzdSecInfoProviderCacheClient) ResetTimeOutInCacheAfterGettingScanResults(podSpecCacheKey string) error {
	ret := _m.Called(podSpecCacheKey)

	var r0 error
	if rf, ok := ret.Get(0).(func(string) error); ok {
		r0 = rf(podSpecCacheKey)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SetContainerVulnerabilityScanInfoInCache provides a mock function with given fields: podSpecCacheKey, containerVulnerabilityScanInfo, err
func (_m *IAzdSecInfoProviderCacheClient) SetContainerVulnerabilityScanInfoInCache(podSpecCacheKey string, containerVulnerabilityScanInfo []*contracts.ContainerVulnerabilityScanInfo, err error) error {
	ret := _m.Called(podSpecCacheKey, containerVulnerabilityScanInfo, err)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, []*contracts.ContainerVulnerabilityScanInfo, error) error); ok {
		r0 = rf(podSpecCacheKey, containerVulnerabilityScanInfo, err)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SetTimeOutStatusAfterEncounteredTimeout provides a mock function with given fields: podSpecCacheKey, timeOutStatus
func (_m *IAzdSecInfoProviderCacheClient) SetTimeOutStatusAfterEncounteredTimeout(podSpecCacheKey string, timeOutStatus int) error {
	ret := _m.Called(podSpecCacheKey, timeOutStatus)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, int) error); ok {
		r0 = rf(podSpecCacheKey, timeOutStatus)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
