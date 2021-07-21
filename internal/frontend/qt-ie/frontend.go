// Copyright (c) 2021 Proton Technologies AG
//
// This file is part of ProtonMail Bridge.
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

// +build build_qt

package qtie

import (
	"errors"
	"os"
	"sync"

	"github.com/ProtonMail/proton-bridge/internal/config/settings"
	"github.com/ProtonMail/proton-bridge/internal/events"
	qtcommon "github.com/ProtonMail/proton-bridge/internal/frontend/qt-common"
	"github.com/ProtonMail/proton-bridge/internal/frontend/types"
	"github.com/ProtonMail/proton-bridge/internal/importexport"
	"github.com/ProtonMail/proton-bridge/internal/locations"
	"github.com/ProtonMail/proton-bridge/internal/transfer"
	"github.com/ProtonMail/proton-bridge/internal/updater"
	"github.com/ProtonMail/proton-bridge/pkg/listener"

	"github.com/therecipe/qt/core"
	"github.com/therecipe/qt/gui"
	"github.com/therecipe/qt/qml"
	"github.com/therecipe/qt/widgets"

	"github.com/sirupsen/logrus"
	"github.com/skratchdot/open-golang/open"
)

var log = logrus.WithField("pkg", "frontend-qt-ie")

// FrontendQt is API between Import-Export and Qt
//
// With this interface it is possible to control Qt-Gui interface using pointers to
// Qt and QML objects. QML signals and slots are connected via methods of GoQMLInterface.
type FrontendQt struct {
	panicHandler  types.PanicHandler
	locations     *locations.Locations
	settings      *settings.Settings
	eventListener listener.Listener
	updater       types.Updater
	ie            types.ImportExporter

	App      *widgets.QApplication      // Main Application pointer
	View     *qml.QQmlApplicationEngine // QML engine pointer
	MainWin  *core.QObject              // Pointer to main window inside QML
	Qml      *GoQMLInterface            // Object accessible from both Go and QML for methods and signals
	Accounts qtcommon.Accounts          // Providing data for accounts ListView

	programName    string // App name
	programVersion string // Program version
	buildVersion   string // Program build version

	TransferRules *TransferRules
	ErrorList     *ErrorListModel // Providing data for error reporting

	transfer *transfer.Transfer
	progress *transfer.Progress

	restarter types.Restarter

	// saving most up-to-date update info to install it manually
	updateInfo updater.VersionInfo

	initializing       sync.WaitGroup
	initializationDone sync.Once
}

// New is constructor for Import-Export Qt-Go interface
func New(
	version, buildVersion, programName string,
	panicHandler types.PanicHandler,
	locations *locations.Locations,
	settings *settings.Settings,
	eventListener listener.Listener,
	updater types.Updater,
	ie types.ImportExporter,
	restarter types.Restarter,
) *FrontendQt {
	f := &FrontendQt{
		panicHandler:   panicHandler,
		locations:      locations,
		settings:       settings,
		programName:    programName,
		programVersion: "v" + version,
		eventListener:  eventListener,
		updater:        updater,
		buildVersion:   buildVersion,
		ie:             ie,
		restarter:      restarter,
	}

	// Initializing.Done is only called sync.Once. Please keep the increment
	// set to 1
	f.initializing.Add(1)

	log.Debugf("New Qt frontend: %p", f)
	return f
}

// Loop function for Import-Export interface. It runs QtExecute in main thread
// with no additional function.
func (f *FrontendQt) Loop() (err error) {
	err = f.QtExecute(func(f *FrontendQt) error { return nil })
	return err
}

func (f *FrontendQt) NotifyManualUpdate(update updater.VersionInfo, canInstall bool) {
	f.SetVersion(update)
	f.Qml.SetUpdateCanInstall(canInstall)
	f.Qml.NotifyManualUpdate()
}

func (f *FrontendQt) SetVersion(version updater.VersionInfo) {
	f.Qml.SetUpdateVersion(version.Version.String())
	f.Qml.SetUpdateLandingPage(version.LandingPage)
	f.Qml.SetUpdateReleaseNotesLink(version.ReleaseNotesPage)
	f.updateInfo = version
}

func (f *FrontendQt) NotifySilentUpdateInstalled() {
	//f.Qml.NotifySilentUpdateRestartNeeded()
}

func (f *FrontendQt) NotifySilentUpdateError(err error) {
	//f.Qml.NotifySilentUpdateError()
}

