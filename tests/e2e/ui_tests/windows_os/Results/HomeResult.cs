using FlaUI.Core.AutomationElements;
using FlaUI.Core.Definitions;
using ProtonMailBridge.UI.Tests.TestsHelper;
using FlaUI.Core.Input;
using System.DirectoryServices;

namespace ProtonMailBridge.UI.Tests.Results
{
    public class HomeResult : UIActions
    {
        private Button SignOutButton => AccountView.FindFirstDescendant(cf => cf.ByControlType(ControlType.Button).And(cf.ByName("Sign out"))).AsButton();
        private AutomationElement NotificationWindow => Window.FindFirstDescendant(cf => cf.ByControlType(ControlType.Window));
        private TextBox FreeAccountErrorText => NotificationWindow.FindFirstDescendant(cf => cf.ByControlType(ControlType.Text)).AsTextBox();
        private TextBox SignedOutAccount => AccountView.FindFirstDescendant(cf => cf.ByControlType(ControlType.Text)).AsTextBox();
        private TextBox AlreadySignedInText => NotificationWindow.FindFirstDescendant(cf => cf.ByControlType(ControlType.Text)).AsTextBox();
        private Button OkToAcknowledgeAccountAlreadySignedIn => NotificationWindow.FindFirstDescendant(cf => cf.ByControlType(ControlType.Button).And(cf.ByName("OK"))).AsButton();
        private AutomationElement[] TextFields => Window.FindAllDescendants(cf => cf.ByControlType(ControlType.Text));
        private TextBox SynchronizingField => TextFields[4].AsTextBox();
        private TextBox AccountDisabledErrorText => Window.FindFirstDescendant(cf => cf.ByControlType(ControlType.Text).And(cf.ByName("failed to create new API client: 422 POST https://mail-api.proton.me/auth/v4: This account has been suspended due to a potential policy violation. If you believe this is in error, please contact us at https://proton.me/support/appeal-abuse (Code=10003, Status=422)"))).AsTextBox();
        private TextBox AccountDelinquentErrorText => Window.FindFirstDescendant(cf => cf.ByControlType(ControlType.Text).And(cf.ByName("failed to create new API client: 422 POST https://mail-api.proton.me/auth/v4: Use of this client requires permissions not available to your account (Code=2011, Status=422)"))).AsTextBox();
        private TextBox IncorrectLoginCredentialsErrorText => Window.FindFirstDescendant(cf => cf.ByControlType(ControlType.Text).And(cf.ByName("Incorrect login credentials"))).AsTextBox();
        private TextBox EnterEmailOrUsernameErrorText => Window.FindFirstDescendant(cf => cf.ByControlType(ControlType.Text).And(cf.ByName("Enter email or username"))).AsTextBox();
        private TextBox EnterPasswordErrorText => Window.FindFirstDescendant(cf => cf.ByControlType(ControlType.Text).And(cf.ByName("Enter password"))).AsTextBox();
        private TextBox ConnectedStateText => Window.FindFirstDescendant(cf => cf.ByControlType(ControlType.Text).And(cf.ByName("Connected"))).AsTextBox();
        
        public HomeResult CheckConnectedState()
        {
            Assert.That(ConnectedStateText.IsAvailable, Is.True);
            return this;
        }
        public HomeResult CheckIfLoggedIn()
        {
            Assert.That(SignOutButton.IsAvailable, Is.True);
            return this;
        }
        public HomeResult CheckIfSynchronizingBarIsShown()
        {
            Assert.That(SynchronizingField.IsAvailable && SynchronizingField.Name.StartsWith("Synchronizing"), Is.True);
            return this;
        }
        public HomeResult CheckIfFreeAccountErrorIsDisplayed(string ErrorText)
        {
            Assert.That(FreeAccountErrorText.Name == ErrorText, Is.True);
            return this;
        }
        public HomeResult CheckIfAccountIsSignedOut()
        {
            Assert.That(SignedOutAccount.IsAvailable, Is.True);
            return this;
        }
        public HomeResult CheckIfAccountAlreadySignedInIsDisplayed()
        {
            Assert.That(AlreadySignedInText.IsAvailable, Is.True);
            return this;
        }
        public HomeResult ClickOkToAcknowledgeAccountAlreadySignedIn ()
        {
            OkToAcknowledgeAccountAlreadySignedIn.Click();
            return this;
        }

        public HomeResult CheckIfIncorrectCredentialsErrorIsDisplayed()
        {
            Assert.That(IncorrectLoginCredentialsErrorText.IsAvailable, Is.True);
            return this;
        }

        public HomeResult CheckIfEnterUsernameAndEnterPasswordErrorMsgsAreDisplayed()
        { 
            Assert.That(EnterEmailOrUsernameErrorText.IsAvailable && EnterPasswordErrorText.IsAvailable, Is.True);
            return this;
        }

        public HomeResult CheckIfDsabledAccountErrorIsDisplayed()
        {
            Assert.That(AccountDisabledErrorText.IsAvailable, Is.True);
            return this;
        }

        public HomeResult CheckIfDelinquentAccountErrorIsDisplayed()
        {
            Assert.That(AccountDelinquentErrorText.IsAvailable, Is.True);
            return this;
    
        }

        public HomeResult CheckIfNotificationTextIsShown()
        {
            Assert.That(AlreadySignedInText.IsAvailable, Is.True);
            return this;
        }
    }
}