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

package mocks

import (
	a "github.com/stretchr/testify/assert"
)

type SMTPResponse struct {
	t      TestingT
	err    error
	result string
}

func (sr *SMTPResponse) AssertOK() *SMTPResponse {
	a.NoError(sr.t, sr.err)
	return sr
}

func (sr *SMTPResponse) AssertError(wantErrMsg string) *SMTPResponse {
	if sr.err == nil {
		a.Fail(sr.t, "Error is nil", "Expected to have %q", wantErrMsg)
	} else {
		a.Regexp(sr.t, wantErrMsg, sr.err.Error(), "Expected error %s but got %s", wantErrMsg, sr.err)
	}
	return sr
}
