// Copyright (c) 2021 Proton Technologies AG
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

package fakeapi

import (
	"encoding/json"
	"fmt"

	"github.com/ProtonMail/proton-bridge/pkg/pmapi"
	"github.com/nsf/jsondiff"
)

type method string

const (
	GET    method = "GET"
	POST   method = "POST"
	PUT    method = "PUT"
	DELETE method = "DELETE"
)

type fakeCall struct {
	method  method
	path    string
	request []byte
}

func (ctl *Controller) checkAndRecordCall(method method, path string, req interface{}) error {
	ctl.lock.Lock()
	defer ctl.lock.Unlock()

	var request []byte

	if req != nil {
		var err error

		if request, err = json.Marshal(req); err != nil {
			panic(err)
		}
	}

	ctl.calls = append(ctl.calls, &fakeCall{
		method:  method,
		path:    path,
		request: request,
	})

	if ctl.noInternetConnection {
		return pmapi.ErrNoConnection
	}

	return nil
}

func (ctl *Controller) PrintCalls() {
	fmt.Println("API calls:")
	for idx, call := range ctl.calls {
		fmt.Printf("%02d: [%s] %s\n", idx+1, call.method, call.path)
		if call.request != nil && string(call.request) != "null" {
			fmt.Printf("\t%s\n", call.request)
		}
	}
}

func (ctl *Controller) WasCalled(method, path string, expectedRequest []byte) bool {
	for _, call := range ctl.calls {
		if string(call.method) != method || call.path != path {
			continue
		}
		if string(expectedRequest) == "" {
			return true
		}
		diff, _ := jsondiff.Compare(call.request, expectedRequest, &jsondiff.Options{})
		isSuperset := diff == jsondiff.FullMatch || diff == jsondiff.SupersetMatch
		if isSuperset {
			return true
		}
	}
	return false
}

func (ctl *Controller) GetCalls(method, path string) [][]byte {
	requests := [][]byte{}
	for _, call := range ctl.calls {
		if string(call.method) == method && call.path == path {
			requests = append(requests, call.request)
		}
	}
	return requests
}
