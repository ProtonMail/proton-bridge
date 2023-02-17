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

package main

// bridgelib export bridge go code as a C library. [More info](https://pkg.go.dev/cmd/go#hdr-Build_modes).
// on Windows bridge-gui is built using the MSVC compiler with cannot link against mingw static library,
// As a consequence, the library should be built as a C shared library (.dll/.so/.dylib depending on the platform):
//
// macOS:   go build -buildmode=c-shared -o bridgelib.dylib bridgelib.go
// Linux:   go build -buildmode=c-shared -o bridgelib.so bridgelib.go
// Windows: go build -buildmode=c-shared -o bridgelib.dll bridgelib.go
//
// In addition to the library file, the header will export a C header file container the relevant type declarations.
//
// Requirements to export a go library
//
// - The package name must be main, and as a consequence, must contain a main() function.
// - The package must import "C"
// - Functions to be exported must be annotated with the cgo //export comment.
//
// Heap allocated data such as go string needs to be released. On macOS and linux, the caller of the library function can call free/delete,
// but it crashes on Windows because of the VC++/MinGW incompatibility. As a consequence, the caller is responsible for freeing the memory
// using the DeleteCString function.

//#include <stdlib.h>
import "C"
import (
	"path/filepath"
	"runtime"
	"sync"
	"unsafe"

	"github.com/ProtonMail/proton-bridge/v3/internal/constants"
	"github.com/ProtonMail/proton-bridge/v3/internal/locations"
)

// main is empty but required.
func main() {
}

var locs *locations.DefaultProvider //nolint:gochecknoglobals
var locsMutex sync.Mutex            //nolint:gochecknoglobals

// GoOS returns the value of runtime.GOOS.
//
//export GoOS
func GoOS() *C.char {
	return C.CString(runtime.GOOS)
}

// UserCacheDir returns the path of the user's cache directory.
//
//export UserCacheDir
func UserCacheDir() *C.char {
	return withLocationProvider(func(loc *locations.DefaultProvider) string { return loc.UserCache() })
}

// UserConfigDir returns the path of the user's config directory.
//
//export UserConfigDir
func UserConfigDir() *C.char {
	return withLocationProvider(func(loc *locations.DefaultProvider) string { return loc.UserConfig() })
}

// UserDataDir returns the path of the user's data directory.
//
//export UserDataDir
func UserDataDir() *C.char {
	return withLocationProvider(func(loc *locations.DefaultProvider) string { return loc.UserData() })
}

// DeleteCString deletes a C-style string allocated by one of the library calls.
//
//export DeleteCString
func DeleteCString(cStr *C.char) {
	if cStr != nil {
		C.free(unsafe.Pointer(cStr))
	}
}

func withLocationProvider(fn func(provider *locations.DefaultProvider) string) *C.char {
	locsMutex.Lock()
	defer locsMutex.Unlock()

	if locs == nil {
		var err error
		if locs, err = locations.NewDefaultProvider(filepath.Join(constants.VendorName, constants.ConfigName)); err != nil {
			return nil
		}
	}

	return C.CString(fn(locs))
}
