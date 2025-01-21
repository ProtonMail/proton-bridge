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
	"fmt"
	"strconv"
	"strings"

	"github.com/elastic/go-sysinfo/types"
	"github.com/sirupsen/logrus"
)

func (sysVer SystemVersion) IsHostVersionEligible(log *logrus.Entry, host types.Host, getHostOSVersion func(host types.Host) string) (bool, error) {
	if sysVer.Minimum == "" && sysVer.Maximum == "" {
		return true, nil
	}

	// We use getHostOSVersion simply for testing; It's passed via Bridge.
	var hostVersion string
	if getHostOSVersion == nil {
		hostVersion = host.Info().OS.Version
	} else {
		hostVersion = getHostOSVersion(host)
	}

	log.Debugf("Checking host OS and update system version requirements. Host: %s; Maximum: %s; Minimum: %s",
		hostVersion, sysVer.Maximum, sysVer.Minimum)

	hostVersionArr := strings.Split(hostVersion, ".")
	if len(hostVersionArr) == 0 || hostVersion == "" {
		return true, fmt.Errorf("could not get host version: %v", hostVersion)
	}

	hostVersionArrInt := make([]int, len(hostVersionArr))
	for i := 0; i < len(hostVersionArr); i++ {
		hostNum, err := strconv.Atoi(hostVersionArr[i])
		if err != nil {
			// If we receive an alphanumeric version - we should continue with the update and stop checking for
			// OS version requirements.
			return true, fmt.Errorf("invalid host version number: %s - %s", hostVersionArr[i], hostVersion)
		}
		hostVersionArrInt[i] = hostNum
	}

	if sysVer.Minimum != "" {
		pass, err := compareMinimumVersion(hostVersionArrInt, sysVer.Minimum)
		if err != nil {
			return false, err
		}

		if !pass {
			return false, fmt.Errorf("host version is below minimum: hostVersion %v - minimumVersion %v", hostVersion, sysVer.Minimum)
		}
	}

	if sysVer.Maximum != "" {
		pass, err := compareMaximumVersion(hostVersionArrInt, sysVer.Maximum)
		if err != nil {
			return false, err
		}

		if !pass {
			return false, fmt.Errorf("host version is above maximum version: hostVersion %v - minimumVersion %v", hostVersion, sysVer.Maximum)
		}
	}

	return true, nil
}

func compareMinimumVersion(hostVersionArr []int, minVersion string) (bool, error) {
	minVersionArr := strings.Split(minVersion, ".")
	iterationDepth := min(len(hostVersionArr), len(minVersionArr))

	for i := 0; i < iterationDepth; i++ {
		hostNum := hostVersionArr[i]

		minNum, err := strconv.Atoi(minVersionArr[i])
		if err != nil {
			return false, fmt.Errorf("invalid minimum version number: %s - %s", minVersionArr[i], minVersion)
		}

		if hostNum < minNum {
			return false, nil
		}

		if hostNum > minNum {
			return true, nil
		}
	}

	return true, nil // minVersion is inclusive
}

func compareMaximumVersion(hostVersionArr []int, maxVersion string) (bool, error) {
	maxVersionArr := strings.Split(maxVersion, ".")
	iterationDepth := min(len(maxVersionArr), len(hostVersionArr))

	for i := 0; i < iterationDepth; i++ {
		hostNum := hostVersionArr[i]

		maxNum, err := strconv.Atoi(maxVersionArr[i])
		if err != nil {
			return false, fmt.Errorf("invalid maximum version number: %s - %s", maxVersionArr[i], maxVersion)
		}

		if hostNum > maxNum {
			return false, nil
		}

		if hostNum < maxNum {
			return true, nil
		}
	}

	return true, nil // maxVersion is inclusive
}
