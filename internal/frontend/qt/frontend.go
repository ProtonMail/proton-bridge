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

//go:build build_qt
// +build build_qt

// Package qt provides communication between Qt/QML frontend and Go backend
package qt

import (
	"fmt"
	"sync"

	"github.com/ProtonMail/proton-bridge/v2/internal/config/settings"
	"github.com/ProtonMail/proton-bridge/v2/internal/config/useragent"
	"github.com/ProtonMail/proton-bridge/v2/internal/frontend/types"
	"github.com/ProtonMail/proton-bridge/v2/internal/locations"
	"github.com/ProtonMail/proton-bridge/v2/internal/updater"
	"github.com/ProtonMail/proton-bridge/v2/pkg/listener"
	"github.com/ProtonMail/proton-bridge/v2/pkg/pmapi"
	"github.com/sirupsen/logrus"
	"github.com/therecipe/qt/core"
	"github.com/therecipe/qt/qml"
	"github.com/therecipe/qt/widgets"
)

type FrontendQt struct {
	programName, programVersion string

	panicHandler     types.PanicHandler
	locations        *locations.Locations
	settings         *settings.Settings
	eventListener    listener.Listener
	updater          types.Updater
	userAgent        *useragent.UserAgent
	bridge           types.Bridger
	noEncConfirmator types.NoEncConfirmator
	restarter        types.Restarter
	showOnStartup    bool

	authClient pmapi.Client
	auth       *pmapi.Auth
	password   []byte

	newVersionInfo updater.VersionInfo

	log                *logrus.Entry
	initializing       sync.WaitGroup
	initializationDone sync.Once
	firstTimeAutostart sync.Once

	app    *widgets.QApplication
	engine *qml.QQmlEngine
	qml    *QMLBackend
}

func New(
	version,
	buildVersion,
	programName string,
	showWindowOnStart bool,
	panicHandler types.PanicHandler,
	locations *locations.Locations,
	settings *settings.Settings,
	eventListener listener.Listener,
	updater types.Updater,
	userAgent *useragent.UserAgent,
	bridge types.Bridger,
	_ types.NoEncConfirmator,
	restarter types.Restarter,
) *FrontendQt {
	userAgent.SetPlatform(core.QSysInfo_PrettyProductName())

	f := &FrontendQt{
		programName:    programName,
		programVersion: version,
		log:            logrus.WithField("pkg", "frontend/qt"),

		panicHandler:  panicHandler,
		locations:     locations,
		settings:      settings,
		eventListener: eventListener,
		updater:       updater,
		userAgent:     userAgent,
		bridge:        bridge,
		restarter:     restarter,
		showOnStartup: showWindowOnStart,
	}

	// Initializing.Done is only called sync.Once. Please keep the increment
	// set to 1
	f.initializing.Add(1)

	return f
}

func (f *FrontendQt) Loop() error {
	err := f.initiateQtApplication()
	if err != nil {
		return err
	}

	go func() {
		defer f.panicHandler.HandlePanic()
		f.watchEvents()
	}()

	// Set whether this is the first time GUI starts.
	f.qml.SetIsFirstGUIStart(f.settings.GetBool(settings.FirstStartGUIKey))
	defer func() {
		f.settings.SetBool(settings.FirstStartGUIKey, false)
	}()

	if ret := f.app.Exec(); ret != 0 {
		err := fmt.Errorf("Event loop ended with return value: %v", ret)
		f.log.Warn("App exec", err)
		return err
	}

	return nil
}

func (f *FrontendQt) NotifyManualUpdate(version updater.VersionInfo, canInstall bool) {
	if canInstall {
		f.qml.UpdateManualReady(version.Version.String())
	} else {
		f.qml.UpdateManualError()
	}
}

func (f *FrontendQt) SetVersion(version updater.VersionInfo) {
	f.newVersionInfo = version
	f.qml.SetReleaseNotesLink(core.NewQUrl3(version.ReleaseNotesPage, core.QUrl__TolerantMode))
	f.qml.SetLandingPageLink(core.NewQUrl3(version.LandingPage, core.QUrl__TolerantMode))
}

func (f *FrontendQt) NotifySilentUpdateInstalled() {
	f.qml.UpdateSilentRestartNeeded()
}

func (f *FrontendQt) NotifySilentUpdateError(err error) {
	f.log.WithError(err).Warn("In-app update failed, asking for manual.")
	f.qml.UpdateManualError()
}

func (f *FrontendQt) WaitUntilFrontendIsReady() {
	f.initializing.Wait()
}
