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

// Package qt is the Qt User interface for Desktop bridge.
//
// The FrontendQt implements Frontend interface: `frontend.go`.
// The helper functions are in `helpers.go`.
// Notification specific is written in `notification.go`.
// The AccountsModel is container providing account info to QML ListView.
//
// Since we are using QML there is only one Qt loop in  `ui.go`.
package qt

import (
	"errors"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/ProtonMail/go-autostart"
	"github.com/ProtonMail/proton-bridge/internal/bridge"
	"github.com/ProtonMail/proton-bridge/internal/events"
	"github.com/ProtonMail/proton-bridge/internal/frontend/autoconfig"
	"github.com/ProtonMail/proton-bridge/internal/frontend/types"
	"github.com/ProtonMail/proton-bridge/internal/preferences"
	"github.com/ProtonMail/proton-bridge/pkg/config"
	"github.com/ProtonMail/proton-bridge/pkg/logs"
	"github.com/ProtonMail/proton-bridge/pkg/ports"
	"github.com/ProtonMail/proton-bridge/pkg/useragent"

	//"github.com/ProtonMail/proton-bridge/pkg/keychain"
	"github.com/ProtonMail/proton-bridge/pkg/listener"
	"github.com/ProtonMail/proton-bridge/pkg/pmapi"
	"github.com/ProtonMail/proton-bridge/pkg/updates"
	"github.com/kardianos/osext"
	"github.com/skratchdot/open-golang/open"
	"github.com/therecipe/qt/core"
	"github.com/therecipe/qt/gui"
	"github.com/therecipe/qt/qml"
	"github.com/therecipe/qt/widgets"
)

var log = logs.GetLogEntry("frontend-qt")
var accountMutex = &sync.Mutex{}

// API between Bridge and Qt.
//
// With this interface it is possible to control Qt-Gui interface using pointers to
// Qt and QML objects. QML signals and slots are connected via methods of GoQMLInterface.
type FrontendQt struct {
	version           string
	buildVersion      string
	showWindowOnStart bool
	panicHandler      types.PanicHandler
	config            *config.Config
	preferences       *config.Preferences
	eventListener     listener.Listener
	updates           types.Updater
	bridge            types.Bridger
	noEncConfirmator  types.NoEncConfirmator

	App         *widgets.QApplication      // Main Application pointer.
	View        *qml.QQmlApplicationEngine // QML engine pointer.
	MainWin     *core.QObject              // Pointer to main window inside QML.
	Qml         *GoQMLInterface            // Object accessible from both Go and QML for methods and signals.
	Accounts    *AccountsModel             // Providing data for  accounts ListView.
	programName string                     // Program name (shown in taskbar).
	programVer  string                     // Program version (shown in help).

	authClient pmapi.Client

	auth *pmapi.Auth

	AutostartEntry *autostart.App

	// expand userID when added
	userIDAdded string

	notifyHasNoKeychain bool
}

// New returns a new Qt frontendend for the bridge.
func New(
	version,
	buildVersion string,
	showWindowOnStart bool,
	panicHandler types.PanicHandler,
	config *config.Config,
	preferences *config.Preferences,
	eventListener listener.Listener,
	updates types.Updater,
	bridge types.Bridger,
	noEncConfirmator types.NoEncConfirmator,
) *FrontendQt {
	prgName := "ProtonMail Bridge"
	tmp := &FrontendQt{
		version:           version,
		buildVersion:      buildVersion,
		showWindowOnStart: showWindowOnStart,
		panicHandler:      panicHandler,
		config:            config,
		preferences:       preferences,
		eventListener:     eventListener,
		updates:           updates,
		bridge:            bridge,
		noEncConfirmator:  noEncConfirmator,

		programName: prgName,
		programVer:  "v" + version,
		AutostartEntry: &autostart.App{
			Name:        prgName,
			DisplayName: prgName,
			Exec:        []string{"", "--no-window"},
		},
	}

	// Handle autostart if wanted.
	if p, err := osext.Executable(); err == nil {
		tmp.AutostartEntry.Exec[0] = p
		log.Info("Autostart ", p)
	} else {
		log.Error("Cannot get current executable path: ", err)
	}

	// Nicer string for OS.
	currentOS := core.QSysInfo_PrettyProductName()
	bridge.SetCurrentOS(currentOS)

	return tmp
}

