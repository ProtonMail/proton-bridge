// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/ProtonMail/proton-bridge/v3/internal/events (interfaces: EventPublisher)

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	reflect "reflect"

	events "github.com/ProtonMail/proton-bridge/v3/internal/events"
	gomock "github.com/golang/mock/gomock"
)

// MockEventPublisher is a mock of EventPublisher interface.
type MockEventPublisher struct {
	ctrl     *gomock.Controller
	recorder *MockEventPublisherMockRecorder
}

// MockEventPublisherMockRecorder is the mock recorder for MockEventPublisher.
type MockEventPublisherMockRecorder struct {
	mock *MockEventPublisher
}

// NewMockEventPublisher creates a new mock instance.
func NewMockEventPublisher(ctrl *gomock.Controller) *MockEventPublisher {
	mock := &MockEventPublisher{ctrl: ctrl}
	mock.recorder = &MockEventPublisherMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockEventPublisher) EXPECT() *MockEventPublisherMockRecorder {
	return m.recorder
}

// PublishEvent mocks base method.
func (m *MockEventPublisher) PublishEvent(arg0 context.Context, arg1 events.Event) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "PublishEvent", arg0, arg1)
}

// PublishEvent indicates an expected call of PublishEvent.
func (mr *MockEventPublisherMockRecorder) PublishEvent(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "PublishEvent", reflect.TypeOf((*MockEventPublisher)(nil).PublishEvent), arg0, arg1)
}
