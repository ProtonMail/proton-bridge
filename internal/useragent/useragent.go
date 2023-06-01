// Copyright (c) 2023 Proton AG
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
	"sync"
)

const DefaultUserAgent = "NoClient/0.0.1"
const DefaultVersion = "0.0.1"
const UnknownClient = "UnknownClient"

type UserAgent struct {
	client, platform string

	lock sync.RWMutex
}

func New() *UserAgent {
	return &UserAgent{
		client:   "",
		platform: runtime.GOOS,
	}
}

func (ua *UserAgent) SetClient(name, version string) {
	ua.lock.Lock()
	defer ua.lock.Unlock()

	ua.client = fmt.Sprintf("%v/%v", name, regexp.MustCompile(`(.*) \((.*)\)`).ReplaceAllString(version, "$1-$2"))
}

func (ua *UserAgent) SetClientString(client string) {
	ua.lock.Lock()
	defer ua.lock.Unlock()

	ua.client = client
}

func (ua *UserAgent) GetClientString() string {
	ua.lock.RLock()
	defer ua.lock.RUnlock()

	return ua.client
}

func (ua *UserAgent) HasClient() bool {
	ua.lock.RLock()
	defer ua.lock.RUnlock()

	return ua.client != ""
}

func (ua *UserAgent) SetPlatform(platform string) {
	ua.lock.Lock()
	defer ua.lock.Unlock()

	ua.platform = platform
}

func (ua *UserAgent) GetUserAgent() string {
	ua.lock.RLock()
	defer ua.lock.RUnlock()

	var client string

	if ua.client != "" {
		client = ua.client
	} else {
		client = DefaultUserAgent
	}

	return fmt.Sprintf("%v (%v)", client, ua.platform)
}
