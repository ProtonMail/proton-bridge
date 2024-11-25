using System;
using System.Collections.Generic;
using System.Linq;
using System.Text;
using System.Threading.Tasks;
using ProtonMailBridge.UI.Tests.Results;
using ProtonMailBridge.UI.Tests.Windows;
using NUnit.Framework;
using ProtonMailBridge.UI.Tests.TestsHelper;
using FlaUI.Core.Input;
using FlaUI.Core.AutomationElements;
using FlaUI.UIA3;

namespace ProtonMailBridge.UI.Tests.Tests
{
    [TestFixture]
    public class SettingsMenuTests : TestSession
    {
        private readonly LoginWindow _loginWindow = new();
        private readonly HomeWindow _mainWindow = new();
        private readonly HelpMenuResult _helpMenuResult = new();
        private readonly HelpMenuWindow _helpMenuWindow = new();
        private readonly HomeResult _homeResult = new();
        private readonly SettingsMenuWindow _settingsMenuWindow = new();
        private readonly SettingsMenuResults _settingsMenuResults = new();

        [SetUp]
        public void TestInitialize()
        {
            LaunchApp();
        }

        [Test]

        public void OpenSettingsMenuAndSwitchBackToAccountView()
        {
            _loginWindow.SignIn(TestUserData.GetPaidUser());
            _settingsMenuWindow.ClickSettingsButton();
            _settingsMenuWindow.ClickBackFromSettingsMenu();
            Thread.Sleep(2000);
            _homeResult.CheckIfLoggedIn();
        }

        [Test]
        public void VerifyAutomaticUpdateIsEnabledByDefault()
        {
            _loginWindow.SignIn(TestUserData.GetPaidUser());
            _settingsMenuWindow.ClickSettingsButton();
            _settingsMenuResults.AutomaticUpdatesIsEnabledByDefault();
            _settingsMenuWindow.ClickBackFromSettingsMenu();
        }

        [Test]
        public void VerifyDisableAndEnableAutomaticUpdates()
        {
            _loginWindow.SignIn(TestUserData.GetPaidUser());
            _settingsMenuWindow.ClickSettingsButton();
            _settingsMenuWindow.DisableAndEnableAutomaticUpdates();
            _settingsMenuWindow.ClickBackFromSettingsMenu();
        }

        [Test]
        public void VerifyOpenOnStartUpIsEnabledByDefault()
        {
            _loginWindow.SignIn(TestUserData.GetPaidUser());
            _settingsMenuWindow.ClickSettingsButton();
            _settingsMenuResults.OpenOnStartUpIsEnabledByDefault();
            _settingsMenuWindow.ClickBackFromSettingsMenu();
        }

        [Test]
        public void VerifyDisableAndEnableOpenOnStartUp()
        {
            _loginWindow.SignIn(TestUserData.GetPaidUser());
            _settingsMenuWindow.ClickSettingsButton();
            _settingsMenuWindow.DisableAndEnableOpenOnStartUp();
            _settingsMenuWindow.ClickBackFromSettingsMenu();
        }
        [Test]
        public void VerifyBetaAccessIsDisabledByDefault()
        {
            _loginWindow.SignIn(TestUserData.GetPaidUser());
            _settingsMenuWindow.ClickSettingsButton();
            _settingsMenuResults.BetaAccessIsDisabledByDefault();
            _settingsMenuWindow.ClickBackFromSettingsMenu();
        }

        [Test]
        public void VerifyEnableAndDisableBetaAccess()
        {
            _loginWindow.SignIn(TestUserData.GetPaidUser());
            _settingsMenuWindow.ClickSettingsButton();
            _settingsMenuWindow.EnableAndDisableBetaAccess();
            _settingsMenuWindow.ClickBackFromSettingsMenu();
        }

        [Test]
        public void VerifyExpandAndCollapseAdvancedSettings()
        {
            _loginWindow.SignIn(TestUserData.GetPaidUser());
            _settingsMenuWindow.ClickSettingsButton();
            _settingsMenuWindow.ExpandAdvancedSettings();
            _settingsMenuWindow.CollapseAdvancedSettings();
            _settingsMenuWindow.ClickBackFromSettingsMenu();
        }

        [Test]
        public void VerifyAlternativeRoutingIsDisabledByDefault()
        {
            _loginWindow.SignIn(TestUserData.GetPaidUser());
            _settingsMenuWindow.ClickSettingsButton();
            _settingsMenuWindow.ExpandAdvancedSettings();
            _settingsMenuResults.AlternativeRoutingIsDisabledByDefault();
            _settingsMenuWindow.CollapseAdvancedSettings();
            _settingsMenuWindow.ClickBackFromSettingsMenu();
        }

