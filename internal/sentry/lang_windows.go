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

//go:build windows
// +build windows

package sentry

import (
	"syscall"
	"unsafe"
)

const (
	defaultLocaleUser   = "GetUserDefaultLocaleName"   // https://learn.microsoft.com/en-us/windows/win32/api/winnls/nf-winnls-getuserdefaultlocalename
	defaultLocaleSystem = "GetSystemDefaultLocaleName" // https://learn.microsoft.com/en-us/windows/win32/api/winnls/nf-winnls-getsystemdefaultlocalename
	localeNameMaxLength = 85                           // https://learn.microsoft.com/en-us/windows/win32/intl/locale-name-constants
)

func getLocale(dll *syscall.DLL, procName string) (string, error) {
	proc, err := dll.FindProc(procName)
	if err != nil {
		return "errProc", err
	}

	b := make([]uint16, localeNameMaxLength)

	r, _, err := proc.Call(uintptr(unsafe.Pointer(&b[0])), uintptr(localeNameMaxLength))
	if r == 0 || err != nil {
		return "errCall", err
	}

	return syscall.UTF16ToString(b), nil
}

func GetSystemLang() string {
	dll, err := syscall.LoadDLL("kernel32")
	if err != nil {
		return "errDll"
	}

	defer func() {
		_ = dll.Release()
	}()

	if lang, err := getLocale(dll, defaultLocaleUser); err == nil {
		return lang
	}

	lang, _ := getLocale(dll, defaultLocaleSystem)

	return lang
}
