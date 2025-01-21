// Copyright (c) 2025 Proton AG
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

//go:build darwin

package versioncompare

import (
	"testing"

	"github.com/elastic/go-sysinfo"
	"github.com/elastic/go-sysinfo/types"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func Test_IsHost_EligibleDarwin(t *testing.T) {
	host, err := sysinfo.Host()
	require.NoError(t, err)

	testData := []struct {
		sysVer             SystemVersion
		getHostOsVersionFn func(host types.Host) string
		shouldContinue     bool
		wantErr            bool
	}{
		{
			sysVer:             SystemVersion{Minimum: "9.5", Maximum: "12.0"},
			getHostOsVersionFn: func(_ types.Host) string { return "10.0" },
			shouldContinue:     true,
		},
		{
			sysVer:             SystemVersion{Minimum: "9.5.5.5", Maximum: "10.1.1.0"},
			getHostOsVersionFn: func(_ types.Host) string { return "10.0" },
			shouldContinue:     true,
		},
		{
			sysVer:             SystemVersion{Minimum: "10.0.1", Maximum: "12.0"},
			getHostOsVersionFn: func(_ types.Host) string { return "10.0" },
			shouldContinue:     true,
		},
		{
			sysVer:             SystemVersion{Minimum: "11.0", Maximum: "12.0"},
			getHostOsVersionFn: func(_ types.Host) string { return "10.0" },
			shouldContinue:     false,
			wantErr:            true,
		},
		{
			sysVer:             SystemVersion{Minimum: "11.1.0", Maximum: "12.0.0"},
			getHostOsVersionFn: func(_ types.Host) string { return "11.0.0" },
			shouldContinue:     false,
			wantErr:            true,
		},
		{
			sysVer:             SystemVersion{Minimum: "10.0", Maximum: "12.0"},
			getHostOsVersionFn: func(_ types.Host) string { return "12.0" },
			shouldContinue:     true,
		},
		{
			sysVer:             SystemVersion{Minimum: "11.1.0", Maximum: "12.0.0"},
			getHostOsVersionFn: func(_ types.Host) string { return "" },
			shouldContinue:     true,
			wantErr:            true,
		},
		{
			sysVer:             SystemVersion{Minimum: "11.1.0", Maximum: "12.0.0"},
			getHostOsVersionFn: func(_ types.Host) string { return "a.b.c" },
			shouldContinue:     true,
			wantErr:            true,
		},
		{
			sysVer:             SystemVersion{},
			getHostOsVersionFn: func(_ types.Host) string { return "1.2.3" },
			shouldContinue:     true,
			wantErr:            false,
		},
	}

	for _, test := range testData {
		l := logrus.WithField("test", "test")
		shouldContinue, err := test.sysVer.IsHostVersionEligible(l, host, test.getHostOsVersionFn)

		if test.wantErr {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
		}

		require.Equal(t, test.shouldContinue, shouldContinue)
	}
}
