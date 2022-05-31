// Copyright (c) 2022 Proton AG
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

package theme

import (
	"github.com/sirupsen/logrus"
	"golang.org/x/sys/windows/registry"
)

func detectSystemTheme() Theme {
	log := logrus.WithField("pkg", "theme")
	k, err := registry.OpenKey(
		registry.CURRENT_USER,
		`SOFTWARE\Microsoft\Windows\CurrentVersion\Themes\Personalize`,
		registry.QUERY_VALUE,
	)
	if err != nil {
		log.WithError(err).Error("Not able to open register")
		return Light
	}
	defer k.Close()

	i, _, err := k.GetIntegerValue("AppsUseLightTheme")
	if err != nil {
		log.WithError(err).Error("Cannot get value")
		return Light
	}

	if i == 0 {
		return Dark
	}

	return Light
}
