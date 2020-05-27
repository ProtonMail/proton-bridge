// Copyright (c) 2020 Proton Technologies AG
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

// +build !nogui

package qtie

import (
	"errors"
	"os"
	"strconv"
	"time"

	"github.com/ProtonMail/proton-bridge/internal/events"
	qtcommon "github.com/ProtonMail/proton-bridge/internal/frontend/qt-common"
	"github.com/ProtonMail/proton-bridge/internal/frontend/types"
	"github.com/ProtonMail/proton-bridge/internal/transfer"
	"github.com/ProtonMail/proton-bridge/pkg/config"
	"github.com/ProtonMail/proton-bridge/pkg/listener"
	"github.com/ProtonMail/proton-bridge/pkg/updates"

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
	config        *config.Config
	eventListener listener.Listener
	updates       types.Updater
	ie            types.ImportExporter

	App      *widgets.QApplication      // Main Application pointer
	View     *qml.QQmlApplicationEngine // QML engine pointer
	MainWin  *core.QObject              // Pointer to main window inside QML
	Qml      *GoQMLInterface            // Object accessible from both Go and QML for methods and signals
	Accounts qtcommon.Accounts          // Providing data for accounts ListView

	programName    string // Program name
	programVersion string // Program version
	buildVersion   string // Program build version

	PMStructure       *FolderStructure // Providing data for account labels and folders for ProtonMail account
	ExternalStructure *FolderStructure // Providing data for account labels and folders for MBOX, EML or external IMAP account
	ErrorList         *ErrorListModel  // Providing data for error reporting

	transfer *transfer.Transfer

	notifyHasNoKeychain bool
}

// New is constructor for Import-Export Qt-Go interface
func New(
	version, buildVersion string,
	panicHandler types.PanicHandler,
	config *config.Config,
	eventListener listener.Listener,
	updates types.Updater,
	ie types.ImportExporter,
) *FrontendQt {
	f := &FrontendQt{
		panicHandler:   panicHandler,
		config:         config,
		programName:    "ProtonMail Import-Export",
		programVersion: "v" + version,
		eventListener:  eventListener,
		buildVersion:   buildVersion,
		updates:        updates,
		ie:             ie,
	}

	// Nicer string for OS
	currentOS := core.QSysInfo_PrettyProductName()
	ie.SetCurrentOS(currentOS)

	log.Debugf("New Qt frontend: %p", f)
	return f
}

// IsAppRestarting for Import-Export is always false i.e never restarts
func (s *FrontendQt) IsAppRestarting() bool {
	return false
}

// Loop function for Import-Export interface. It runs QtExecute in main thread
// with no additional function.
func (s *FrontendQt) Loop(setupError error) (err error) {
	if setupError != nil {
		s.notifyHasNoKeychain = true
	}
	go func() {
		defer s.panicHandler.HandlePanic()
		s.watchEvents()
	}()
	err = s.QtExecute(func(s *FrontendQt) error { return nil })
	return err
}

func (s *FrontendQt) watchEvents() {
	internetOffCh := qtcommon.MakeAndRegisterEvent(s.eventListener, events.InternetOffEvent)
	internetOnCh := qtcommon.MakeAndRegisterEvent(s.eventListener, events.InternetOnEvent)
	restartBridgeCh := qtcommon.MakeAndRegisterEvent(s.eventListener, events.RestartBridgeEvent)
	addressChangedCh := qtcommon.MakeAndRegisterEvent(s.eventListener, events.AddressChangedEvent)
	addressChangedLogoutCh := qtcommon.MakeAndRegisterEvent(s.eventListener, events.AddressChangedLogoutEvent)
	logoutCh := qtcommon.MakeAndRegisterEvent(s.eventListener, events.LogoutEvent)
	updateApplicationCh := qtcommon.MakeAndRegisterEvent(s.eventListener, events.UpgradeApplicationEvent)
	newUserCh := qtcommon.MakeAndRegisterEvent(s.eventListener, events.UserRefreshEvent)
	for {
		select {
		case <-internetOffCh:
			s.Qml.SetConnectionStatus(false)
		case <-internetOnCh:
			s.Qml.SetConnectionStatus(true)
		case <-restartBridgeCh:
			s.Qml.SetIsRestarting(true)
			s.App.Quit()
		case address := <-addressChangedCh:
			s.Qml.NotifyAddressChanged(address)
		case address := <-addressChangedLogoutCh:
			s.Qml.NotifyAddressChangedLogout(address)
		case userID := <-logoutCh:
			user, err := s.ie.GetUser(userID)
			if err != nil {
				return
			}
			s.Qml.NotifyLogout(user.Username())
		case <-updateApplicationCh:
			s.Qml.ProcessFinished()
			s.Qml.NotifyUpdate()
		case <-newUserCh:
			s.Qml.LoadAccounts()
		}
	}
}

