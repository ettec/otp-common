// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/ettec/otp-common/api/staticdataservice (interfaces: StaticDataServiceClient)

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	reflect "reflect"

	staticdataservice "github.com/ettec/otp-common/api/staticdataservice"
	model "github.com/ettec/otp-common/model"
	gomock "github.com/golang/mock/gomock"
	grpc "google.golang.org/grpc"
)

// MockStaticDataServiceClient is a mock of StaticDataServiceClient interface.
type MockStaticDataServiceClient struct {
	ctrl     *gomock.Controller
	recorder *MockStaticDataServiceClientMockRecorder
}

// MockStaticDataServiceClientMockRecorder is the mock recorder for MockStaticDataServiceClient.
type MockStaticDataServiceClientMockRecorder struct {
	mock *MockStaticDataServiceClient
}

// NewMockStaticDataServiceClient creates a new mock instance.
func NewMockStaticDataServiceClient(ctrl *gomock.Controller) *MockStaticDataServiceClient {
	mock := &MockStaticDataServiceClient{ctrl: ctrl}
	mock.recorder = &MockStaticDataServiceClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockStaticDataServiceClient) EXPECT() *MockStaticDataServiceClientMockRecorder {
	return m.recorder
}

// GetListing mocks base method.
func (m *MockStaticDataServiceClient) GetListing(arg0 context.Context, arg1 *staticdataservice.ListingId, arg2 ...grpc.CallOption) (*model.Listing, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "GetListing", varargs...)
	ret0, _ := ret[0].(*model.Listing)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetListing indicates an expected call of GetListing.
func (mr *MockStaticDataServiceClientMockRecorder) GetListing(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetListing", reflect.TypeOf((*MockStaticDataServiceClient)(nil).GetListing), varargs...)
}

// GetListingMatching mocks base method.
func (m *MockStaticDataServiceClient) GetListingMatching(arg0 context.Context, arg1 *staticdataservice.ExactMatchParameters, arg2 ...grpc.CallOption) (*model.Listing, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "GetListingMatching", varargs...)
	ret0, _ := ret[0].(*model.Listing)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetListingMatching indicates an expected call of GetListingMatching.
func (mr *MockStaticDataServiceClientMockRecorder) GetListingMatching(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetListingMatching", reflect.TypeOf((*MockStaticDataServiceClient)(nil).GetListingMatching), varargs...)
}

// GetListings mocks base method.
func (m *MockStaticDataServiceClient) GetListings(arg0 context.Context, arg1 *staticdataservice.ListingIds, arg2 ...grpc.CallOption) (*staticdataservice.Listings, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "GetListings", varargs...)
	ret0, _ := ret[0].(*staticdataservice.Listings)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetListings indicates an expected call of GetListings.
func (mr *MockStaticDataServiceClientMockRecorder) GetListings(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetListings", reflect.TypeOf((*MockStaticDataServiceClient)(nil).GetListings), varargs...)
}

// GetListingsMatching mocks base method.
func (m *MockStaticDataServiceClient) GetListingsMatching(arg0 context.Context, arg1 *staticdataservice.MatchParameters, arg2 ...grpc.CallOption) (*staticdataservice.Listings, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "GetListingsMatching", varargs...)
	ret0, _ := ret[0].(*staticdataservice.Listings)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetListingsMatching indicates an expected call of GetListingsMatching.
func (mr *MockStaticDataServiceClientMockRecorder) GetListingsMatching(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetListingsMatching", reflect.TypeOf((*MockStaticDataServiceClient)(nil).GetListingsMatching), varargs...)
}

// GetListingsWithSameInstrument mocks base method.
func (m *MockStaticDataServiceClient) GetListingsWithSameInstrument(arg0 context.Context, arg1 *staticdataservice.ListingId, arg2 ...grpc.CallOption) (*staticdataservice.Listings, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "GetListingsWithSameInstrument", varargs...)
	ret0, _ := ret[0].(*staticdataservice.Listings)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetListingsWithSameInstrument indicates an expected call of GetListingsWithSameInstrument.
func (mr *MockStaticDataServiceClientMockRecorder) GetListingsWithSameInstrument(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetListingsWithSameInstrument", reflect.TypeOf((*MockStaticDataServiceClient)(nil).GetListingsWithSameInstrument), varargs...)
}