// InstanceExistAlert is a global warning window indicating an instance already exists.
func (s *FrontendQt) InstanceExistAlert() {
	log.Warn("Instance already exists")
	s.QtSetupCoreAndControls()
	s.App = widgets.NewQApplication(len(os.Args), os.Args)
	s.View = qml.NewQQmlApplicationEngine(s.App)
	s.View.AddImportPath("qrc:///")
	s.View.Load(core.NewQUrl3("qrc:/BridgeUI/InstanceExistsWindow.qml", 0))
	_ = gui.QGuiApplication_Exec()
}

// Loop function for Bridge interface.
//
// It runs QtExecute in main thread with no additional function.
func (s *FrontendQt) Loop(credentialsError error) (err error) {
	if credentialsError != nil {
		s.notifyHasNoKeychain = true
	}
	go func() {
		defer s.panicHandler.HandlePanic()
		s.watchEvents()
	}()
	err = s.qtExecute(func(s *FrontendQt) error { return nil })
	return err
}

func (s *FrontendQt) watchEvents() {
	errorCh := s.getEventChannel(events.ErrorEvent)
	outgoingNoEncCh := s.getEventChannel(events.OutgoingNoEncEvent)
	noActiveKeyForRecipientCh := s.getEventChannel(events.NoActiveKeyForRecipientEvent)
	internetOffCh := s.getEventChannel(events.InternetOffEvent)
	internetOnCh := s.getEventChannel(events.InternetOnEvent)
	secondInstanceCh := s.getEventChannel(events.SecondInstanceEvent)
	restartBridgeCh := s.getEventChannel(events.RestartBridgeEvent)
	addressChangedCh := s.getEventChannel(events.AddressChangedEvent)
	addressChangedLogoutCh := s.getEventChannel(events.AddressChangedLogoutEvent)
	logoutCh := s.getEventChannel(events.LogoutEvent)
	updateApplicationCh := s.getEventChannel(events.UpgradeApplicationEvent)
	newUserCh := s.getEventChannel(events.UserRefreshEvent)
	certIssue := s.getEventChannel(events.TLSCertIssue)
	for {
		select {
		case errorDetails := <-errorCh:
			imapIssue := strings.Contains(errorDetails, "IMAP failed")
			smtpIssue := strings.Contains(errorDetails, "SMTP failed")
			s.Qml.NotifyPortIssue(imapIssue, smtpIssue)
		case idAndSubject := <-outgoingNoEncCh:
			idAndSubjectSlice := strings.SplitN(idAndSubject, ":", 2)
			messageID := idAndSubjectSlice[0]
			subject := idAndSubjectSlice[1]
			s.Qml.ShowOutgoingNoEncPopup(messageID, subject)
		case email := <-noActiveKeyForRecipientCh:
			s.Qml.ShowNoActiveKeyForRecipient(email)
		case <-internetOffCh:
			s.Qml.SetConnectionStatus(false)
		case <-internetOnCh:
			s.Qml.SetConnectionStatus(true)
		case <-secondInstanceCh:
			s.Qml.ShowWindow()
		case <-restartBridgeCh:
			s.Qml.SetIsRestarting(true)
			s.App.Quit()
		case address := <-addressChangedCh:
			s.Qml.NotifyAddressChanged(address)
		case address := <-addressChangedLogoutCh:
			s.Qml.NotifyAddressChangedLogout(address)
		case userID := <-logoutCh:
			user, err := s.bridge.GetUser(userID)
			if err != nil {
				return
			}
			s.Qml.NotifyLogout(user.Username())
		case <-updateApplicationCh:
			s.Qml.ProcessFinished()
			s.Qml.NotifyUpdate()
		case <-newUserCh:
			s.Qml.LoadAccounts()
		case <-certIssue:
			s.Qml.ShowCertIssue()
		}
	}
}

func (s *FrontendQt) getEventChannel(event string) <-chan string {
	ch := make(chan string)
	s.eventListener.Add(event, ch)
	return ch
}

// Loop function for tests.
//
// It runs QtExecute in new thread with function returning itself after setup.
// Therefore it is possible to run tests on background.
func (s *FrontendQt) Start() (err error) {
	uiready := make(chan *FrontendQt)
	go func() {
		err := s.qtExecute(func(self *FrontendQt) error {
			// NOTE: Trick to send back UI by channel to access functionality
			// inside application thread. Other only uninitialized `ui` is visible.
			uiready <- self
			return nil
		})
		if err != nil {
			log.Error(err)
		}
		uiready <- nil
	}()

	// Receive UI pointer and set all pointers.
	running := <-uiready
	s.App = running.App
	s.View = running.View
	s.MainWin = running.MainWin
	return nil
}

