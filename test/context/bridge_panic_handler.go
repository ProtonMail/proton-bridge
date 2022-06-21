// Copyright (c) 2022 Proton AG
//
// This file is part of Proton Mail Bridge.Bridge.
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

package context

import (
	"bytes"
	"io/ioutil"
	"runtime/pprof"
)

type panicHandler struct {
	t *bddT
}

func newPanicHandler(t *bddT) *panicHandler {
	return &panicHandler{
		t: t,
	}
}

// HandlePanic makes the panicHandler implement the panicHandler interface for bridge.
func (ph *panicHandler) HandlePanic() {
	r := recover()
	if r != nil {
		ph.t.Errorf("panic: %s", r)

		r := bytes.NewBufferString("")
		_ = pprof.Lookup("goroutine").WriteTo(r, 2)
		b, err := ioutil.ReadAll(r)
		ph.t.Errorf("pprof details: %s %s", err, b)

		ph.t.FailNow()
	}
}
