// Code generated by MockGen. DO NOT EDIT.
// Source: /Users/anon/github/Fractapp/fractapp-server/db/db.go

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	db "fractapp-server/db"
	notification "fractapp-server/notification"
	gomock "github.com/golang/mock/gomock"
	reflect "reflect"
)

// MockDB is a mock of DB interface
type MockDB struct {
	ctrl     *gomock.Controller
	recorder *MockDBMockRecorder
}

// MockDBMockRecorder is the mock recorder for MockDB
type MockDBMockRecorder struct {
	mock *MockDB
}

// NewMockDB creates a new mock instance
func NewMockDB(ctrl *gomock.Controller) *MockDB {
	mock := &MockDB{ctrl: ctrl}
	mock.recorder = &MockDBMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockDB) EXPECT() *MockDBMockRecorder {
	return m.recorder
}

// SubscribersCountByToken mocks base method
func (m *MockDB) SubscribersCountByToken(token string) (int, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SubscribersCountByToken", token)
	ret0, _ := ret[0].(int)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// SubscribersCountByToken indicates an expected call of SubscribersCountByToken
func (mr *MockDBMockRecorder) SubscribersCountByToken(token interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SubscribersCountByToken", reflect.TypeOf((*MockDB)(nil).SubscribersCountByToken), token)
}

// SubscriberByAddress mocks base method
func (m *MockDB) SubscriberByAddress(address string) (*db.Subscriber, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SubscriberByAddress", address)
	ret0, _ := ret[0].(*db.Subscriber)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// SubscriberByAddress indicates an expected call of SubscriberByAddress
func (mr *MockDBMockRecorder) SubscriberByAddress(address interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SubscriberByAddress", reflect.TypeOf((*MockDB)(nil).SubscriberByAddress), address)
}

// AuthByValue mocks base method
func (m *MockDB) AuthByValue(value string, codeType notification.NotificatorType, checkType notification.CheckType) (*db.Auth, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AuthByValue", value, codeType, checkType)
	ret0, _ := ret[0].(*db.Auth)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// AuthByValue indicates an expected call of AuthByValue
func (mr *MockDBMockRecorder) AuthByValue(value, codeType, checkType interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AuthByValue", reflect.TypeOf((*MockDB)(nil).AuthByValue), value, codeType, checkType)
}

// AddressesById mocks base method
func (m *MockDB) AddressesById(id string) ([]db.Address, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AddressesById", id)
	ret0, _ := ret[0].([]db.Address)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// AddressesById indicates an expected call of AddressesById
func (mr *MockDBMockRecorder) AddressesById(id interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddressesById", reflect.TypeOf((*MockDB)(nil).AddressesById), id)
}

// ProfileById mocks base method
func (m *MockDB) ProfileById(id string) (*db.Profile, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ProfileById", id)
	ret0, _ := ret[0].(*db.Profile)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ProfileById indicates an expected call of ProfileById
func (mr *MockDBMockRecorder) ProfileById(id interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ProfileById", reflect.TypeOf((*MockDB)(nil).ProfileById), id)
}

// ProfileByAddress mocks base method
func (m *MockDB) ProfileByAddress(address string) (*db.Profile, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ProfileByAddress", address)
	ret0, _ := ret[0].(*db.Profile)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ProfileByAddress indicates an expected call of ProfileByAddress
func (mr *MockDBMockRecorder) ProfileByAddress(address interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ProfileByAddress", reflect.TypeOf((*MockDB)(nil).ProfileByAddress), address)
}

// ProfileByUsername mocks base method
func (m *MockDB) ProfileByUsername(username string) (*db.Profile, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ProfileByUsername", username)
	ret0, _ := ret[0].(*db.Profile)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ProfileByUsername indicates an expected call of ProfileByUsername
func (mr *MockDBMockRecorder) ProfileByUsername(username interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ProfileByUsername", reflect.TypeOf((*MockDB)(nil).ProfileByUsername), username)
}

// AddressIsExist mocks base method
func (m *MockDB) AddressIsExist(address string) (bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AddressIsExist", address)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// AddressIsExist indicates an expected call of AddressIsExist
func (mr *MockDBMockRecorder) AddressIsExist(address interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddressIsExist", reflect.TypeOf((*MockDB)(nil).AddressIsExist), address)
}

// UsernameIsExist mocks base method
func (m *MockDB) UsernameIsExist(username string) (bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UsernameIsExist", username)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// UsernameIsExist indicates an expected call of UsernameIsExist
func (mr *MockDBMockRecorder) UsernameIsExist(username interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UsernameIsExist", reflect.TypeOf((*MockDB)(nil).UsernameIsExist), username)
}

// SearchUsersByUsername mocks base method
func (m *MockDB) SearchUsersByUsername(value string, limit int) ([]db.Profile, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SearchUsersByUsername", value, limit)
	ret0, _ := ret[0].([]db.Profile)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// SearchUsersByUsername indicates an expected call of SearchUsersByUsername
func (mr *MockDBMockRecorder) SearchUsersByUsername(value, limit interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SearchUsersByUsername", reflect.TypeOf((*MockDB)(nil).SearchUsersByUsername), value, limit)
}

// SearchUsersByEmail mocks base method
func (m *MockDB) SearchUsersByEmail(value string) (*db.Profile, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SearchUsersByEmail", value)
	ret0, _ := ret[0].(*db.Profile)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// SearchUsersByEmail indicates an expected call of SearchUsersByEmail
func (mr *MockDBMockRecorder) SearchUsersByEmail(value interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SearchUsersByEmail", reflect.TypeOf((*MockDB)(nil).SearchUsersByEmail), value)
}

// ProfileByPhoneNumber mocks base method
func (m *MockDB) ProfileByPhoneNumber(contactPhoneNumber, myPhoneNumber string) (*db.Profile, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ProfileByPhoneNumber", contactPhoneNumber, myPhoneNumber)
	ret0, _ := ret[0].(*db.Profile)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ProfileByPhoneNumber indicates an expected call of ProfileByPhoneNumber
func (mr *MockDBMockRecorder) ProfileByPhoneNumber(contactPhoneNumber, myPhoneNumber interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ProfileByPhoneNumber", reflect.TypeOf((*MockDB)(nil).ProfileByPhoneNumber), contactPhoneNumber, myPhoneNumber)
}

// CreateProfile mocks base method
func (m *MockDB) CreateProfile(ctx context.Context, profile *db.Profile, addresses []*db.Address) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateProfile", ctx, profile, addresses)
	ret0, _ := ret[0].(error)
	return ret0
}

// CreateProfile indicates an expected call of CreateProfile
func (mr *MockDBMockRecorder) CreateProfile(ctx, profile, addresses interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateProfile", reflect.TypeOf((*MockDB)(nil).CreateProfile), ctx, profile, addresses)
}

// IdByToken mocks base method
func (m *MockDB) IdByToken(token string) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "IdByToken", token)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// IdByToken indicates an expected call of IdByToken
func (mr *MockDBMockRecorder) IdByToken(token interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "IdByToken", reflect.TypeOf((*MockDB)(nil).IdByToken), token)
}

// TokenById mocks base method
func (m *MockDB) TokenById(id string) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "TokenById", id)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// TokenById indicates an expected call of TokenById
func (mr *MockDBMockRecorder) TokenById(id interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "TokenById", reflect.TypeOf((*MockDB)(nil).TokenById), id)
}

// AllContacts mocks base method
func (m *MockDB) AllContacts(id string) ([]db.Contact, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AllContacts", id)
	ret0, _ := ret[0].([]db.Contact)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// AllContacts indicates an expected call of AllContacts
func (mr *MockDBMockRecorder) AllContacts(id interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AllContacts", reflect.TypeOf((*MockDB)(nil).AllContacts), id)
}

// AllMatchContacts mocks base method
func (m *MockDB) AllMatchContacts(id, phoneNumber string) ([]db.Profile, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AllMatchContacts", id, phoneNumber)
	ret0, _ := ret[0].([]db.Profile)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// AllMatchContacts indicates an expected call of AllMatchContacts
func (mr *MockDBMockRecorder) AllMatchContacts(id, phoneNumber interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AllMatchContacts", reflect.TypeOf((*MockDB)(nil).AllMatchContacts), id, phoneNumber)
}

// SubscribersByRange mocks base method
func (m *MockDB) SubscribersByRange(from, limit int) ([]db.Subscriber, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SubscribersByRange", from, limit)
	ret0, _ := ret[0].([]db.Subscriber)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// SubscribersByRange indicates an expected call of SubscribersByRange
func (mr *MockDBMockRecorder) SubscribersByRange(from, limit interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SubscribersByRange", reflect.TypeOf((*MockDB)(nil).SubscribersByRange), from, limit)
}

// SubscribersCount mocks base method
func (m *MockDB) SubscribersCount() (int, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SubscribersCount")
	ret0, _ := ret[0].(int)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// SubscribersCount indicates an expected call of SubscribersCount
func (mr *MockDBMockRecorder) SubscribersCount() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SubscribersCount", reflect.TypeOf((*MockDB)(nil).SubscribersCount))
}

// Insert mocks base method
func (m *MockDB) Insert(value interface{}) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Insert", value)
	ret0, _ := ret[0].(error)
	return ret0
}

// Insert indicates an expected call of Insert
func (mr *MockDBMockRecorder) Insert(value interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Insert", reflect.TypeOf((*MockDB)(nil).Insert), value)
}

// UpdateByPK mocks base method
func (m *MockDB) UpdateByPK(value interface{}) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateByPK", value)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdateByPK indicates an expected call of UpdateByPK
func (mr *MockDBMockRecorder) UpdateByPK(value interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateByPK", reflect.TypeOf((*MockDB)(nil).UpdateByPK), value)
}

// Update mocks base method
func (m *MockDB) Update(value interface{}, condition string, params ...interface{}) error {
	m.ctrl.T.Helper()
	varargs := []interface{}{value, condition}
	for _, a := range params {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "Update", varargs...)
	ret0, _ := ret[0].(error)
	return ret0
}

// Update indicates an expected call of Update
func (mr *MockDBMockRecorder) Update(value, condition interface{}, params ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{value, condition}, params...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Update", reflect.TypeOf((*MockDB)(nil).Update), varargs...)
}