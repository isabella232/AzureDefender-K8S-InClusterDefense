// Code generated by mockery v1.0.0. DO NOT EDIT.

package mocks

import (
	adal "github.com/Azure/go-autorest/autorest/adal"

	mock "github.com/stretchr/testify/mock"
)

// IBearerAuthorizer is an autogenerated mock type for the IBearerAuthorizer type
type IBearerAuthorizer struct {
	mock.Mock
}

// TokenProvider provides a mock function with given fields:
func (_m *IBearerAuthorizer) TokenProvider() adal.OAuthTokenProvider {
	ret := _m.Called()

	var r0 adal.OAuthTokenProvider
	if rf, ok := ret.Get(0).(func() adal.OAuthTokenProvider); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(adal.OAuthTokenProvider)
		}
	}

	return r0
}