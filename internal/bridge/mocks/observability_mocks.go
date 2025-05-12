package mocks

import (
	reflect "reflect"

	"github.com/ProtonMail/go-proton-api"
	"github.com/ProtonMail/proton-bridge/v3/internal/services/observability"
	"github.com/golang/mock/gomock"
)

type MockObservabilitySender struct {
	ctrl     *gomock.Controller
	recorder *MockObservabilitySenderRecorder
}

type MockObservabilitySenderRecorder struct {
	mock *MockObservabilitySender
}

func NewMockObservabilitySender(ctrl *gomock.Controller) *MockObservabilitySender {
	mock := &MockObservabilitySender{ctrl: ctrl}
	mock.recorder = &MockObservabilitySenderRecorder{mock: mock}
	return mock
}

func (m *MockObservabilitySender) EXPECT() *MockObservabilitySenderRecorder { return m.recorder }

func (m *MockObservabilitySender) AddDistinctMetrics(errType observability.DistinctionMetricTypeEnum, _ ...proton.ObservabilityMetric) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "AddDistinctMetrics", errType)
}

func (m *MockObservabilitySender) AddMetrics(metrics ...proton.ObservabilityMetric) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "AddMetrics", metrics)
}

func (m *MockObservabilitySender) AddTimeLimitedMetric(metricType observability.DistinctionMetricTypeEnum, metric proton.ObservabilityMetric) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "AddTimeLimitedMetric", metricType, metric)
}

func (m *MockObservabilitySender) GetEmailClient() string {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "GetEmailClient")
	return ""
}

func (mr *MockObservabilitySenderRecorder) AddDistinctMetrics(errType observability.DistinctionMetricTypeEnum, _ ...proton.ObservabilityMetric) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock,
		"AddDistinctMetrics",
		reflect.TypeOf((*MockObservabilitySender)(nil).AddDistinctMetrics),
		errType)
}

func (mr *MockObservabilitySenderRecorder) AddMetrics(metrics ...proton.ObservabilityMetric) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddMetrics", reflect.TypeOf((*MockObservabilitySender)(nil).AddMetrics), metrics)
}

func (mr *MockObservabilitySenderRecorder) AddTimeLimitedMetric(metricType observability.DistinctionMetricTypeEnum, metric proton.ObservabilityMetric) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddTimeLimitedMetric", reflect.TypeOf((*MockObservabilitySender)(nil).AddTimeLimitedMetric), metricType, metric)
}

func (mr *MockObservabilitySenderRecorder) GetEmailClient() {
	mr.mock.ctrl.T.Helper()
	mr.mock.ctrl.Call(mr.mock, "GetEmailClient", reflect.TypeOf((*MockObservabilitySender)(nil).GetEmailClient))
}