func (f *FrontendQt) watchEvents() {
	credentialsErrorCh := f.eventListener.ProvideChannel(events.CredentialsErrorEvent)
	internetOffCh := f.eventListener.ProvideChannel(events.InternetOffEvent)
	internetOnCh := f.eventListener.ProvideChannel(events.InternetOnEvent)
	secondInstanceCh := f.eventListener.ProvideChannel(events.SecondInstanceEvent)
	restartBridgeCh := f.eventListener.ProvideChannel(events.RestartBridgeEvent)
	addressChangedCh := f.eventListener.ProvideChannel(events.AddressChangedEvent)
	addressChangedLogoutCh := f.eventListener.ProvideChannel(events.AddressChangedLogoutEvent)
	logoutCh := f.eventListener.ProvideChannel(events.LogoutEvent)
	updateApplicationCh := f.eventListener.ProvideChannel(events.UpgradeApplicationEvent)
	newUserCh := f.eventListener.ProvideChannel(events.UserRefreshEvent)
	for {
		select {
		case <-credentialsErrorCh:
			f.Qml.NotifyHasNoKeychain()
		case <-internetOffCh:
			f.Qml.SetConnectionStatus(false)
		case <-internetOnCh:
			f.Qml.SetConnectionStatus(true)
		case <-secondInstanceCh:
			f.Qml.ShowWindow()
		case <-restartBridgeCh:
			f.restarter.SetToRestart()
			f.App.Quit()
		case address := <-addressChangedCh:
			f.Qml.NotifyAddressChanged(address)
		case address := <-addressChangedLogoutCh:
			f.Qml.NotifyAddressChangedLogout(address)
		case userID := <-logoutCh:
			user, err := f.ie.GetUser(userID)
			if err != nil {
				return
			}
			f.Qml.NotifyLogout(user.Username())
		case <-updateApplicationCh:
			f.Qml.ProcessFinished()
			f.Qml.NotifyForceUpdate()
		case <-newUserCh:
			f.Qml.LoadAccounts()
		}
	}
}

func (f *FrontendQt) qtSetupQmlAndStructures() {
	f.App = widgets.NewQApplication(len(os.Args), os.Args)
	// view
	f.View = qml.NewQQmlApplicationEngine(f.App)
	// Add Go-QML Import-Export
	f.Qml = NewGoQMLInterface(nil)
	f.Qml.SetFrontend(f) // provides access
	f.View.RootContext().SetContextProperty("go", f.Qml)

	// Add AccountsModel
	f.Accounts.SetupAccounts(f.Qml, f.ie, f.restarter)
	f.View.RootContext().SetContextProperty("accountsModel", f.Accounts.Model)

	// Add TransferRules structure
	f.TransferRules = NewTransferRules(nil)
	f.View.RootContext().SetContextProperty("transferRules", f.TransferRules)

	// Add error list modal
	f.ErrorList = NewErrorListModel(nil)
	f.View.RootContext().SetContextProperty("errorList", f.ErrorList)
	f.Qml.ConnectLoadImportReports(f.ErrorList.load)

	// Import path and load QML files
	f.View.AddImportPath("qrc:///")
	f.View.Load(core.NewQUrl3("qrc:/uiie.qml", 0))

	// TODO set the first start flag
	//log.Error("Get FirstStart: Not implemented")
	//if prefs.Get(prefs.FirstStart) == "true" {
	if false {
		f.Qml.SetIsFirstStart(true)
	} else {
		f.Qml.SetIsFirstStart(false)
	}
}

// QtExecute in main for starting Qt application
//
// It is needed to have just one Qt application per program (at least per same
// thread). This functions reads the main user interface defined in QML files.
// The files are appended to library by Qt-QRC.
func (f *FrontendQt) QtExecute(Procedure func(*FrontendQt) error) error {
	qtcommon.QtSetupCoreAndControls(f.programName, f.programVersion)
	f.qtSetupQmlAndStructures()
	// Check QML is loaded properly
	if len(f.View.RootObjects()) == 0 {
		//return errors.New(errors.ErrQApplication, "QML not loaded properly")
		return errors.New("QML not loaded properly")
	}
	// Obtain main window (need for invoke method)
	f.MainWin = f.View.RootObjects()[0]
	// Injected procedure for out-of-main-thread applications
	if err := Procedure(f); err != nil {
		return err
	}

	// List of used packages
	f.Qml.SetCredits(importexport.Credits)
	f.Qml.SetFullversion(f.buildVersion)

	//if f.settings.GetBool(settings.AutoUpdateKey) {
	//	f.Qml.SetIsAutoUpdate(true)
	//} else {
	//	f.Qml.SetIsAutoUpdate(false)
	//}

	go func() {
		defer f.panicHandler.HandlePanic()
		f.watchEvents()
	}()

	// Loop
	if ret := gui.QGuiApplication_Exec(); ret != 0 {
		//err := errors.New(errors.ErrQApplication, "Event loop ended with return value: %v", string(ret))
		err := errors.New("Event loop ended with return value: " + string(ret))
		log.Warnln("QGuiApplication_Exec: ", err)
		return err
	}
	log.Debug("Closing...")
	//prefs.Set(prefs.FirstStart, "false")
	return nil
}

