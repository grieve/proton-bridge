// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/ProtonMail/proton-bridge/v3/internal/telemetry (interfaces: HeartbeatManager)

// Package mocks is a generated GoMock package.
package mocks

import (
	reflect "reflect"
	time "time"

	telemetry "github.com/ProtonMail/proton-bridge/v3/internal/telemetry"
	gomock "github.com/golang/mock/gomock"
)

// MockHeartbeatManager is a mock of HeartbeatManager interface.
type MockHeartbeatManager struct {
	ctrl     *gomock.Controller
	recorder *MockHeartbeatManagerMockRecorder
}

// MockHeartbeatManagerMockRecorder is the mock recorder for MockHeartbeatManager.
type MockHeartbeatManagerMockRecorder struct {
	mock *MockHeartbeatManager
}

// NewMockHeartbeatManager creates a new mock instance.
func NewMockHeartbeatManager(ctrl *gomock.Controller) *MockHeartbeatManager {
	mock := &MockHeartbeatManager{ctrl: ctrl}
	mock.recorder = &MockHeartbeatManagerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockHeartbeatManager) EXPECT() *MockHeartbeatManagerMockRecorder {
	return m.recorder
}

// GetLastHeartbeatSent mocks base method.
func (m *MockHeartbeatManager) GetLastHeartbeatSent() time.Time {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetLastHeartbeatSent")
	ret0, _ := ret[0].(time.Time)
	return ret0
}

// GetLastHeartbeatSent indicates an expected call of GetLastHeartbeatSent.
func (mr *MockHeartbeatManagerMockRecorder) GetLastHeartbeatSent() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetLastHeartbeatSent", reflect.TypeOf((*MockHeartbeatManager)(nil).GetLastHeartbeatSent))
}

// IsTelemetryAvailable mocks base method.
func (m *MockHeartbeatManager) IsTelemetryAvailable() bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "IsTelemetryAvailable")
	ret0, _ := ret[0].(bool)
	return ret0
}

// IsTelemetryAvailable indicates an expected call of IsTelemetryAvailable.
func (mr *MockHeartbeatManagerMockRecorder) IsTelemetryAvailable() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "IsTelemetryAvailable", reflect.TypeOf((*MockHeartbeatManager)(nil).IsTelemetryAvailable))
}

// SendHeartbeat mocks base method.
func (m *MockHeartbeatManager) SendHeartbeat(arg0 *telemetry.HeartbeatData) bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SendHeartbeat", arg0)
	ret0, _ := ret[0].(bool)
	return ret0
}

// SendHeartbeat indicates an expected call of SendHeartbeat.
func (mr *MockHeartbeatManagerMockRecorder) SendHeartbeat(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SendHeartbeat", reflect.TypeOf((*MockHeartbeatManager)(nil).SendHeartbeat), arg0)
}

// SetLastHeartbeatSent mocks base method.
func (m *MockHeartbeatManager) SetLastHeartbeatSent(arg0 time.Time) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SetLastHeartbeatSent", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// SetLastHeartbeatSent indicates an expected call of SetLastHeartbeatSent.
func (mr *MockHeartbeatManagerMockRecorder) SetLastHeartbeatSent(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetLastHeartbeatSent", reflect.TypeOf((*MockHeartbeatManager)(nil).SetLastHeartbeatSent), arg0)
}
