// Code generated by MockGen. DO NOT EDIT.
// Source: /Users/anon/github/Fractapp/fractapp-server/db/db.go

// Package mocks is a generated GoMock package.
package mocks

import (
	db "fractapp-server/db"
	notification "fractapp-server/notification"
	types "fractapp-server/types"
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

// AuthByValue mocks base method
func (m *MockDB) AuthByValue(value string, codeType notification.NotificatorType) (*db.Auth, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AuthByValue", value, codeType)
	ret0, _ := ret[0].(*db.Auth)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// AuthByValue indicates an expected call of AuthByValue
func (mr *MockDBMockRecorder) AuthByValue(value, codeType interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AuthByValue", reflect.TypeOf((*MockDB)(nil).AuthByValue), value, codeType)
}

// AllContacts mocks base method
func (m *MockDB) AllContacts(profileId db.ID) ([]db.Contact, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AllContacts", profileId)
	ret0, _ := ret[0].([]db.Contact)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// AllContacts indicates an expected call of AllContacts
func (mr *MockDBMockRecorder) AllContacts(profileId interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AllContacts", reflect.TypeOf((*MockDB)(nil).AllContacts), profileId)
}

// AllMatchContacts mocks base method
func (m *MockDB) AllMatchContacts(id db.ID) ([]db.Profile, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AllMatchContacts", id)
	ret0, _ := ret[0].([]db.Profile)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// AllMatchContacts indicates an expected call of AllMatchContacts
func (mr *MockDBMockRecorder) AllMatchContacts(id interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AllMatchContacts", reflect.TypeOf((*MockDB)(nil).AllMatchContacts), id)
}

// MessagesByReceiver mocks base method
func (m *MockDB) MessagesByReceiver(receiver db.ID) ([]db.Message, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "MessagesByReceiver", receiver)
	ret0, _ := ret[0].([]db.Message)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// MessagesByReceiver indicates an expected call of MessagesByReceiver
func (mr *MockDBMockRecorder) MessagesByReceiver(receiver interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "MessagesByReceiver", reflect.TypeOf((*MockDB)(nil).MessagesByReceiver), receiver)
}

// MessagesBySenderAndReceiver mocks base method
func (m *MockDB) MessagesBySenderAndReceiver(sender, receiver db.ID) ([]db.Message, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "MessagesBySenderAndReceiver", sender, receiver)
	ret0, _ := ret[0].([]db.Message)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// MessagesBySenderAndReceiver indicates an expected call of MessagesBySenderAndReceiver
func (mr *MockDBMockRecorder) MessagesBySenderAndReceiver(sender, receiver interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "MessagesBySenderAndReceiver", reflect.TypeOf((*MockDB)(nil).MessagesBySenderAndReceiver), sender, receiver)
}

// SetDelivered mocks base method
func (m *MockDB) SetDelivered(owner, id db.ID) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SetDelivered", owner, id)
	ret0, _ := ret[0].(error)
	return ret0
}

// SetDelivered indicates an expected call of SetDelivered
func (mr *MockDBMockRecorder) SetDelivered(owner, id interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetDelivered", reflect.TypeOf((*MockDB)(nil).SetDelivered), owner, id)
}

// Prices mocks base method
func (m *MockDB) Prices(currency string, startTime, endTime int64) ([]db.Price, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Prices", currency, startTime, endTime)
	ret0, _ := ret[0].([]db.Price)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Prices indicates an expected call of Prices
func (mr *MockDBMockRecorder) Prices(currency, startTime, endTime interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Prices", reflect.TypeOf((*MockDB)(nil).Prices), currency, startTime, endTime)
}

// LastPriceByCurrency mocks base method
func (m *MockDB) LastPriceByCurrency(currency string) (*db.Price, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "LastPriceByCurrency", currency)
	ret0, _ := ret[0].(*db.Price)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// LastPriceByCurrency indicates an expected call of LastPriceByCurrency
func (mr *MockDBMockRecorder) LastPriceByCurrency(currency interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "LastPriceByCurrency", reflect.TypeOf((*MockDB)(nil).LastPriceByCurrency), currency)
}

// SearchUsersByUsername mocks base method
func (m *MockDB) SearchUsersByUsername(value string, limit int64) ([]db.Profile, error) {
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
func (m *MockDB) SearchUsersByEmail(email string) (*db.Profile, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SearchUsersByEmail", email)
	ret0, _ := ret[0].(*db.Profile)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// SearchUsersByEmail indicates an expected call of SearchUsersByEmail
func (mr *MockDBMockRecorder) SearchUsersByEmail(email interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SearchUsersByEmail", reflect.TypeOf((*MockDB)(nil).SearchUsersByEmail), email)
}

// ProfileById mocks base method
func (m *MockDB) ProfileById(id db.ID) (*db.Profile, error) {
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

// ProfileByAuthId mocks base method
func (m *MockDB) ProfileByAuthId(authId string) (*db.Profile, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ProfileByAuthId", authId)
	ret0, _ := ret[0].(*db.Profile)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ProfileByAuthId indicates an expected call of ProfileByAuthId
func (mr *MockDBMockRecorder) ProfileByAuthId(authId interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ProfileByAuthId", reflect.TypeOf((*MockDB)(nil).ProfileByAuthId), authId)
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

// ProfileByAddress mocks base method
func (m *MockDB) ProfileByAddress(network types.Network, address string) (*db.Profile, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ProfileByAddress", network, address)
	ret0, _ := ret[0].(*db.Profile)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ProfileByAddress indicates an expected call of ProfileByAddress
func (mr *MockDBMockRecorder) ProfileByAddress(network, address interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ProfileByAddress", reflect.TypeOf((*MockDB)(nil).ProfileByAddress), network, address)
}

// ProfileByPhoneNumber mocks base method
func (m *MockDB) ProfileByPhoneNumber(phoneNumber string) (*db.Profile, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ProfileByPhoneNumber", phoneNumber)
	ret0, _ := ret[0].(*db.Profile)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ProfileByPhoneNumber indicates an expected call of ProfileByPhoneNumber
func (mr *MockDBMockRecorder) ProfileByPhoneNumber(phoneNumber interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ProfileByPhoneNumber", reflect.TypeOf((*MockDB)(nil).ProfileByPhoneNumber), phoneNumber)
}

// ProfileByEmail mocks base method
func (m *MockDB) ProfileByEmail(email string) (*db.Profile, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ProfileByEmail", email)
	ret0, _ := ret[0].(*db.Profile)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ProfileByEmail indicates an expected call of ProfileByEmail
func (mr *MockDBMockRecorder) ProfileByEmail(email interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ProfileByEmail", reflect.TypeOf((*MockDB)(nil).ProfileByEmail), email)
}

// IsUsernameExist mocks base method
func (m *MockDB) IsUsernameExist(username string) (bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "IsUsernameExist", username)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// IsUsernameExist indicates an expected call of IsUsernameExist
func (mr *MockDBMockRecorder) IsUsernameExist(username interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "IsUsernameExist", reflect.TypeOf((*MockDB)(nil).IsUsernameExist), username)
}

// ProfilesCount mocks base method
func (m *MockDB) ProfilesCount() (int64, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ProfilesCount")
	ret0, _ := ret[0].(int64)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ProfilesCount indicates an expected call of ProfilesCount
func (mr *MockDBMockRecorder) ProfilesCount() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ProfilesCount", reflect.TypeOf((*MockDB)(nil).ProfilesCount))
}

// SubscribersCountByToken mocks base method
func (m *MockDB) SubscribersCountByToken(token string) (int64, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SubscribersCountByToken", token)
	ret0, _ := ret[0].(int64)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// SubscribersCountByToken indicates an expected call of SubscribersCountByToken
func (mr *MockDBMockRecorder) SubscribersCountByToken(token interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SubscribersCountByToken", reflect.TypeOf((*MockDB)(nil).SubscribersCountByToken), token)
}

// SubscriberByProfileId mocks base method
func (m *MockDB) SubscriberByProfileId(id db.ID) (*db.Subscriber, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SubscriberByProfileId", id)
	ret0, _ := ret[0].(*db.Subscriber)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// SubscriberByProfileId indicates an expected call of SubscriberByProfileId
func (mr *MockDBMockRecorder) SubscriberByProfileId(id interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SubscriberByProfileId", reflect.TypeOf((*MockDB)(nil).SubscriberByProfileId), id)
}

// TokenByValue mocks base method
func (m *MockDB) TokenByValue(token string) (*db.Token, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "TokenByValue", token)
	ret0, _ := ret[0].(*db.Token)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// TokenByValue indicates an expected call of TokenByValue
func (mr *MockDBMockRecorder) TokenByValue(token interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "TokenByValue", reflect.TypeOf((*MockDB)(nil).TokenByValue), token)
}

// TokenByProfileId mocks base method
func (m *MockDB) TokenByProfileId(id db.ID) (*db.Token, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "TokenByProfileId", id)
	ret0, _ := ret[0].(*db.Token)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// TokenByProfileId indicates an expected call of TokenByProfileId
func (mr *MockDBMockRecorder) TokenByProfileId(id interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "TokenByProfileId", reflect.TypeOf((*MockDB)(nil).TokenByProfileId), id)
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

// InsertMany mocks base method
func (m *MockDB) InsertMany(values []interface{}) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "InsertMany", values)
	ret0, _ := ret[0].(error)
	return ret0
}

// InsertMany indicates an expected call of InsertMany
func (mr *MockDBMockRecorder) InsertMany(values interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "InsertMany", reflect.TypeOf((*MockDB)(nil).InsertMany), values)
}

// UpdateByPK mocks base method
func (m *MockDB) UpdateByPK(Id db.ID, value interface{}) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateByPK", Id, value)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdateByPK indicates an expected call of UpdateByPK
func (mr *MockDBMockRecorder) UpdateByPK(Id, value interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateByPK", reflect.TypeOf((*MockDB)(nil).UpdateByPK), Id, value)
}