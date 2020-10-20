// Copyright (c) 2020 Proton Technologies AG
//
// This file is part of ProtonMail Bridge.
//
// ProtonMail Bridge is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// ProtonMail Bridge is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with ProtonMail Bridge.  If not, see <https://www.gnu.org/licenses/>.

package transfer

import (
	"testing"
	"time"

	"github.com/pkg/errors"
	a "github.com/stretchr/testify/assert"
	r "github.com/stretchr/testify/require"
)

func TestProgressUpdateCount(t *testing.T) {
	progress := newProgress(log, nil)
	drainProgressUpdateChannel(&progress)

	progress.updateCount("inbox", 10)
	progress.updateCount("archive", 20)
	progress.updateCount("inbox", 12)
	progress.updateCount("sent", 5)
	progress.updateCount("foo", 4)
	progress.updateCount("foo", 5)

	progress.finish()

	_, _, _, _, total := progress.GetCounts() //nolint[dogsled]
	r.Equal(t, uint(42), total)
}

func TestProgressAddingMessages(t *testing.T) {
	progress := newProgress(log, nil)
	drainProgressUpdateChannel(&progress)

	// msg1 has no problem.
	progress.addMessage("msg1", nil)
	progress.messageExported("msg1", []byte(""), nil)
	progress.messageImported("msg1", "", nil)

	// msg2 has an import problem.
	progress.addMessage("msg2", nil)
	progress.messageExported("msg2", []byte(""), nil)
	progress.messageImported("msg2", "", errors.New("failed import"))

	// msg3 has an export problem.
	progress.addMessage("msg3", nil)
	progress.messageExported("msg3", []byte(""), errors.New("failed export"))

	// msg4 has an export problem and import is also called.
	progress.addMessage("msg4", nil)
	progress.messageExported("msg4", []byte(""), errors.New("failed export"))
	progress.messageImported("msg4", "", nil)

	progress.finish()

	failed, imported, exported, added, _ := progress.GetCounts()
	a.Equal(t, uint(4), added)
	a.Equal(t, uint(2), exported)
	a.Equal(t, uint(2), imported)
	a.Equal(t, uint(3), failed)

	errorsMap := map[string]string{}
	for _, status := range progress.GetFailedMessages() {
		errorsMap[status.SourceID] = status.GetErrorMessage()
	}
	a.Equal(t, map[string]string{
		"msg2": "failed to import: failed import",
		"msg3": "failed to export: failed export",
		"msg4": "failed to export: failed export",
	}, errorsMap)
}

func TestProgressFinish(t *testing.T) {
	progress := newProgress(log, nil)
	drainProgressUpdateChannel(&progress)

	progress.finish()
	r.Nil(t, progress.updateCh)

	r.NotPanics(t, func() { progress.addMessage("msg", nil) })
}

func TestProgressFatalError(t *testing.T) {
	progress := newProgress(log, nil)
	drainProgressUpdateChannel(&progress)

	progress.fatal(errors.New("fatal error"))
	r.Nil(t, progress.updateCh)

	r.NotPanics(t, func() { progress.addMessage("msg", nil) })
}

func TestFailUnpauseAndStops(t *testing.T) {
	progress := newProgress(log, nil)
	drainProgressUpdateChannel(&progress)

	progress.Pause("pausing")
	progress.fatal(errors.New("fatal error"))

	r.Nil(t, progress.updateCh)
	r.True(t, progress.isStopped)
	r.False(t, progress.IsPaused())
	r.Eventually(t, progress.shouldStop, 2*time.Millisecond, time.Millisecond)
}

func TestStopClosesUpdates(t *testing.T) {
	progress := newProgress(log, nil)
	ch := progress.updateCh

	progress.Stop()
	r.Nil(t, progress.updateCh)
	r.PanicsWithError(t, "send on closed channel", func() { ch <- struct{}{} })
}

func drainProgressUpdateChannel(progress *Progress) {
	// updateCh is not needed to drain under tests - timeout is implemented.
	// But timeout takes time which would slow down tests.
	go func() {
		for range progress.updateCh {
		}
	}()
}
