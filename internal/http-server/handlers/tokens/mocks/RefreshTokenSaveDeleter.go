// Code generated by mockery v2.49.1. DO NOT EDIT.

package mocks

import (
	context "context"

	mock "github.com/stretchr/testify/mock"
)

// RefreshTokenSaveDeleter is an autogenerated mock type for the RefreshTokenSaveDeleter type
type RefreshTokenSaveDeleter struct {
	mock.Mock
}

// DeleteRefreshToken provides a mock function with given fields: ctx, guid
func (_m *RefreshTokenSaveDeleter) DeleteRefreshToken(ctx context.Context, guid string) error {
	ret := _m.Called(ctx, guid)

	if len(ret) == 0 {
		panic("no return value specified for DeleteRefreshToken")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string) error); ok {
		r0 = rf(ctx, guid)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SaveRefreshToken provides a mock function with given fields: ctx, guid, refreshToken
func (_m *RefreshTokenSaveDeleter) SaveRefreshToken(ctx context.Context, guid string, refreshToken string) error {
	ret := _m.Called(ctx, guid, refreshToken)

	if len(ret) == 0 {
		panic("no return value specified for SaveRefreshToken")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string) error); ok {
		r0 = rf(ctx, guid, refreshToken)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// NewRefreshTokenSaveDeleter creates a new instance of RefreshTokenSaveDeleter. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewRefreshTokenSaveDeleter(t interface {
	mock.TestingT
	Cleanup(func())
}) *RefreshTokenSaveDeleter {
	mock := &RefreshTokenSaveDeleter{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
