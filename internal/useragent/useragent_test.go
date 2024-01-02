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
// along with Proton Mail Bridge. If not, see <https://www.gnu.org/licenses/>.

package useragent

import (
	"fmt"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUserAgent(t *testing.T) {
	tests := []struct {
		name, version, platform string
		want                    string
	}{
		// No name/version, no platform.
		{
			want: fmt.Sprintf("NoClient/0.0.1 (%v)", runtime.GOOS),
		},

		// No name/version, with platform.
		{
			platform: "macOS 10.15",
			want:     "NoClient/0.0.1 (macOS 10.15)",
		},

		// With name/version, with platform.
		{
			name:     "Mac OS X Mail",
			version:  "1.0.0",
			platform: "macOS 10.15",
			want:     "Mac OS X Mail/1.0.0 (macOS 10.15)",
		},

		// With name/version, with platform.
		{
			name:     "Mac OS X Mail",
			version:  "13.4 (3608.120.23.2.4)",
			platform: "macOS 10.15",
			want:     "Mac OS X Mail/13.4-3608.120.23.2.4 (macOS 10.15)",
		},

		// With name/version, with platform.
		{
			name:     "Thunderbird",
			version:  "78.6.1",
			platform: "Windows 10 (10.0)",
			want:     "Thunderbird/78.6.1 (Windows 10 (10.0))",
		},
	}

	for _, test := range tests {
		test := test

		t.Run(test.want, func(t *testing.T) {
			ua := New()

			if test.name != "" && test.version != "" {
				ua.SetClient(test.name, test.version)
			}

			if test.platform != "" {
				ua.SetPlatform(test.platform)
			}

			assert.Equal(t, test.want, ua.GetUserAgent())
		})
	}
}