func (s *FrontendQt) qtSetupQmlAndStructures() {
	s.App = widgets.NewQApplication(len(os.Args), os.Args)
	// view
	s.View = qml.NewQQmlApplicationEngine(s.App)
	// Add Go-QML Import-Export
	s.Qml = NewGoQMLInterface(nil)
	s.Qml.SetFrontend(s) // provides access
	s.View.RootContext().SetContextProperty("go", s.Qml)
	// Add AccountsModel
	s.Accounts.SetupAccounts(s.Qml, s.ie)
	s.View.RootContext().SetContextProperty("accountsModel", s.Accounts.Model)

	// Add ProtonMail FolderStructure
	s.PMStructure = NewFolderStructure(nil)
	s.View.RootContext().SetContextProperty("structurePM", s.PMStructure)

	// Add external FolderStructure
	s.ExternalStructure = NewFolderStructure(nil)
	s.View.RootContext().SetContextProperty("structureExternal", s.ExternalStructure)

	// Add error list modal
	s.ErrorList = NewErrorListModel(nil)
	s.View.RootContext().SetContextProperty("errorList", s.ErrorList)
	s.Qml.ConnectLoadImportReports(s.ErrorList.load)

	// Import path and load QML files
	s.View.AddImportPath("qrc:///")
	s.View.Load(core.NewQUrl3("qrc:/uiie.qml", 0))

	// TODO set the first start flag
	log.Error("Get FirstStart: Not implemented")
	//if prefs.Get(prefs.FirstStart) == "true" {
	if false {
		s.Qml.SetIsFirstStart(true)
	} else {
		s.Qml.SetIsFirstStart(false)
	}

	// Notify user about error during initialization.
	if s.notifyHasNoKeychain {
		s.Qml.NotifyHasNoKeychain()
	}
}

// QtExecute in main for starting Qt application
//
// It is needed to have just one Qt application per program (at least per same
// thread). This functions reads the main user interface defined in QML files.
// The files are appended to library by Qt-QRC.
func (s *FrontendQt) QtExecute(Procedure func(*FrontendQt) error) error {
	qtcommon.QtSetupCoreAndControls(s.programName, s.programVersion)
	s.qtSetupQmlAndStructures()
	// Check QML is loaded properly
	if len(s.View.RootObjects()) == 0 {
		//return errors.New(errors.ErrQApplication, "QML not loaded properly")
		return errors.New("QML not loaded properly")
	}
	// Obtain main window (need for invoke method)
	s.MainWin = s.View.RootObjects()[0]
	// Injected procedure for out-of-main-thread applications
	if err := Procedure(s); err != nil {
		return err
	}
	// Loop
	if ret := gui.QGuiApplication_Exec(); ret != 0 {
		//err := errors.New(errors.ErrQApplication, "Event loop ended with return value: %v", string(ret))
		err := errors.New("Event loop ended with return value: " + string(ret))
		log.Warnln("QGuiApplication_Exec: ", err)
		return err
	}
	log.Debug("Closing...")
	log.Error("Set FirstStart: Not implemented")
	//prefs.Set(prefs.FirstStart, "false")
	return nil
}

func (s *FrontendQt) openLogs() {
	go open.Run(s.config.GetLogDir())
}

func (s *FrontendQt) openReport() {
	go open.Run(s.Qml.ImportLogFileName())
}

