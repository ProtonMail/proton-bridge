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

package sentry

import (
	"testing"

	r "github.com/stretchr/testify/require"

	"github.com/getsentry/sentry-go"
)

func TestFilterOutPanicHandlers(t *testing.T) {
	frames := []sentry.Frame{
		{Module: "github.com/ProtonMail/proton-bridge/internal/cmd", Function: "main"},
		{Module: "github.com/urfave/cli", Function: "(*App).Run"},
		{Module: "github.com/ProtonMail/proton-bridge/internal/cmd", Function: "RegisterHandlePanic"},
		{Module: "github.com/ProtonMail/pkg", Function: "HandlePanic"},
		{Module: "main", Function: "run"},
		{Module: "github.com/ProtonMail/proton-bridge/pkg/config", Function: "(*PanicHandler).HandlePanic"},
		{Module: "github.com/ProtonMail/proton-bridge/pkg/config", Function: "HandlePanic"},
		{Module: "github.com/ProtonMail/proton-bridge/pkg/sentry", Function: "ReportSentryCrash"},
		{Module: "github.com/ProtonMail/proton-bridge/pkg/sentry", Function: "ReportSentryCrash.func1"},
	}

	gotFrames := filterOutPanicHandlers(frames)
	r.Equal(t, frames[:5], gotFrames)
}
