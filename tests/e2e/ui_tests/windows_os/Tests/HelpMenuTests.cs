using NUnit.Framework;
using ProtonMailBridge.UI.Tests.TestsHelper;
using ProtonMailBridge.UI.Tests.Windows;
using ProtonMailBridge.UI.Tests.Results;
using FlaUI.Core.Input;
using FlaUI.Core.AutomationElements;
using FlaUI.UIA3;

namespace ProtonMailBridge.UI.Tests.Tests
{
    [TestFixture]
    public class HelpMenuTests : TestSession
    {
        private readonly LoginWindow _loginWindow = new();
        private readonly HomeWindow _mainWindow = new();
        private readonly HelpMenuResult _helpMenuResult = new();
        private readonly HelpMenuWindow _helpMenuWindow = new();
        private readonly HomeResult _homeResult = new();

        [SetUp]
        public void TestInitialize()
        {
            LaunchApp();
        }

        [Test]
        public void OpenHelpMenuAndSwitchBackToAccountView()
        {
            _loginWindow.SignIn(TestUserData.GetPaidUser());
            _helpMenuWindow.ClickHelpButton();
            _helpMenuWindow.ClickBackFromHelpMenu();
            Thread.Sleep(2000);
            _homeResult.CheckIfLoggedIn();
        }

        [Test]
        public void OpenGoToHelpTopics()
        {
            _loginWindow.SignIn(TestUserData.GetPaidUser());
            _helpMenuWindow.ClickHelpButton();
            _helpMenuWindow.ClickGoToHelpTopics();
            Wait.UntilInputIsProcessed(TimeSpan.FromSeconds(3));
            switchToChromeWindow();
            _helpMenuResult.CheckHelpLinkIsOpen();
            Window.Focus();
            _helpMenuWindow.ClickBackFromHelpMenu();
        }

        [Test]
        public void CheckForUpdates()
        {
            _loginWindow.SignIn(TestUserData.GetPaidUser());
            _helpMenuWindow.ClickHelpButton();
            _helpMenuWindow.ClickCheckNowButton();
            Wait.UntilInputIsProcessed(TimeSpan.FromSeconds(3));
            _helpMenuResult.CheckBridgeIsUpToDateNotification();
            _helpMenuWindow.ConfirmNotification();
            Wait.UntilInputIsProcessed(TimeSpan.FromSeconds(1));
            _helpMenuWindow.ClickBackFromHelpMenu();
        }
        [Test]
        public void OpenLogs()
        {
            _loginWindow.SignIn(TestUserData.GetPaidUser());
            _helpMenuWindow.ClickHelpButton();
            _helpMenuWindow.ClickLogsButton();
            Wait.UntilInputIsProcessed(TimeSpan.FromSeconds(3));
            switchToFileExplorerWindow();
            _helpMenuResult.CheckBridgeLogsAreOpen();
            Window.Focus();
            _helpMenuWindow.ClickBackFromHelpMenu();
        }

        [Test]
        public void OpenMissingEmailsReportProblem()
        {
            _loginWindow.SignIn(TestUserData.GetPaidUser());
            _helpMenuWindow.ClickHelpButton();
            _helpMenuWindow.ClickReportProblemButton();
            _helpMenuWindow.ClickICannotFindEmailsInEmailClient();
            _helpMenuWindow.EnterMissingEmailsProblemDetails();
            _helpMenuResult.CheckIfProblemIsSuccReported();
            _helpMenuWindow.ConfirmNotification();
        }

        [Test]
        public void OpenNotAbleToSendEmailsReportProblem()
        {
            _loginWindow.SignIn(TestUserData.GetPaidUser());
            _helpMenuWindow.ClickHelpButton();
            _helpMenuWindow.ClickReportProblemButton();
            _helpMenuWindow.ClickNotAbleToSendEmails();
            _helpMenuWindow.EnterNotAbleToSendEmailProblemDetails();
            _helpMenuResult.CheckIfProblemIsSuccReported();
            _helpMenuWindow.ConfirmNotification();
        }

        [Test]
        public void OpenBridgeIsNotStartingCorrectlyReportProblem()
        {
            _loginWindow.SignIn(TestUserData.GetPaidUser());
            _helpMenuWindow.ClickHelpButton();
            _helpMenuWindow.ClickReportProblemButton();
            _helpMenuWindow.ClickBridgeIsNotStartingCorrectly();
            _helpMenuWindow.EnterBridgeIsNotStartingCorrectlyProblemDetails();
            _helpMenuResult.CheckIfProblemIsSuccReported();
            _helpMenuWindow.ConfirmNotification();
        }

        [Test]
        public void OpenBridgeIsRunningSlowReportProblem()
        {
            _loginWindow.SignIn(TestUserData.GetPaidUser());
            _helpMenuWindow.ClickHelpButton();
            _helpMenuWindow.ClickReportProblemButton();
            _helpMenuWindow.ClickBridgeIsRunningSlow();
            _helpMenuWindow.EnterBridgeIsRunningSlowProblemDetails();
            _helpMenuResult.CheckIfProblemIsSuccReported();
            _helpMenuWindow.ConfirmNotification();

        }
        [Test]
        public void OpenSomethingElseReportProblem()
        {
            _loginWindow.SignIn(TestUserData.GetPaidUser());
            _helpMenuWindow.ClickHelpButton();
            _helpMenuWindow.ClickReportProblemButton();
            _helpMenuWindow.ClickSomethingElse();
            _helpMenuWindow.EnterSomethingElseProblemDetails();
            _helpMenuResult.CheckIfProblemIsSuccReported();
            _helpMenuWindow.ConfirmNotification();
        }

        [TearDown]
        public void TestCleanup()
        {
            _mainWindow.RemoveAccount();
            ClientCleanup();
        }
    }
}