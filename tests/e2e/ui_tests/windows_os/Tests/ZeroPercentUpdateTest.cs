using System;
using System.Collections.Generic;
using System.Diagnostics;
using System.Linq;
using System.Text;
using System.Text.Encodings.Web;
using System.Text.Json;
using System.Text.Json.Nodes;
using System.Threading.Tasks;
using ProtonMailBridge.UI.Tests.Results;
using ProtonMailBridge.UI.Tests.TestsHelper;
using ProtonMailBridge.UI.Tests.Windows;

namespace ProtonMailBridge.UI.Tests.Tests
{
    public class ZeroPercentUpdateTest : TestSession
    {
        private readonly LoginWindow _loginWindow = new();
        private readonly SettingsMenuWindow _settingsMenuWindow = new();
        private readonly HelpMenuWindow _helpMenuWindow = new();
        private readonly ZeroPercentUpdateWindow _zeroPercentWindow = new();

        [SetUp]
        public void TestInitialize()
        {
            LaunchApp();
        }

        [Test]
        [Category("ZeroPercentUpdateRollout")]
        public void EnableBetaAccessVerifyBetaIsEnabledVerifyNotificationAndRestartBridge()
        {
            _zeroPercentWindow.ClickStartSetupButton();
            _zeroPercentWindow.CLickCancelButton();
            _helpMenuWindow.ClickHelpButton();
            _zeroPercentWindow.SaveCurrentVersionAndTagNumber();
            Thread.Sleep(2000);
            ClientCleanup();
            _zeroPercentWindow.editTheVault();
            LaunchApp();
            _loginWindow.SignIn(TestUserData.GetPaidUser());
            _settingsMenuWindow.ClickSettingsButton();
            _zeroPercentWindow.VerifyBetaAccessIsEnabled();
            _zeroPercentWindow.RestartBridgeNotification();
            _helpMenuWindow.ClickHelpButton();
            Thread.Sleep(2000);
            _zeroPercentWindow.VerifyVersionAndTagNumberOnRelaunch();
        }


        [TearDown]
        public void TestCleanup()
        {
            ClientCleanup();
        }
    }
}
