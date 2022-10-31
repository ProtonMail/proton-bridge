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

package calls

import (
	"fmt"
	"regexp"

	"github.com/nsf/jsondiff"
)

type CallRecord struct {
	method  string
	path    string
	request []byte
}

type Calls []CallRecord

func (c *Calls) Register(method, path string, request []byte) {
	*c = append(*c, CallRecord{
		method:  method,
		path:    path,
		request: request,
	})
}

func (c *Calls) PrintCalls() {
	fmt.Println("API calls:")
	for idx, call := range *c {
		fmt.Printf("%02d: [%s] %s\n", idx+1, call.method, call.path)
		if call.request != nil && string(call.request) != "null" {
			fmt.Printf("\t%s\n", call.request)
		}
	}
}

func (c *Calls) WasCalled(method, path string, expectedRequest []byte) bool {
	res, err := c.WasCalledRegex("^"+regexp.QuoteMeta(method)+"$", "^"+regexp.QuoteMeta(path)+"$", expectedRequest)
	if err != nil {
		panic(err)
	}
	return res
}

func (c *Calls) WasCalledRegex(methodRegex, pathRegex string, expectedRequest []byte) (bool, error) {
	for _, call := range *c {
		matched, err := regexp.Match(methodRegex, []byte(call.method))
		if err != nil {
			return false, err
		}
		if !matched {
			continue
		}

		matched, err = regexp.Match(pathRegex, []byte(call.path))
		if err != nil {
			return false, err
		}
		if !matched {
			continue
		}

		if string(expectedRequest) == "" {
			return true, nil
		}
		diff, _ := jsondiff.Compare(call.request, expectedRequest, &jsondiff.Options{})
		isSuperset := diff == jsondiff.FullMatch || diff == jsondiff.SupersetMatch
		if isSuperset {
			return true, nil
		}
	}
	return false, nil
}

func (c *Calls) GetCalls(method, path string) [][]byte {
	requests := [][]byte{}
	for _, call := range *c {
		if call.method == method && call.path == path {
			requests = append(requests, call.request)
		}
	}
	return requests
}
