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

package useragent

import (
	"fmt"
	"regexp"
	"runtime"
)

type UserAgent struct {
	client, platform string
}

func New() *UserAgent {
	return &UserAgent{
		client:   "",
		platform: runtime.GOOS,
	}
}

func (ua *UserAgent) SetClient(name, version string) {
	ua.client = fmt.Sprintf("%v/%v", name, regexp.MustCompile(`(.*) \((.*)\)`).ReplaceAllString(version, "$1-$2"))
}

func (ua *UserAgent) HasClient() bool {
	return ua.client != ""
}

func (ua *UserAgent) SetPlatform(platform string) {
	ua.platform = platform
}

func (ua *UserAgent) String() string {
	var client string

	if ua.client != "" {
		client = ua.client
	} else {
		client = "NoClient/0.0.1"
	}

	return fmt.Sprintf("%v (%v)", client, ua.platform)
}