func (s *FrontendQt) IsAppRestarting() bool {
	return s.Qml.IsRestarting()
}

// InvMethod runs the function with name `method` defined in RootObject of the QML.
// Used for tests.
func (s *FrontendQt) InvMethod(method string) error {
	arg := core.NewQGenericArgument("", nil)
	PauseLong()
	isGoodMethod := core.QMetaObject_InvokeMethod4(s.MainWin, method, arg, arg, arg, arg, arg, arg, arg, arg, arg, arg)
	if isGoodMethod == false {
		return errors.New("Wrong method " + method)
	}
	return nil
}

// QtSetupCoreAndControls hanldes global setup of Qt.
// Should be called once per program. Probably once per thread is fine.
func (s *FrontendQt) QtSetupCoreAndControls() {
	installMessageHandler()
	// Core setup.
	core.QCoreApplication_SetApplicationName(s.programName)
	core.QCoreApplication_SetApplicationVersion(s.programVer)
	// High DPI scaling for windows.
	core.QCoreApplication_SetAttribute(core.Qt__AA_EnableHighDpiScaling, false)
	// Software OpenGL: to avoid dedicated GPU.
	core.QCoreApplication_SetAttribute(core.Qt__AA_UseSoftwareOpenGL, true)
	// Basic style for QuickControls2 objects.
	//quickcontrols2.QQuickStyle_SetStyle("material")
}

// qtExecute is the main function for starting the Qt application.
//
// It is better to have just one Qt application per program (at least per same
// thread). This functions reads the main user interface defined in QML files.
// The files are appended to library by Qt-QRC.
func (s *FrontendQt) qtExecute(Procedure func(*FrontendQt) error) error {
	s.QtSetupCoreAndControls()
	s.App = widgets.NewQApplication(len(os.Args), os.Args)
	if runtime.GOOS == "linux" { // Fix default font.
		s.App.SetFont(gui.NewQFont2(FcMatchSans(), 12, int(gui.QFont__Normal), false), "")
	}
	s.App.SetQuitOnLastWindowClosed(false) // Just to make sure it's not closed.

	s.View = qml.NewQQmlApplicationEngine(s.App)
	// Add Go-QML bridge.
	s.Qml = NewGoQMLInterface(nil)
	s.Qml.SetIsShownOnStart(s.showWindowOnStart)
	s.Qml.SetFrontend(s) // provides access
	s.View.RootContext().SetContextProperty("go", s.Qml)

	// Set first start flag.
	s.Qml.SetIsFirstStart(s.preferences.GetBool(preferences.FirstStartKey))
	// Don't repeat next start.
	s.preferences.SetBool(preferences.FirstStartKey, false)

	// Check if it is first start after update (fresh version).
	lastVersion := s.preferences.Get(preferences.LastVersionKey)
	s.Qml.SetIsFreshVersion(lastVersion != "" && s.version != lastVersion)
	s.preferences.Set(preferences.LastVersionKey, s.version)

	// Add AccountsModel.
	s.Accounts = NewAccountsModel(nil)
	s.View.RootContext().SetContextProperty("accountsModel", s.Accounts)
	// Import path and load QML files.
	s.View.AddImportPath("qrc:///")
	s.View.Load(core.NewQUrl3("qrc:/ui.qml", 0))

	// List of used packages.
	s.Qml.SetCredits(bridge.Credits)
	s.Qml.SetFullversion(s.buildVersion)

	// Autostart.
	if s.Qml.IsFirstStart() {
		if s.AutostartEntry.IsEnabled() {
			if err := s.AutostartEntry.Disable(); err != nil {
				log.Error("First disable ", err)
				s.autostartError(err)
			}
		}
		s.toggleAutoStart()
	}
	if s.AutostartEntry.IsEnabled() {
		s.Qml.SetIsAutoStart(true)
	} else {
		s.Qml.SetIsAutoStart(false)
	}

	if s.preferences.GetBool(preferences.AllowProxyKey) {
		s.Qml.SetIsProxyAllowed(true)
	} else {
		s.Qml.SetIsProxyAllowed(false)
	}

	// Notify user about error during initialization.
	if s.notifyHasNoKeychain {
		s.Qml.NotifyHasNoKeychain()
	}

	s.eventListener.RetryEmit(events.TLSCertIssue)
	s.eventListener.RetryEmit(events.ErrorEvent)

	// Set reporting of outgoing email without encryption.
	s.Qml.SetIsReportingOutgoingNoEnc(s.preferences.GetBool(preferences.ReportOutgoingNoEncKey))

	// IMAP/SMTP ports.
	s.Qml.SetIsDefaultPort(
		s.config.GetDefaultIMAPPort() == s.preferences.GetInt(preferences.IMAPPortKey) &&
			s.config.GetDefaultSMTPPort() == s.preferences.GetInt(preferences.SMTPPortKey),
	)

	// Check QML is loaded properly.
	if len(s.View.RootObjects()) == 0 {
		return errors.New("QML not loaded properly")
	}

	// Obtain main window (need for invoke method).
	s.MainWin = s.View.RootObjects()[0]
	SetupSystray(s)

	// Injected procedure for out-of-main-thread applications.
	if err := Procedure(s); err != nil {
		return err
	}

	// Loop
	if ret := gui.QGuiApplication_Exec(); ret != 0 {
		err := errors.New("Event loop ended with return value:" + string(ret))
		log.Warn("QGuiApplication_Exec: ", err)
		return err
	}
	HideSystray()
	return nil
}