func (s *FrontendQt) openDownloadLink() {
	go open.Run(s.updates.GetDownloadLink())
}

func (s *FrontendQt) sendImportReport(address, reportFile string) (isOK bool) {
	/*
		accname := "[No account logged in]"
		if s.Accounts.Count() > 0 {
			accname = s.Accounts.get(0).Account()
		}

		basename := filepath.Base(reportFile)
		req := pmapi.ReportReq{
			OS:          core.QSysInfo_ProductType(),
			OSVersion:   core.QSysInfo_PrettyProductName(),
			Title:       "[Import Export] Import report: " + basename,
			Description: "Sending import report file in attachment.",
			Username:    accname,
			Email:       address,
		}

		report, err := os.Open(reportFile)
		if err != nil {
			log.Errorln("report file open:", err)
			isOK = false
		}
		req.AddAttachment("log", basename, report)

		c := pmapi.NewClient(backend.APIConfig, "import_reporter")
		err = c.Report(req)
		if err != nil {
			log.Errorln("while sendReport:", err)
			isOK = false
			return
		}
		log.Infof("Report %q send successfully", basename)
		isOK = true
	*/
	return false
}

// sendBug is almost idetical to bridge
func (s *FrontendQt) sendBug(description, emailClient, address string) (isOK bool) {
	isOK = true
	var accname = "No account logged in"
	if s.Accounts.Model.Count() > 0 {
		accname = s.Accounts.Model.Get(0).Account()
	}
	if err := s.ie.ReportBug(
		core.QSysInfo_ProductType(),
		core.QSysInfo_PrettyProductName(),
		description,
		accname,
		address,
		emailClient,
	); err != nil {
		log.Errorln("while sendBug:", err)
		isOK = false
	}
	return
}

// checkInternet is almost idetical to bridge
func (s *FrontendQt) checkInternet() {
	s.Qml.SetConnectionStatus(s.ie.CheckConnection() == nil)
}

func (s *FrontendQt) showError(err error) {
	code := 0 // TODO err.Code()
	s.Qml.SetErrorDescription(err.Error())
	log.WithField("code", code).Errorln(err.Error())
	s.Qml.NotifyError(code)
}

func (s *FrontendQt) emitEvent(evType, msg string) {
	s.eventListener.Emit(evType, msg)
}

func (s *FrontendQt) setProgressManager(progress *transfer.Progress) {
	s.Qml.ConnectPauseProcess(func() { progress.Pause("user") })
	s.Qml.ConnectResumeProcess(progress.Resume)
	s.Qml.ConnectCancelProcess(func(clearUnfinished bool) {
		// TODO clear unfinished
		progress.Stop()
	})

	go func() {
		defer func() {
			s.Qml.DisconnectPauseProcess()
			s.Qml.DisconnectResumeProcess()
			s.Qml.DisconnectCancelProcess()
			s.Qml.SetProgress(1)
		}()

		//TODO get log file (in old code it was here, but this is ugly place probably somewhere else)
		updates := progress.GetUpdateChannel()
		for range updates {
			if progress.IsStopped() {
				break
			}
			failed, imported, _, _, total := progress.GetCounts()
			if total != 0 { // udate total
				s.Qml.SetTotal(int(total))
			}
			s.Qml.SetProgressFails(int(failed))
			s.Qml.SetProgressDescription(progress.PauseReason()) // TODO add description when changing folders?
			if total > 0 {
				newProgress := float32(imported+failed) / float32(total)
				if newProgress >= 0 && newProgress != s.Qml.Progress() {
					s.Qml.SetProgress(newProgress)
					s.Qml.ProgressChanged(newProgress)
				}
			}
		}

		// TODO fatal error?
	}()
}

