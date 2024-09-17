using FlaUI.Core.AutomationElements;
using FlaUI.Core.Definitions;

namespace ProtonMailBridge.UI.Tests.Results
{
    public class HomeResult : UIActions
    {
        private Button SignOutButton => AccountView.FindFirstDescendant(cf => cf.ByControlType(ControlType.Button).And(cf.ByName("Sign out"))).AsButton();
        private AutomationElement NotificationWindow => Window.FindFirstDescendant(cf => cf.ByControlType(ControlType.Window));
        private TextBox FreeAccountErrorText => NotificationWindow.FindFirstDescendant(cf => cf.ByControlType(ControlType.Text)).AsTextBox();
        private TextBox SignedOutAccount => AccountView.FindFirstDescendant(cf => cf.ByControlType(ControlType.Text)).AsTextBox();
        public HomeResult CheckIfLoggedIn()
        {
            Assert.That(SignOutButton.IsAvailable, Is.True);
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
    }
}