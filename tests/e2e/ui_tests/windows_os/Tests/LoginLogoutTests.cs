using NUnit.Framework;
using ProtonMailBridge.UI.Tests.TestsHelper;
using ProtonMailBridge.UI.Tests.Windows;
using ProtonMailBridge.UI.Tests.Results;

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
        public void LoginAsPaidUser()
        {
            _loginWindow.SignIn(TestUserData.GetPaidUser());
            _homeResult.CheckIfLoggedIn();
        }

        [Test]
        public void LoginAsFreeUser()
        {
            _loginWindow.SignIn(TestUserData.GetFreeUser());
            _homeResult.CheckIfFreeAccountErrorIsDisplayed(FreeAccountErrorText);
        }

        [Test]
        public void SuccessfullLogout()
        {
            _loginWindow.SignIn(TestUserData.GetPaidUser());
            _mainWindow.SignOutAccount();
            _homeResult.CheckIfAccountIsSignedOut();
        }

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