func (f *FrontendQt) openLogs() {
	logsPath, err := f.locations.ProvideLogsPath()
	if err != nil {
		return
	}

	go open.Run(logsPath)
}

func (f *FrontendQt) openReport() {
	go open.Run(f.Qml.ImportLogFileName())
}

func (f *FrontendQt) openDownloadLink() {
	// NOTE: Fix this.
}

// sendImportReport sends an anonymized import or export report file to our customer support
func (f *FrontendQt) sendImportReport(address string) bool { // Todo_: Rename to sendReport?
	var accname = "No account logged in"
	if f.Accounts.Model.Count() > 0 {
		accname = f.Accounts.Model.Get(0).Account()
	}

	if f.progress == nil {
		log.Errorln("Failed to send process report: Missing progress")
		return false
	}

	report := f.progress.GenerateBugReport()

	if err := f.ie.ReportFile(
		core.QSysInfo_ProductType(),
		core.QSysInfo_PrettyProductName(),
		accname,
		address,
		report,
	); err != nil {
		log.Errorln("Failed to send process report:", err)
		return false
	}

	log.Info("Report send successfully")
	return true
}

// sendBug sends a bug report described by user to our customer support
func (f *FrontendQt) sendBug(description, emailClient, address string) bool {
	var accname = "No account logged in"
	if f.Accounts.Model.Count() > 0 {
		accname = f.Accounts.Model.Get(0).Account()
	}
	if accname == "" {
		accname = "Unknown account"
	}

	if err := f.ie.ReportBug(
		core.QSysInfo_ProductType(),
		core.QSysInfo_PrettyProductName(),
		description,
		accname,
		address,
		emailClient,
	); err != nil {
		log.Errorln("while sendBug:", err)
		return false
	}

	return true
}

//func (f *FrontendQt) toggleAutoUpdate() {
//	defer f.Qml.ProcessFinished()
//
//	if f.settings.GetBool(settings.AutoUpdateKey) {
//		f.settings.SetBool(settings.AutoUpdateKey, false)
//		f.Qml.SetIsAutoUpdate(false)
//	} else {
//		f.settings.SetBool(settings.AutoUpdateKey, true)
//		f.Qml.SetIsAutoUpdate(true)
//	}
//}

func (f *FrontendQt) showError(code int, err error) {
	f.Qml.SetErrorDescription(err.Error())
	log.WithField("code", code).Errorln(err.Error())
	f.Qml.NotifyError(code)
}

func (f *FrontendQt) emitEvent(evType, msg string) {
	f.eventListener.Emit(evType, msg)
}

func (f *FrontendQt) setProgressManager(progress *transfer.Progress) {
	f.progress = progress
	f.ErrorList.Progress = progress

	f.Qml.ConnectPauseProcess(func() { progress.Pause("paused") })
	f.Qml.ConnectResumeProcess(progress.Resume)
	f.Qml.ConnectCancelProcess(func() {
		progress.Stop()
	})
	f.Qml.SetProgress(0)

	go func() {
		log.Trace("Start reading updates")
		defer func() {
			log.Trace("Finishing reading updates")
			f.Qml.DisconnectPauseProcess()
			f.Qml.DisconnectResumeProcess()
			f.Qml.DisconnectCancelProcess()
			f.Qml.SetProgress(1)
			f.progress = nil
		}()

		updates := progress.GetUpdateChannel()
		for range updates {
			if progress.IsStopped() {
				break
			}
			counts := progress.GetCounts()
			f.Qml.SetTotal(int(counts.Total))
			f.Qml.SetProgressImported(int(counts.Imported))
			f.Qml.SetProgressSkipped(int(counts.Skipped))
			f.Qml.SetProgressFails(int(counts.Failed))
			f.Qml.SetProgressDescription(progress.PauseReason())
			if counts.Total > 0 {
				newProgress := counts.Progress()
				if newProgress >= 0 && newProgress != f.Qml.Progress() {
					f.Qml.SetProgress(newProgress)
					f.Qml.ProgressChanged(newProgress)
				}
			}
		}
		// Counts will add lost messages only once the progress is completeled.
		counts := progress.GetCounts()
		f.Qml.SetProgressImported(int(counts.Imported))
		f.Qml.SetProgressSkipped(int(counts.Skipped))
		f.Qml.SetProgressFails(int(counts.Failed))

		if err := progress.GetFatalError(); err != nil {
			f.Qml.SetProgressDescription(err.Error())
		} else {
			f.Qml.SetProgressDescription("")
		}
	}()
}

