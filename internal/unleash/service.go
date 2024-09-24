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
// along with Proton Mail Bridge. If not, see <https://www.gnu.org/licenses/>.

package unleash

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/ProtonMail/gluon/async"
	"github.com/ProtonMail/go-proton-api"
	"github.com/ProtonMail/proton-bridge/v3/internal/service"
	"github.com/sirupsen/logrus"
)

var pollPeriod = 10 * time.Minute //nolint:gochecknoglobals
var pollJitter = 2 * time.Minute  //nolint:gochecknoglobals

const filename = "unleash_flags"

type requestFeaturesFn func(ctx context.Context) (proton.FeatureFlagResult, error)
type GetFlagValueFn func(key string) bool

type Service struct {
	panicHandler async.PanicHandler
	timer        *proton.Ticker

	ctx    context.Context
	cancel context.CancelFunc

	log *logrus.Entry

	ffStore     map[string]bool
	ffStoreLock sync.Mutex

	cacheFilepath string
	cacheFileLock sync.Mutex

	channel chan map[string]bool

	getFeaturesFn func(ctx context.Context) (proton.FeatureFlagResult, error)
}

func NewBridgeService(ctx context.Context, api *proton.Manager, locator service.Locator, panicHandler async.PanicHandler) *Service {
	log := logrus.WithField("service", "unleash")
	cacheDir, err := locator.ProvideUnleashCachePath()
	if err != nil {
		log.Warn("Could not find or create unleash cache directory")
	}
	cachePath := filepath.Clean(filepath.Join(cacheDir, filename))

	return newService(ctx, func(ctx context.Context) (proton.FeatureFlagResult, error) {
		return api.GetFeatures(ctx)
	}, log, cachePath, panicHandler)
}

func newService(ctx context.Context, fn requestFeaturesFn, log *logrus.Entry, cachePath string, panicHandler async.PanicHandler) *Service {
	ctx, cancel := context.WithCancel(ctx)

	unleashService := &Service{
		panicHandler: panicHandler,
		timer:        proton.NewTicker(pollPeriod, pollJitter, panicHandler),

		ctx:    ctx,
		cancel: cancel,

		log: log,

		ffStore:       make(map[string]bool),
		cacheFilepath: cachePath,

		channel: make(chan map[string]bool),

		getFeaturesFn: fn,
	}

	unleashService.readCacheFile()
	return unleashService
}

func readResponseData(data proton.FeatureFlagResult) map[string]bool {
	featureData := make(map[string]bool)
	for _, el := range data.Toggles {
		featureData[el.Name] = el.Enabled
	}

	return featureData
}

func (s *Service) readCacheFile() {
	defer s.cacheFileLock.Unlock()
	s.cacheFileLock.Lock()

	file, err := os.Open(s.cacheFilepath)
	if err != nil {
		s.log.WithError(err).Info("Unable to open cache file")
		return
	}

	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			s.log.WithError(err).Error("Unable to close cache file after read")
		}
	}(file)

	s.ffStoreLock.Lock()
	defer s.ffStoreLock.Unlock()
	if err = json.NewDecoder(file).Decode(&s.ffStore); err != nil {
		s.log.WithError(err).Error("Unable to decode cache file")
	}
}

func (s *Service) writeCacheFile() {
	defer s.cacheFileLock.Unlock()
	s.cacheFileLock.Lock()

	file, err := os.Create(s.cacheFilepath)
	if err != nil {
		s.log.WithError(err).Error("Unable to create cache file")
		return
	}

	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			s.log.WithError(err).Error("Unable to close cache file after write")
		}
	}(file)

	s.ffStoreLock.Lock()
	defer s.ffStoreLock.Unlock()
	if err = json.NewEncoder(file).Encode(s.ffStore); err != nil {
		s.log.WithError(err).Error("Unable to encode data to cache file")
	}
}

func (s *Service) Run() {
	s.log.Info("Starting service")

	go func() {
		s.runFlagPoll()
	}()

	go func() {
		s.runReceiver()
	}()
}

func (s *Service) runFlagPoll() {
	defer async.HandlePanic(s.panicHandler)
	defer s.timer.Stop()
	s.log.Info("Starting poll service")

	data, err := s.getFeaturesFn(s.ctx)
	if err != nil {
		s.log.WithError(err).Error("Failed to get flags from server")
	} else {
		s.channel <- readResponseData(data)
	}

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-s.timer.C:
			s.log.Info("Polling flag service")
			data, err := s.getFeaturesFn(s.ctx)
			if err != nil {
				s.log.WithError(err).Error("Failed to get feature flags from server")
				continue
			}
			s.channel <- readResponseData(data)
		}
	}
}

func (s *Service) runReceiver() {
	defer async.HandlePanic(s.panicHandler)
	s.log.Info("Starting receiver service")

	for {
		select {
		case <-s.ctx.Done():
			return
		case res := <-s.channel:
			s.ffStoreLock.Lock()
			s.ffStore = res
			s.ffStoreLock.Unlock()
			s.writeCacheFile()
		}
	}
}

func (s *Service) GetFlagValue(key string) bool {
	defer s.ffStoreLock.Unlock()
	s.ffStoreLock.Lock()

	val, ok := s.ffStore[key]
	if !ok {
		return false
	}

	return val
}

func (s *Service) Close() {
	s.log.Info("Closing service")
	s.cancel()
	close(s.channel)
}

// ModifyPollPeriodAndJitter is only used for testing.
func ModifyPollPeriodAndJitter(pollInterval, jitterInterval time.Duration) {
	pollPeriod = pollInterval
	pollJitter = jitterInterval
}
