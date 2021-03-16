// Code generated by MockGen. DO NOT EDIT.
// Source: /Users/anon/github/Fractapp/fractapp-server/adaptors/adaptor.go

// Package mocks is a generated GoMock package.
package mocks

import (
	types "fractapp-server/types"
	gomock "github.com/golang/mock/gomock"
	reflect "reflect"
)

// MockAdaptor is a mock of Adaptor interface
type MockAdaptor struct {
	ctrl     *gomock.Controller
	recorder *MockAdaptorMockRecorder
}

// MockAdaptorMockRecorder is the mock recorder for MockAdaptor
type MockAdaptorMockRecorder struct {
	mock *MockAdaptor
}

// NewMockAdaptor creates a new mock instance
func NewMockAdaptor(ctrl *gomock.Controller) *MockAdaptor {
	mock := &MockAdaptor{ctrl: ctrl}
	mock.recorder = &MockAdaptorMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockAdaptor) EXPECT() *MockAdaptorMockRecorder {
	return m.recorder
}

// Connect mocks base method
func (m *MockAdaptor) Connect() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Connect")
	ret0, _ := ret[0].(error)
	return ret0
}

// Connect indicates an expected call of Connect
func (mr *MockAdaptorMockRecorder) Connect() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Connect", reflect.TypeOf((*MockAdaptor)(nil).Connect))
}

// Subscribe mocks base method
func (m *MockAdaptor) Subscribe() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Subscribe")
	ret0, _ := ret[0].(error)
	return ret0
}

// Subscribe indicates an expected call of Subscribe
func (mr *MockAdaptorMockRecorder) Subscribe() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Subscribe", reflect.TypeOf((*MockAdaptor)(nil).Subscribe))
}

// Unsubscribe mocks base method
func (m *MockAdaptor) Unsubscribe() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Unsubscribe")
}

// Unsubscribe indicates an expected call of Unsubscribe
func (mr *MockAdaptorMockRecorder) Unsubscribe() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Unsubscribe", reflect.TypeOf((*MockAdaptor)(nil).Unsubscribe))
}

// WaitNewBlock mocks base method
func (m *MockAdaptor) WaitNewBlock() (uint64, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "WaitNewBlock")
	ret0, _ := ret[0].(uint64)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// WaitNewBlock indicates an expected call of WaitNewBlock
func (mr *MockAdaptorMockRecorder) WaitNewBlock() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "WaitNewBlock", reflect.TypeOf((*MockAdaptor)(nil).WaitNewBlock))
}

// Err mocks base method
func (m *MockAdaptor) Err() <-chan error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Err")
	ret0, _ := ret[0].(<-chan error)
	return ret0
}

// Err indicates an expected call of Err
func (mr *MockAdaptorMockRecorder) Err() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Err", reflect.TypeOf((*MockAdaptor)(nil).Err))
}

// Transfers mocks base method
func (m *MockAdaptor) Transfers(blockNumber uint64) ([]types.Tx, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Transfers", blockNumber)
	ret0, _ := ret[0].([]types.Tx)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Transfers indicates an expected call of Transfers
func (mr *MockAdaptorMockRecorder) Transfers(blockNumber interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Transfers", reflect.TypeOf((*MockAdaptor)(nil).Transfers), blockNumber)
}