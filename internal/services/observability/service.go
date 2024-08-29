// Copyright (c) 2024 Proton AG
//
// This file is part of Proton Mail Bridge.
//
// Proton Mail Bridge is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Proton Mail Bridge is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with Proton Mail Bridge.  If not, see <https://www.gnu.org/licenses/>.

package observability

import (
	"context"
	"sync"
	"time"

	"github.com/ProtonMail/gluon/async"
	"github.com/ProtonMail/go-proton-api"
	"github.com/ProtonMail/proton-bridge/v3/internal/services/telemetry"
	"github.com/sirupsen/logrus"
)

// Non-const for testing.
var throttleDuration = 5 * time.Second //nolint:gochecknoglobals

const (
	maxStorageSize = 5000
	maxBatchSize   = 1000
)

type PushObsMetricFn func(metric proton.ObservabilityMetric)

type client struct {
	isTelemetryEnabled func(context.Context) bool
	sendMetrics        func(context.Context, proton.ObservabilityBatch) error
}

type Service struct {
	ctx    context.Context
	cancel context.CancelFunc

	panicHandler async.PanicHandler

	lastDispatch        time.Time
	isDispatchScheduled bool

	signalDataArrived chan struct{}
	signalDispatch    chan struct{}

	log *logrus.Entry

	metricStore     []proton.ObservabilityMetric
	metricStoreLock sync.Mutex

	userClientStore     map[string]*client
	userClientStoreLock sync.Mutex
}

func NewService(ctx context.Context, panicHandler async.PanicHandler) *Service {
	ctx, cancel := context.WithCancel(ctx)

	service := &Service{
		ctx:    ctx,
		cancel: cancel,

		panicHandler: panicHandler,

		lastDispatch: time.Now().Add(-throttleDuration),

		signalDataArrived: make(chan struct{}, 1),
		signalDispatch:    make(chan struct{}, 1),

		log: logrus.WithFields(logrus.Fields{"pkg": "observability"}),

		metricStore: make([]proton.ObservabilityMetric, 0),

		userClientStore: make(map[string]*client),
	}

	return service
}

func (s *Service) Run() {
	s.log.Info("Starting service")
	go func() {
		s.start()
	}()
}

// When new data is received, we determine if we can immediately send the request.
// First, we check if a dispatch operation is already scheduled. If it is, we do nothing.
// If no dispatch is scheduled, we verify if the required time interval has passed since the last send.
// If the interval hasn't passed, we schedule the dispatch to occur when the threshold is met.
// If the interval has passed, we initiate an immediate dispatch.
func (s *Service) start() {
	defer async.HandlePanic(s.panicHandler)
	for {
		select {
		case <-s.ctx.Done():
			return

		case <-s.signalDispatch:
			s.dispatchData()

		case <-s.signalDataArrived:
			if s.isDispatchScheduled {
				continue
			}

			if time.Since(s.lastDispatch) <= throttleDuration {
				s.scheduleDispatch()
				continue
			}

			s.sendSignal(s.signalDispatch)
		}
	}
}

func (s *Service) dispatchData() {
	s.isDispatchScheduled = false // Only accessed via a single goroutine, so no mutexes.
	if !s.haveMetricsAndClients() {
		return
	}

	// Get a copy of the metrics we want to send and batch them accordingly
	var numberOfRemainingMetrics int
	var metricsToSend []proton.ObservabilityMetric

	s.withMetricStoreLock(func() {
		numberOfMetricsToSend := min(len(s.metricStore), maxBatchSize)
		metricsToSend = make([]proton.ObservabilityMetric, numberOfMetricsToSend)
		copy(metricsToSend, s.metricStore[:numberOfMetricsToSend])
		s.metricStore = s.metricStore[numberOfMetricsToSend:]
		numberOfRemainingMetrics = len(s.metricStore)
	})

	// Send them out to the endpoint
	telemetryEnabled := s.dispatchViaClient(&metricsToSend)

	// If there are more metric updates than the max batch limit and telemetry is enabled for one of the clients
	// then we immediately schedule another dispatch.
	if numberOfRemainingMetrics > 0 && telemetryEnabled {
		s.scheduleDispatch()
	}
}