func (s *FrontendQt) openLogs() {
	go open.Run(s.config.GetLogDir())
}

// Check version in separate goroutine to not block the GUI (avoid program not responding message).
func (s *FrontendQt) isNewVersionAvailable(showMessage bool) {
	go func() {
		defer s.panicHandler.HandlePanic()
		defer s.Qml.ProcessFinished()
		isUpToDate, latestVersionInfo, err := s.updates.CheckIsBridgeUpToDate()
		if err != nil {
			log.Warn("Can not retrieve version info: ", err)
			s.checkInternet()
			return
		}
		s.Qml.SetConnectionStatus(true) // If we are here connection is ok.
		if isUpToDate {
			s.Qml.SetUpdateState("upToDate")
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
		s.Qml.ShowWindow()
		s.Qml.SetUpdateState("oldVersion")
	}()
}

func (s *FrontendQt) getLocalVersionInfo() {
	defer s.Qml.ProcessFinished()
	localVersion := s.updates.GetLocalVersion()
	s.Qml.SetNewversion(localVersion.Version)
	s.Qml.SetChangelog(localVersion.ReleaseNotes)
	s.Qml.SetBugfixes(localVersion.ReleaseFixedBugs)
}

func (s *FrontendQt) sendBug(description, client, address string) (isOK bool) {
	isOK = true
	var accname = "No account logged in"
	if s.Accounts.Count() > 0 {
		accname = s.Accounts.get(0).Account()
	}
	if err := s.bridge.ReportBug(
		core.QSysInfo_ProductType(),
		core.QSysInfo_PrettyProductName(),
		description,
		accname,
		address,
		client,
	); err != nil {
		log.Error("while sendBug: ", err)
		isOK = false
	}
	return
}

func (s *FrontendQt) getLastMailClient() string {
	return s.bridge.GetCurrentClient()
}

func (s *FrontendQt) configureAppleMail(iAccount, iAddress int) {
	acc := s.Accounts.get(iAccount)

	user, err := s.bridge.GetUser(acc.UserID())
	if err != nil {
		log.Warn("UserConfigFromKeychain failed: ", acc.Account(), err)
		s.SendNotification(TabAccount, s.Qml.GenericErrSeeLogs())
		return
	}

	imapPort := s.preferences.GetInt(preferences.IMAPPortKey)
	imapSSL := false
	smtpPort := s.preferences.GetInt(preferences.SMTPPortKey)
	smtpSSL := s.preferences.GetBool(preferences.SMTPSSLKey)

	// If configuring apple mail for Catalina or newer, users should use SSL.
	doRestart := false
	if !smtpSSL && useragent.IsCatalinaOrNewer() {
		smtpSSL = true
		s.preferences.SetBool(preferences.SMTPSSLKey, true)
		log.Warn("Detected Catalina or newer with bad SMTP SSL settings, now using SSL, bridge needs to restart")
		doRestart = true
	}

	for _, autoConf := range autoconfig.Available() {
		if err := autoConf.Configure(imapPort, smtpPort, imapSSL, smtpSSL, user, iAddress); err != nil {
			log.Warn("Autoconfig failed: ", autoConf.Name(), err)
			s.SendNotification(TabAccount, s.Qml.GenericErrSeeLogs())
			return
		}
	}

	if doRestart {
		time.Sleep(2 * time.Second)
		s.Qml.SetIsRestarting(true)
		s.App.Quit()
	}
	return
}

func (s *FrontendQt) toggleAutoStart() {
	defer s.Qml.ProcessFinished()
	var err error
	if s.AutostartEntry.IsEnabled() {
		err = s.AutostartEntry.Disable()
	} else {
		err = s.AutostartEntry.Enable()
	}
	if err != nil {
		log.Error("Enable autostart: ", err)
		s.autostartError(err)
	}
	if s.AutostartEntry.IsEnabled() {
		s.Qml.SetIsAutoStart(true)
	} else {
		s.Qml.SetIsAutoStart(false)
	}
}

func (s *FrontendQt) toggleAllowProxy() {
	defer s.Qml.ProcessFinished()

	if s.preferences.GetBool(preferences.AllowProxyKey) {
		s.preferences.SetBool(preferences.AllowProxyKey, false)
		s.bridge.DisallowProxy()
		s.Qml.SetIsProxyAllowed(false)
	} else {
		s.preferences.SetBool(preferences.AllowProxyKey, true)
		s.bridge.AllowProxy()
		s.Qml.SetIsProxyAllowed(true)
	}
}

func (s *FrontendQt) getIMAPPort() string {
	return s.preferences.Get(preferences.IMAPPortKey)
}

func (s *FrontendQt) getSMTPPort() string {
	return s.preferences.Get(preferences.SMTPPortKey)
}

// Return 0 -- port is free to use for server.
// Return 1 -- port is occupied.
func (s *FrontendQt) isPortOpen(portStr string) int {
	portInt, err := strconv.Atoi(portStr)
	if err != nil {
		return 1
	}
	if !ports.IsPortFree(portInt) {
		return 1
	}
	return 0
}

func (s *FrontendQt) setPortsAndSecurity(imapPort, smtpPort string, useSTARTTLSforSMTP bool) {
	s.preferences.Set(preferences.IMAPPortKey, imapPort)
	s.preferences.Set(preferences.SMTPPortKey, smtpPort)
	s.preferences.SetBool(preferences.SMTPSSLKey, !useSTARTTLSforSMTP)
}

func (s *FrontendQt) isSMTPSTARTTLS() bool {
	return !s.preferences.GetBool(preferences.SMTPSSLKey)
}

func (s *FrontendQt) checkInternet() {
	s.Qml.SetConnectionStatus(s.bridge.CheckConnection() == nil)
}

func (s *FrontendQt) switchAddressModeUser(iAccount int) {
	defer s.Qml.ProcessFinished()
	userID := s.Accounts.get(iAccount).UserID()
	user, err := s.bridge.GetUser(userID)
	if err != nil {
		log.Error("Get user for switch address mode failed: ", err)
		s.SendNotification(TabAccount, s.Qml.GenericErrSeeLogs())
		return
	}
	if err := user.SwitchAddressMode(); err != nil {
		log.Error("Switch address mode failed: ", err)
		s.SendNotification(TabAccount, s.Qml.GenericErrSeeLogs())
		return
	}
	s.userIDAdded = userID
}

func (s *FrontendQt) autostartError(err error) {
	if strings.Contains(err.Error(), "permission denied") {
		s.Qml.FailedAutostartCode("permission")
	} else if strings.Contains(err.Error(), "error code: 0x") {
		errorCode := err.Error()
		errorCode = errorCode[len(errorCode)-8:]
		s.Qml.FailedAutostartCode(errorCode)
	} else {
		s.Qml.FailedAutostartCode("")
	}
}

func (s *FrontendQt) toggleIsReportingOutgoingNoEnc() {
	shouldReport := !s.Qml.IsReportingOutgoingNoEnc()
	s.preferences.SetBool(preferences.ReportOutgoingNoEncKey, shouldReport)
	s.Qml.SetIsReportingOutgoingNoEnc(shouldReport)
}

func (s *FrontendQt) shouldSendAnswer(messageID string, shouldSend bool) {
	s.noEncConfirmator.ConfirmNoEncryption(messageID, shouldSend)
}

func (s *FrontendQt) saveOutgoingNoEncPopupCoord(x, y float32) {
	//prefs.SetFloat(prefs.OutgoingNoEncPopupCoordX, x)
	//prefs.SetFloat(prefs.OutgoingNoEncPopupCoordY, y)
}

func (s *FrontendQt) StartUpdate() {
	progress := make(chan updates.Progress)
	go func() { // Update progress in QML.
		defer s.panicHandler.HandlePanic()
		for current := range progress {
			s.Qml.SetProgress(current.Processed)
			s.Qml.SetProgressDescription(current.Description)
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
