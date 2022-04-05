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

package liveapi

func (ctl *Controller) recordCall(method, path string, request []byte) {
	ctl.lock.Lock()
	defer ctl.lock.Unlock()

	ctl.calls.Register(method, path, request)
}

func (ctl *Controller) PrintCalls() {
	ctl.calls.PrintCalls()
}

func (ctl *Controller) WasCalled(method, path string, expectedRequest []byte) bool {
	return ctl.calls.WasCalled(method, path, expectedRequest)
}

func (ctl *Controller) WasCalledRegex(methodRegex, pathRegex string, expectedRequest []byte) (bool, error) {
	return ctl.calls.WasCalledRegex(methodRegex, pathRegex, expectedRequest)
}

func (ctl *Controller) GetCalls(method, path string) [][]byte {
	return ctl.calls.GetCalls(method, path)
}