// dispatchViaClient - return value tells us whether telemetry is enabled
// such that we know whether to schedule another dispatch if more data is present.
func (s *Service) dispatchViaClient(metricsToSend *[]proton.ObservabilityMetric) bool {
	s.log.Info("Sending observability data.")
	s.userClientStoreLock.Lock()
	defer s.userClientStoreLock.Unlock()

	for _, value := range s.userClientStore {
		if !value.isTelemetryEnabled(s.ctx) {
			continue
		}

		if err := value.sendMetrics(s.ctx, proton.ObservabilityBatch{Metrics: *metricsToSend}); err != nil {
			s.log.WithError(err).Error("Issue occurred when sending observability data.")
		} else {
			s.log.Info("Successfully sent observability data.")
		}

		s.lastDispatch = time.Now()
		return true
	}

	s.log.Info("Could not send observability data. Telemetry is not enabled.")
	return false
}

func (s *Service) scheduleDispatch() {
	waitTime := throttleDuration - time.Since(s.lastDispatch)
	if waitTime <= 0 {
		s.sendSignal(s.signalDispatch)
		return
	}

	s.log.Info("Scheduling observability data sending")

	s.isDispatchScheduled = true
	go func() {
		defer async.HandlePanic(s.panicHandler)
		select {
		case <-s.ctx.Done():
			return
		case <-time.After(waitTime):
			s.sendSignal(s.signalDispatch)
		}
	}()
}

func (s *Service) AddMetric(metric proton.ObservabilityMetric) {
	s.withMetricStoreLock(func() {
		metricStoreLength := len(s.metricStore)
		if metricStoreLength >= maxStorageSize {
			s.log.Info("Max metric storage size has been exceeded. Dropping oldest metrics.")

			dropCount := metricStoreLength - maxStorageSize + 1
			s.metricStore = s.metricStore[dropCount:]
		}
		s.metricStore = append(s.metricStore, metric)
	})

	s.sendSignal(s.signalDataArrived)
}

func (s *Service) RegisterUserClient(userID string, protonClient *proton.Client, telemetryService *telemetry.Service) {
	s.log.Info("Registering user client, ID:", userID)

	s.withUserClientStoreLock(func() {
		s.userClientStore[userID] = &client{
			isTelemetryEnabled: telemetryService.IsTelemetryEnabled,
			sendMetrics:        protonClient.SendObservabilityBatch,
		}
	})

	// There may be a case where we already have metric updates stored, so try to flush;
	s.sendSignal(s.signalDataArrived)
}

func (s *Service) DeregisterUserClient(userID string) {
	s.log.Info("De-registering user client, ID:", userID)

	s.withUserClientStoreLock(func() {
		delete(s.userClientStore, userID)
	})
}

func (s *Service) Stop() {
	s.log.Info("Stopping service")

	s.cancel()
	close(s.signalDataArrived)
	close(s.signalDispatch)
}

// Utility functions below.
func (s *Service) haveMetricsAndClients() bool {
	s.metricStoreLock.Lock()
	s.userClientStoreLock.Lock()
	defer s.metricStoreLock.Unlock()
	defer s.userClientStoreLock.Unlock()

	return len(s.metricStore) > 0 && len(s.userClientStore) > 0
}

func (s *Service) withUserClientStoreLock(fn func()) {
	s.userClientStoreLock.Lock()
	defer s.userClientStoreLock.Unlock()
	fn()
}

func (s *Service) withMetricStoreLock(fn func()) {
	s.metricStoreLock.Lock()
	defer s.metricStoreLock.Unlock()
	fn()
}

// We use buffered channels; we shouldn't block them.
func (s *Service) sendSignal(channel chan struct{}) {
	select {
	case channel <- struct{}{}:
	default:
	}
}

// ModifyThrottlePeriod - used for testing.
func ModifyThrottlePeriod(duration time.Duration) {
	throttleDuration = duration
}
