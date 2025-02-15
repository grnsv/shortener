// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/grnsv/shortener/internal/service (interfaces: Shortener)

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	models "github.com/grnsv/shortener/internal/models"
)

// MockShortener is a mock of Shortener interface.
type MockShortener struct {
	ctrl     *gomock.Controller
	recorder *MockShortenerMockRecorder
}

// MockShortenerMockRecorder is the mock recorder for MockShortener.
type MockShortenerMockRecorder struct {
	mock *MockShortener
}

// NewMockShortener creates a new mock instance.
func NewMockShortener(ctrl *gomock.Controller) *MockShortener {
	mock := &MockShortener{ctrl: ctrl}
	mock.recorder = &MockShortenerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockShortener) EXPECT() *MockShortenerMockRecorder {
	return m.recorder
}

// ExpandURL mocks base method.
func (m *MockShortener) ExpandURL(arg0 context.Context, arg1 string) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ExpandURL", arg0, arg1)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ExpandURL indicates an expected call of ExpandURL.
func (mr *MockShortenerMockRecorder) ExpandURL(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ExpandURL", reflect.TypeOf((*MockShortener)(nil).ExpandURL), arg0, arg1)
}

// GetAll mocks base method.
func (m *MockShortener) GetAll(arg0 context.Context, arg1 string) ([]models.URL, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetAll", arg0, arg1)
	ret0, _ := ret[0].([]models.URL)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetAll indicates an expected call of GetAll.
func (mr *MockShortenerMockRecorder) GetAll(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAll", reflect.TypeOf((*MockShortener)(nil).GetAll), arg0, arg1)
}

// PingStorage mocks base method.
func (m *MockShortener) PingStorage(arg0 context.Context) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "PingStorage", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// PingStorage indicates an expected call of PingStorage.
func (mr *MockShortenerMockRecorder) PingStorage(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "PingStorage", reflect.TypeOf((*MockShortener)(nil).PingStorage), arg0)
}

// ShortenBatch mocks base method.
func (m *MockShortener) ShortenBatch(arg0 context.Context, arg1 models.BatchRequest, arg2 string) (models.BatchResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ShortenBatch", arg0, arg1, arg2)
	ret0, _ := ret[0].(models.BatchResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ShortenBatch indicates an expected call of ShortenBatch.
func (mr *MockShortenerMockRecorder) ShortenBatch(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ShortenBatch", reflect.TypeOf((*MockShortener)(nil).ShortenBatch), arg0, arg1, arg2)
}

// ShortenURL mocks base method.
func (m *MockShortener) ShortenURL(arg0 context.Context, arg1, arg2 string) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ShortenURL", arg0, arg1, arg2)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ShortenURL indicates an expected call of ShortenURL.
func (mr *MockShortenerMockRecorder) ShortenURL(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ShortenURL", reflect.TypeOf((*MockShortener)(nil).ShortenURL), arg0, arg1, arg2)
}