        [Test]
        public void VerifyEnableAndDisableAlternativeRouting()
        {
            _loginWindow.SignIn(TestUserData.GetPaidUser());
            _settingsMenuWindow.ClickSettingsButton();
            _settingsMenuWindow.ExpandAdvancedSettings();
            _settingsMenuWindow.EnableAndDisableAlternativeRouting();
            _settingsMenuWindow.CollapseAdvancedSettings();
            _settingsMenuWindow.ClickBackFromSettingsMenu();
        }

        [Test]
        public void VerifyDarkModeIsDisabledByDefault()
        {
            _loginWindow.SignIn(TestUserData.GetPaidUser());
            _settingsMenuWindow.ClickSettingsButton();
            _settingsMenuWindow.ExpandAdvancedSettings();
            _settingsMenuResults.DarkModeIsDisabledByDefault();
            _settingsMenuWindow.CollapseAdvancedSettings();
            _settingsMenuWindow.ClickBackFromSettingsMenu();
        }

        [Test]
        public void EnableAndDisableDarkMode()
        {
            _loginWindow.SignIn(TestUserData.GetPaidUser());
            _settingsMenuWindow.ClickSettingsButton();
            _settingsMenuWindow.ExpandAdvancedSettings();
            _settingsMenuWindow.CheckEnableAndDisableDarkMode();
            _settingsMenuWindow.CollapseAdvancedSettings();
            _settingsMenuWindow.ClickBackFromSettingsMenu();
        }
        [Test]
        public void VerifyShowAllMailIsEnabledByDefault()
        {
            _loginWindow.SignIn(TestUserData.GetPaidUser());
            _settingsMenuWindow.ClickSettingsButton();
            _settingsMenuWindow.ExpandAdvancedSettings();
            _settingsMenuResults.ShowAllMailIsEnabledByDefault();
            _settingsMenuWindow.CollapseAdvancedSettings();
            _settingsMenuWindow.ClickBackFromSettingsMenu();
        }

        [Test]
        public void VerifyDisableAndEnableShowAllMail()
        {
            _loginWindow.SignIn(TestUserData.GetPaidUser());
            _settingsMenuWindow.ClickSettingsButton();
            _settingsMenuWindow.ExpandAdvancedSettings();
            _settingsMenuWindow.DisableAndEnableShowAllMail();
            _settingsMenuWindow.CollapseAdvancedSettings();
            _settingsMenuWindow.ClickBackFromSettingsMenu();
        }
        [Test]
        public void VerifyCollectUsageDiagnosticsIsEnabledByDefault()
        {
            _loginWindow.SignIn(TestUserData.GetPaidUser());
            _settingsMenuWindow.ClickSettingsButton();
            _settingsMenuWindow.ExpandAdvancedSettings();
            Mouse.Scroll(-20);
            //Thread.Sleep(3000);
            _settingsMenuResults.CollectUsageDiagnosticsIsEnabledByDefault();
            Mouse.Scroll(20);
            _settingsMenuWindow.CollapseAdvancedSettings();
            _settingsMenuWindow.ClickBackFromSettingsMenu();
        }

        [Test]
        public void VerifyDisableAndEnableCollectUsageDiagnostics()
        {
            _loginWindow.SignIn(TestUserData.GetPaidUser());
            _settingsMenuWindow.ClickSettingsButton();
            _settingsMenuWindow.ExpandAdvancedSettings();
            Mouse.Scroll(-20);
            _settingsMenuWindow.DisableAndEnableCollectUsageDiagnostics();
            Mouse.Scroll(20);
            _settingsMenuWindow.CollapseAdvancedSettings();
            _settingsMenuWindow.ClickBackFromSettingsMenu();
        }

        [Test]
        public void VerifyDefaultImapSmtpPorts()
        {
            _loginWindow.SignIn(TestUserData.GetPaidUser());
            _settingsMenuWindow.ClickSettingsButton();
            _settingsMenuWindow.ExpandAdvancedSettings();
            Mouse.Scroll(-20);
            Thread.Sleep(2000);
            _settingsMenuWindow.OpenChangeDefaultPorts();
            Thread.Sleep(2000);
            _settingsMenuResults.VerifyDefaultPorts();
            _settingsMenuWindow.CancelChangingDefaultPorts();
            _settingsMenuWindow.ClickBackFromSettingsMenu();
        }

        [Test]
        public void ChangeAndSwitchToDefaultIMAPandSMTPports()
        {
            _loginWindow.SignIn(TestUserData.GetPaidUser());
            _settingsMenuWindow.ClickSettingsButton();
            _settingsMenuWindow.ExpandAdvancedSettings();
            Mouse.Scroll(-20);
            Thread.Sleep(5000);
            _settingsMenuWindow.ChangeDefaultPorts();
            _settingsMenuWindow.SwitchBackToDefaultPorts();
            _settingsMenuWindow.ClickBackFromSettingsMenu();
        }

