using NUnit.Framework;
using ProtonMailBridge.UI.Tests.TestsHelper;
using ProtonMailBridge.UI.Tests.Windows;
using ProtonMailBridge.UI.Tests.Results;
using FlaUI.Core.Input;

namespace ProtonMailBridge.UI.Tests.Tests
{
    [TestFixture]
    public class LoginLogoutTests : TestSession
    {
        private readonly LoginWindow _loginWindow = new();
        private readonly HomeWindow _mainWindow = new();
        private readonly HomeResult _homeResult = new();
        private readonly string FreeAccountErrorText = "Bridge is exclusive to our mail paid plans. Upgrade your account to use Bridge.";

        [Test]
        public void LoginAsFreeUser()
        {
            _loginWindow.SignIn(TestUserData.GetFreeUser());
            _homeResult.CheckIfFreeAccountErrorIsDisplayed(FreeAccountErrorText);
        }

        [Test]
        public void LoginAsPaidUser()
        {
            _loginWindow.SignIn(TestUserData.GetPaidUser());
            _homeResult.CheckIfLoggedIn();
        }

        [Test]
        public void VerifyConnectedState()
        {
            _loginWindow.SignIn(TestUserData.GetPaidUser());
            _homeResult.CheckIfLoggedIn();
            _homeResult.CheckConnectedState();
        }

        [Test]
        public void VerifyAccountSynchronizingBar()
        {
            _loginWindow.SignIn(TestUserData.GetPaidUser());
            _homeResult.CheckIfSynchronizingBarIsShown();
        }

        [Test]
        public void AddAliasAddress()
        {
            _loginWindow.SignIn(TestUserData.GetPaidUser());
            _homeResult.CheckIfLoggedIn();
            _mainWindow.AddNewAccount();
            _loginWindow.SignIn(TestUserData.GetAliasUser());
            _homeResult.CheckIfAccountAlreadySignedInIsDisplayed();
            _homeResult.ClickOkToAcknowledgeAccountAlreadySignedIn();
            _loginWindow.ClickCancelToSignIn();
        }

        [Test]
        public void LoginWithMailboxPassword()
        {
            _loginWindow.SignInMailbox(TestUserData.GetMailboxUser());
            _homeResult.CheckIfLoggedIn();
            _mainWindow.SignOutAccount();
            _homeResult.CheckIfAccountIsSignedOut();
        }

        [Test]
        public void AddSameAccountTwice()
        {
            _loginWindow.SignIn(TestUserData.GetPaidUser());
            _homeResult.CheckIfLoggedIn();
            _mainWindow.AddNewAccount();
            _loginWindow.SignIn(TestUserData.GetPaidUser());
            _homeResult.CheckIfAccountAlreadySignedInIsDisplayed();
            _homeResult.ClickOkToAcknowledgeAccountAlreadySignedIn();
            _loginWindow.ClickCancelToSignIn();
            _homeResult.CheckIfLoggedIn();
        }

        [Test]
        public void AddAccountWithWrongCredentials()
        {
            _loginWindow.SignIn(TestUserData.GetIncorrectCredentialsUser());
            _homeResult.CheckIfIncorrectCredentialsErrorIsDisplayed();
            _loginWindow.ClickCancelToSignIn();
        }

        [Test, Order (1)]
        public void AddAccountWithEmptyCredentials()
        {
            _loginWindow.SignIn(TestUserData.GetEmptyCredentialsUser());
            _homeResult.CheckIfEnterUsernameAndEnterPasswordErrorMsgsAreDisplayed();
            _loginWindow.ClickCancelToSignIn();
            _loginWindow.SignIn(TestUserData.GetPaidUser());
            _homeResult.CheckIfLoggedIn();
        }

        [Test]
        public void AddSameAccountAfterBeingSignedOut()
        {
            _loginWindow.SignIn(TestUserData.GetPaidUser());
            _homeResult.CheckIfLoggedIn();
            _mainWindow.SignOutAccount();
            Wait.UntilInputIsProcessed(TimeSpan.FromSeconds(3));
            _mainWindow.ClickSignInMainWindow();
            _loginWindow.SignIn(TestUserData.GetPaidUser());
            _homeResult.CheckIfLoggedIn();
            _mainWindow.SignOutAccount();
        }

        /*
        [Test]
        public void AddSecondAccount()
        {
            _loginWindow.SignIn(TestUserData.GetPaidUser());
            _homeResult.CheckIfLoggedIn();
            _mainWindow.AddNewAccount();
            _loginWindow.SignInMailbox(TestUserData.GetMailboxUser());
            _homeResult.CheckIfLoggedIn();
        }
        */

        [Test]
        public void AddDisabledAccount()
        {
            _loginWindow.SignIn(TestUserData.GetDisabledUser());
            _homeResult.CheckIfDsabledAccountErrorIsDisplayed();
            _loginWindow.ClickCancelToSignIn();
        }

        [Test]
        public void AddDeliquentAccount()
        {
            _loginWindow.SignIn(TestUserData.GetDeliquentUser());
            _homeResult.CheckIfDelinquentAccountErrorIsDisplayed();
            _loginWindow.ClickCancelToSignIn();
        }

        //[Test]
        //public void SuccessfullLogout()
        //{
        //    _loginWindow.SignIn(TestUserData.GetPaidUser());
        //    _mainWindow.SignOutAccount();
        //    _homeResult.CheckIfAccountIsSignedOut();
        //}

        [SetUp]
        public void TestInitialize()
        {
            LaunchApp();
        }
        
        [TearDown]
        public void TestCleanup()
        {
            _mainWindow.RemoveAccount();
            ClientCleanup();
        }
    }
}