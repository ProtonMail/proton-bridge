// Copyright (c) 2020 Proton Technologies AG
//
// This file is part of ProtonMail Bridge.Bridge.
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

package liveapi

import (
	"fmt"

	"github.com/nsf/jsondiff"
)

type fakeCall struct {
	method  string
	path    string
	request []byte
}

func (cntrl *Controller) recordCall(method, path string, request []byte) {
	cntrl.lock.Lock()
	defer cntrl.lock.Unlock()

	cntrl.calls = append(cntrl.calls, &fakeCall{
		method:  method,
		path:    path,
		request: request,
	})
}

func (cntrl *Controller) PrintCalls() {
	fmt.Println("API calls:")
	for idx, call := range cntrl.calls {
		fmt.Printf("%02d: [%s] %s\n", idx+1, call.method, call.path)
		if call.request != nil && string(call.request) != "null" {
			fmt.Printf("\t%s\n", call.request)
		}
	}
}

func (cntrl *Controller) WasCalled(method, path string, expectedRequest []byte) bool {
	for _, call := range cntrl.calls {
		if call.method != method && call.path != path {
			continue
		}
		diff, _ := jsondiff.Compare(call.request, expectedRequest, &jsondiff.Options{})
		isSuperset := diff == jsondiff.FullMatch || diff == jsondiff.SupersetMatch
		if isSuperset {
			return true
		}
	}
	return false
}

func (cntrl *Controller) GetCalls(method, path string) [][]byte {
	requests := [][]byte{}
	for _, call := range cntrl.calls {
		if call.method == method && call.path == path {
			requests = append(requests, call.request)
		}
	}
	return requests
}