// StartUpdate is identical to bridge
func (s *FrontendQt) StartUpdate() {
	progress := make(chan updates.Progress)
	go func() { // Update progress in QML.
		defer s.panicHandler.HandlePanic()
		for current := range progress {
			s.Qml.SetProgress(current.Processed)
			s.Qml.SetProgressDescription(strconv.Itoa(current.Description))
			// Error happend
			if current.Err != nil {
				log.Error("update progress: ", current.Err)
				s.Qml.UpdateFinished(true)
				return
			}
			// Finished everything OK.
			if current.Description >= updates.InfoQuitApp {
				s.Qml.UpdateFinished(false)
				time.Sleep(3 * time.Second) // Just notify.
				s.Qml.SetIsRestarting(current.Description == updates.InfoRestartApp)
				s.App.Quit()
				return
			}
		}
	}()
	go func() {
		defer s.panicHandler.HandlePanic()
		s.updates.StartUpgrade(progress)
	}()
}

// isNewVersionAvailable is identical to bridge
// return 0 when local version is fine
// return 1 when new version is available
func (s *FrontendQt) isNewVersionAvailable(showMessage bool) {
	go func() {
		defer s.Qml.ProcessFinished()
		isUpToDate, latestVersionInfo, err := s.updates.CheckIsUpToDate()
		if err != nil {
			log.Warnln("Cannot retrieve version info: ", err)
			s.checkInternet()
			return
		}
		s.Qml.SetConnectionStatus(true) // if we are here connection is ok
		if isUpToDate {
			s.Qml.SetUpdateState(StatusUpToDate)
			if showMessage {
				s.Qml.NotifyVersionIsTheLatest()
			}
			return
		}
		s.Qml.SetNewversion(latestVersionInfo.Version)
		s.Qml.SetChangelog(latestVersionInfo.ReleaseNotes)
		s.Qml.SetBugfixes(latestVersionInfo.ReleaseFixedBugs)
		s.Qml.SetLandingPage(latestVersionInfo.LandingPage)
		s.Qml.SetDownloadLink(latestVersionInfo.GetDownloadLink())
		s.Qml.SetUpdateState(StatusNewVersionAvailable)
	}()
}

func (s *FrontendQt) resetSource() {
	if s.transfer != nil {
		s.transfer.ResetRules()
		if err := s.loadStructuresForImport(); err != nil {
			log.WithError(err).Error("Cannot reload structures after reseting rules.")
		}
	}
}

// getLocalVersionInfo is identical to bridge.
func (s *FrontendQt) getLocalVersionInfo() {
	defer s.Qml.ProcessFinished()
	localVersion := s.updates.GetLocalVersion()
	s.Qml.SetNewversion(localVersion.Version)
	s.Qml.SetChangelog(localVersion.ReleaseNotes)
	s.Qml.SetBugfixes(localVersion.ReleaseFixedBugs)
}

// LeastUsedColor is intended to return color for creating a new inbox or label.
func (s *FrontendQt) leastUsedColor() string {
	if s.transfer == nil {
		log.Errorln("Getting least used color before transfer exist.")
		return "#7272a7"
	}

	m, err := s.transfer.TargetMailboxes()

	if err != nil {
		log.Errorln("Getting least used color:", err)
		s.showError(err)
	}

	return transfer.LeastUsedColor(m)
}

// createLabelOrFolder performs an IE target mailbox creation.
func (s *FrontendQt) createLabelOrFolder(email, name, color string, isLabel bool, sourceID string) bool {
	// Prepare new mailbox.
	m := transfer.Mailbox{
		Name:        name,
		Color:       color,
		IsExclusive: !isLabel,
	}

	// Select least used color if no color given.
	if m.Color == "" {
		m.Color = s.leastUsedColor()
	}

	// Create mailbox.
	newLabel, err := s.transfer.CreateTargetMailbox(m)

	if err != nil {
		log.Errorln("Folder/Label creating:", err)
		s.showError(err)
		return false
	}

	// TODO: notify UI of newly added folders/labels
	/*errc := s.PMStructure.Load(email, false)
	if errc != nil {
		s.showError(errc)
		return false
	}*/

	if sourceID != "" {
		if isLabel {
			s.ExternalStructure.addTargetLabelID(sourceID, newLabel.ID)
		} else {
			s.ExternalStructure.setTargetFolderID(sourceID, newLabel.ID)
		}
	}

	return true
}
