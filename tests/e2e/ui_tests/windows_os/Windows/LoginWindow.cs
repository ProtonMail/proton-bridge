using FlaUI.Core.AutomationElements;
using FlaUI.Core.Input;
using FlaUI.Core.Definitions;
using ProtonMailBridge.UI.Tests.TestsHelper;

namespace ProtonMailBridge.UI.Tests.Windows
{
    public class LoginWindow : UIActions
    {
        private AutomationElement[] InputFields => Window.FindAllDescendants(cf => cf.ByControlType(ControlType.Edit));
        private TextBox UsernameInput => InputFields[0].AsTextBox();
        private TextBox PasswordInput => InputFields[1].AsTextBox();
        private Button SignInButton => Window.FindFirstDescendant(cf => cf.ByControlType(ControlType.Button).And(cf.ByName("Sign in"))).AsButton();
        private Button StartSetupButton => Window.FindFirstDescendant(cf => cf.ByName("Start setup")).AsButton();
        private Button SetUpLater => Window.FindFirstDescendant(cf => cf.ByName("Setup later")).AsButton();

        public LoginWindow SignIn(TestUserData user)
        {
            ClickStartSetupButton();
            EnterCredentials(user);
            Wait.UntilInputIsProcessed(TestData.TenSecondsTimeout);
            SetUpLater?.Click();

            return this;
        }

        public LoginWindow SignIn(string username, string password)
        {
            TestUserData user = new TestUserData(username, password);
            SignIn(user);
            return this;
        }

        public LoginWindow ClickStartSetupButton()
        {
            StartSetupButton?.Click();

            return this;
        }

        public LoginWindow EnterCredentials(TestUserData user)
        {
            UsernameInput.Text = user.Username;
            PasswordInput.Text = user.Password;
            SignInButton.Click();
            return this;
        }
    }
}