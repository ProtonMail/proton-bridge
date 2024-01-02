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

package tests

import (
	"sync"
	"testing"

	"github.com/ProtonMail/gluon/reporter"
	"github.com/bradenaw/juniper/xslices"
	"github.com/golang/mock/gomock"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

type reportRecord struct {
	isException bool
	message     string
	context     reporter.Context
}

type reportRecorder struct {
	assert  *assert.Assertions
	reports []reportRecord

	lock       sync.Locker
	isClosed   bool
	skipAssert bool
}

func newReportRecorder(tb testing.TB) *reportRecorder {
	return &reportRecorder{
		assert:   assert.New(tb),
		reports:  []reportRecord{},
		lock:     &sync.Mutex{},
		isClosed: false,
	}
}

func (r *reportRecorder) skipAsserts() {
	r.skipAssert = true
}

func (r *reportRecorder) add(isException bool, message string, context reporter.Context) {
	r.lock.Lock()
	defer r.lock.Unlock()

	l := logrus.WithFields(logrus.Fields{
		"isException": isException,
		"message":     message,
		"context":     context,
		"pkg":         "test/reportRecorder",
	})

	if r.isClosed {
		l.Warn("Reporter closed, report skipped")
		return
	}

	r.reports = append(r.reports, reportRecord{
		isException: isException,
		message:     message,
		context:     context,
	})

	l.Warn("Report recorded")
}

func (r *reportRecorder) close() {
	r.lock.Lock()
	defer r.lock.Unlock()

	r.isClosed = true
}

func (r *reportRecorder) assertEmpty() {
	if !r.skipAssert {
		r.assert.Empty(r.reports)
	}
}

func (r *reportRecorder) removeMatchingRecords(isException, message, context gomock.Matcher, n int) {
	if n == 0 {
		n = len(r.reports)
	}

	r.reports = xslices.Filter(r.reports, func(rec reportRecord) bool {
		if n <= 0 {
			return true
		}

		l := logrus.WithFields(logrus.Fields{
			"rec": rec,
		})
		if !isException.Matches(rec.isException) {
			l.WithField("matcher", isException).Debug("Not matching")
			return true
		}

		if !message.Matches(rec.message) {
			l.WithField("matcher", message).Debug("Not matching")
			return true
		}

		if !context.Matches(rec.context) {
			l.WithField("matcher", context).Debug("Not matching")
			return true
		}

		n--

		return false
	})
}

func (r *reportRecorder) ReportException(data any) error {
	r.add(true, "exception", reporter.Context{"data": data})
	return nil
}

func (r *reportRecorder) ReportMessage(message string) error {
	r.add(false, message, reporter.Context{})
	return nil
}

func (r *reportRecorder) ReportMessageWithContext(message string, context reporter.Context) error {
	r.add(false, message, context)
	return nil
}

func (r *reportRecorder) ReportExceptionWithContext(data any, context reporter.Context) error {
	if context == nil {
		context = reporter.Context{}
	}

	context["data"] = data

	r.add(true, "exception", context)

	return nil
}

func (s *scenario) skipReporterChecks() error {
	s.t.reporter.skipAsserts()
	return nil
}