        [Test]
        public void VerifyDefaultConnectionMode()
        {
            _loginWindow.SignIn(TestUserData.GetPaidUser());
            _settingsMenuWindow.ClickSettingsButton();
            _settingsMenuWindow.ExpandAdvancedSettings();
            Mouse.Scroll(-20);
            Thread.Sleep(5000);
            _settingsMenuWindow.OpenChangeConnectionMode();
            _settingsMenuResults.VerifyDefaultConnectionMode();
            _settingsMenuWindow.CancelChangeConnectionMode();
            _settingsMenuWindow.ClickBackFromSettingsMenu();
        }

        [Test]
        public void ChangeConnectionModeAndSwitchToDefault()
        {
            _loginWindow.SignIn(TestUserData.GetPaidUser());
            _settingsMenuWindow.ClickSettingsButton();
            _settingsMenuWindow.ExpandAdvancedSettings();
            Mouse.Scroll(-20);
            Thread.Sleep(5000);
            _settingsMenuWindow.OpenChangeConnectionMode();
            _settingsMenuWindow.ChangeConnectionMode();
            _settingsMenuWindow.OpenChangeConnectionMode();
            _settingsMenuResults.AssertTheChangedConnectionMode();
            _settingsMenuWindow.CancelChangeConnectionMode();
            _settingsMenuWindow.OpenChangeConnectionMode();
            _settingsMenuWindow.SwitchBackToDefaultConnectionMode();
            _settingsMenuWindow.ClickBackFromSettingsMenu();
        }

        [Test]
        public void OpenConfigureLocalCache()
        {
            _loginWindow.SignIn(TestUserData.GetPaidUser());
            _settingsMenuWindow.ClickSettingsButton();
            _settingsMenuWindow.ExpandAdvancedSettings();
            Mouse.Scroll(-20);
            Thread.Sleep(2000);
            _settingsMenuWindow.ConfigureLocalCache();
            Thread.Sleep(2000);
            _settingsMenuResults.DefaultCacheLocation();
            _settingsMenuWindow.CancelToConfigureLocalCache();
            _settingsMenuWindow.ClickBackFromSettingsMenu();
        }

        [Test]
        public void ChangeLocationSwitchBackToDefaultAndDeleteOldLocalCacheLocation()
        {
            _loginWindow.SignIn(TestUserData.GetPaidUser());
            _settingsMenuWindow.ClickSettingsButton();
            _settingsMenuWindow.ExpandAdvancedSettings();
            Mouse.Scroll(-20);
            Thread.Sleep(2000);
            _settingsMenuWindow.ConfigureLocalCache();
            Thread.Sleep(2000);
            _settingsMenuWindow.ChangeAndSwitchBackLocalCacheLocation();
            _settingsMenuWindow.ClickBackFromSettingsMenu();
        }

        [Test]
        public void ExportTlsCertificatesVerifyExportAndDeleteTheExportFolder()
        {
            _loginWindow.SignIn(TestUserData.GetPaidUser());
            _settingsMenuWindow.ClickSettingsButton();
            _settingsMenuWindow.ExpandAdvancedSettings();
            Mouse.Scroll(-20);
            Thread.Sleep(2000);
            _settingsMenuWindow.ExportAssertDeleteTLSCertificates();
            Thread.Sleep(2000);
            _settingsMenuWindow.ClickBackFromSettingsMenu();
        }
        [Test]
        public void RepairBridge()
        {
            _loginWindow.SignIn(TestUserData.GetPaidUser());
            _settingsMenuWindow.ClickSettingsButton();
            _settingsMenuWindow.ExpandAdvancedSettings();
            Mouse.Scroll(-20);
            Thread.Sleep(2000);
            _settingsMenuWindow.VerifyRepairRestartsSync();
            Thread.Sleep(2000);
            _settingsMenuWindow.ClickBackFromSettingsMenu();
        }

        [Test]
        public void ResetBridge()
        {
            _loginWindow.SignIn(TestUserData.GetPaidUser());
            _settingsMenuWindow.ClickSettingsButton();
            _settingsMenuWindow.ExpandAdvancedSettings();
            Mouse.Scroll(-20);
            Thread.Sleep(2000);
            _settingsMenuWindow.VerifyResetAndRestartBridge();
            Thread.Sleep(2000);
            _loginWindow.SignIn(TestUserData.GetPaidUser());
        }

        [TearDown]
        public void TestCleanup()
        {
            _mainWindow.RemoveAccount();
            ClientCleanup();
        }
    }
}
