// Copyright (c) 2022 Proton AG
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

package sentry

import (
	"testing"

	r "github.com/stretchr/testify/require"

	"github.com/getsentry/sentry-go"
)

func TestSkipDuringUnwind(t *testing.T) {
	// More calls in one function adds it only once.
	SkipDuringUnwind()
	SkipDuringUnwind()
	func() {
		SkipDuringUnwind()
		SkipDuringUnwind()
	}()

	wantSkippedFunctions := []string{
		"github.com/ProtonMail/proton-bridge/v2/internal/sentry.TestSkipDuringUnwind",
		"github.com/ProtonMail/proton-bridge/v2/internal/sentry.TestSkipDuringUnwind.func1",
	}
	r.Equal(t, wantSkippedFunctions, skippedFunctions)
}

func TestFilterOutPanicHandlers(t *testing.T) {
	skippedFunctions = []string{
		"github.com/ProtonMail/proton-bridge/v2/pkg/config.(*PanicHandler).HandlePanic",
		"github.com/ProtonMail/proton-bridge/v2/pkg/config.HandlePanic",
		"github.com/ProtonMail/proton-bridge/v2/internal/sentry.ReportSentryCrash",
		"github.com/ProtonMail/proton-bridge/v2/internal/sentry.ReportSentryCrash.func1",
	}

	frames := []sentry.Frame{
		{Module: "github.com/ProtonMail/proton-bridge/v2/internal/cmd", Function: "main"},
		{Module: "github.com/urfave/cli", Function: "(*App).Run"},
		{Module: "github.com/ProtonMail/proton-bridge/v2/internal/cmd", Function: "RegisterHandlePanic"},
		{Module: "github.com/ProtonMail/pkg", Function: "HandlePanic"},
		{Module: "main", Function: "run"},
		{Module: "github.com/ProtonMail/proton-bridge/v2/pkg/config", Function: "(*PanicHandler).HandlePanic"},
		{Module: "github.com/ProtonMail/proton-bridge/v2/pkg/config", Function: "HandlePanic"},
		{Module: "github.com/ProtonMail/proton-bridge/v2/internal/sentry", Function: "ReportSentryCrash"},
		{Module: "github.com/ProtonMail/proton-bridge/v2/internal/sentry", Function: "ReportSentryCrash.func1"},
	}

	gotFrames := filterOutPanicHandlers(frames)
	r.Equal(t, frames[:5], gotFrames)
}
