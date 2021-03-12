// Code generated by MockGen. DO NOT EDIT.
// Source: /Users/anon/github/Fractapp/fractapp-server/firebase/notificator.go

// Package mocks is a generated GoMock package.
package mocks

import (
	firebase "fractapp-server/firebase"
	types "fractapp-server/types"
	gomock "github.com/golang/mock/gomock"
	reflect "reflect"
)

// MockTxNotificator is a mock of TxNotificator interface
type MockTxNotificator struct {
	ctrl     *gomock.Controller
	recorder *MockTxNotificatorMockRecorder
}

// MockTxNotificatorMockRecorder is the mock recorder for MockTxNotificator
type MockTxNotificatorMockRecorder struct {
	mock *MockTxNotificator
}

// NewMockTxNotificator creates a new mock instance
func NewMockTxNotificator(ctrl *gomock.Controller) *MockTxNotificator {
	mock := &MockTxNotificator{ctrl: ctrl}
	mock.recorder = &MockTxNotificatorMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockTxNotificator) EXPECT() *MockTxNotificatorMockRecorder {
	return m.recorder
}

// Notify mocks base method
func (m *MockTxNotificator) Notify(msg, token string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Notify", msg, token)
	ret0, _ := ret[0].(error)
	return ret0
}

// Notify indicates an expected call of Notify
func (mr *MockTxNotificatorMockRecorder) Notify(msg, token interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Notify", reflect.TypeOf((*MockTxNotificator)(nil).Notify), msg, token)
}

// Msg mocks base method
func (m *MockTxNotificator) Msg(member string, txType firebase.TxType, amount float64, currency types.Currency) string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Msg", member, txType, amount, currency)
	ret0, _ := ret[0].(string)
	return ret0
}

// Msg indicates an expected call of Msg
func (mr *MockTxNotificatorMockRecorder) Msg(member, txType, amount, currency interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Msg", reflect.TypeOf((*MockTxNotificator)(nil).Msg), member, txType, amount, currency)
}