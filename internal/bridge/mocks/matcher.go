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

package mocks

import (
	"strings"

	"github.com/ProtonMail/go-proton-api"
)

type refreshContextMatcher struct {
	wantRefresh proton.RefreshFlag
}

func NewRefreshContextMatcher(refreshFlag proton.RefreshFlag) *refreshContextMatcher { //nolint:revive
	return &refreshContextMatcher{wantRefresh: refreshFlag}
}

func (m *refreshContextMatcher) Matches(x interface{}) bool {
	context, ok := x.(map[string]interface{})
	if !ok {
		return false
	}

	i, ok := context["EventLoop"]
	if !ok {
		return false
	}

	el, ok := i.(map[string]interface{})
	if !ok {
		return false
	}

	vID, ok := el["EventID"]
	if !ok {
		return false
	}

	id, ok := vID.(string)
	if !ok {
		return false
	}

	if id == "" {
		return false
	}

	vRefresh, ok := el["Refresh"]
	if !ok {
		return false
	}

	refresh, ok := vRefresh.(proton.RefreshFlag)
	if !ok {
		return false
	}

	return refresh == m.wantRefresh
}

func (m *refreshContextMatcher) String() string {
	return `map[string]interface which contains "Refresh" field with value proton.RefreshAll`
}

type closedConnectionMatcher struct{}

func NewClosedConnectionMatcher() *closedConnectionMatcher { //nolint:revive
	return &closedConnectionMatcher{}
}

func (m *closedConnectionMatcher) Matches(x interface{}) bool {
	context, ok := x.(map[string]interface{})
	if !ok {
		return false
	}

	vErr, ok := context["error"]
	if !ok {
		return false
	}

	err, ok := vErr.(error)
	if !ok {
		return false
	}

	return strings.Contains(err.Error(), "used of closed network connection")
}

func (m *closedConnectionMatcher) String() string {
	return "map containing error of closed network connection"
}