func (f *FrontendQt) startManualUpdate() {
	go func() {
		err := f.updater.InstallUpdate(f.updateInfo)

		if err != nil {
			logrus.WithError(err).Error("An error occurred while installing updates manually")
			f.Qml.NotifyManualUpdateError()
		} else {
			f.Qml.NotifyManualUpdateRestartNeeded()
		}
	}()
}

func (f *FrontendQt) checkIsLatestVersionAndUpdate() bool {
	version, err := f.updater.Check()

	if err != nil {
		logrus.WithError(err).Error("An error occurred while checking updates manually")
		f.Qml.NotifyManualUpdateError()
		return false
	}

	f.SetVersion(version)

	if !f.updater.IsUpdateApplicable(version) {
		logrus.Debug("No need to update")
		return true
	}

	logrus.WithField("version", version.Version).Info("An update is available")

	if !f.updater.CanInstall(version) {
		logrus.Debug("A manual update is required")
		f.NotifyManualUpdate(version, false)
		return false
	}

	f.NotifyManualUpdate(version, true)
	return false
}

func (s *FrontendQt) checkAndOpenReleaseNotes() {
	go func() {
		_ = s.checkIsLatestVersionAndUpdate()
		s.Qml.OpenReleaseNotesExternally()
	}()
}

func (s *FrontendQt) checkForUpdates() {
	go func() {
		if s.checkIsLatestVersionAndUpdate() {
			s.Qml.NotifyVersionIsTheLatest()
		}
	}()
}

func (f *FrontendQt) resetSource() {
	if f.transfer != nil {
		f.transfer.ResetRules()
		if err := f.loadStructuresForImport(); err != nil {
			log.WithError(err).Error("Cannot reload structures after reseting rules.")
		}
	}
}

func (f *FrontendQt) openLicenseFile() {
	go open.Run(f.locations.GetLicenseFilePath())
}

// getLocalVersionInfo is identical to bridge.
func (f *FrontendQt) getLocalVersionInfo() {
	// NOTE: Fix this.
}

// LeastUsedColor is intended to return color for creating a new inbox or label.
func (f *FrontendQt) leastUsedColor() string {
	if f.transfer == nil {
		log.Warnln("Getting least used color before transfer exist.")
		return "#7272a7"
	}

	m, err := f.transfer.TargetMailboxes()

	if err != nil {
		log.Errorln("Getting least used color:", err)
		f.showError(errUnknownError, err)
	}

	return transfer.LeastUsedColor(m)
}

// createLabelOrFolder performs an IE target mailbox creation.
func (f *FrontendQt) createLabelOrFolder(email, name, color string, isLabel bool, sourceID string) bool {
	// Prepare new mailbox.
	m := transfer.Mailbox{
		Name:        name,
		Color:       color,
		IsExclusive: !isLabel,
	}

	// Select least used color if no color given.
	if m.Color == "" {
		m.Color = f.leastUsedColor()
	}

	f.TransferRules.BeginResetModel()
	defer f.TransferRules.EndResetModel()

	// Create mailbox.
	m, err := f.transfer.CreateTargetMailbox(m)
	if err != nil {
		log.Errorln("Folder/Label creating:", err)
		err = errors.New(name + "\n" + err.Error()) // GUI splits by \n.
		if isLabel {
			f.showError(errCreateLabelFailed, err)
		} else {
			f.showError(errCreateFolderFailed, err)
		}
		return false
	}

	if sourceID == "-1" {
		f.transfer.SetGlobalMailbox(&m)
	} else {
		f.TransferRules.addTargetID(sourceID, m.Hash())
	}
	return true
}

func (f *FrontendQt) WaitUntilFrontendIsReady() {
	f.initializing.Wait()
}

// setGUIIsReady unlocks the WaitUntilFrontendIsReady.
func (f *FrontendQt) setGUIIsReady() {
	f.initializationDone.Do(func() {
		f.initializing.Done()
	})
}
