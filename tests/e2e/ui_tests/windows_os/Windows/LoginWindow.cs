using FlaUI.Core.AutomationElements;
using FlaUI.Core.Input;
using FlaUI.Core.Definitions;
using ProtonMailBridge.UI.Tests.TestsHelper;
using ProtonMailBridge.UI.Tests.Results;
using System.Diagnostics;

namespace ProtonMailBridge.UI.Tests.Windows
{
    public class LoginWindow : UIActions
    {
        private AutomationElement[] InputFields => Window.FindAllDescendants(cf => cf.ByControlType(ControlType.Edit));
        private TextBox UsernameInput => InputFields[0].AsTextBox();
        private TextBox PasswordInput => InputFields[1].AsTextBox();
        private Button SignInButton => Window.FindFirstDescendant(cf => cf.ByControlType(ControlType.Button).And(cf.ByName("Sign in"))).AsButton();
        private Button SigningInButton => Window.FindFirstDescendant(cf => cf.ByControlType(ControlType.Button).And(cf.ByName("Signing in"))).AsButton();
        private Button StartSetupButton => Window.FindFirstDescendant(cf => cf.ByName("Start setup")).AsButton();
        private Button SetUpLater => Window.FindFirstDescendant(cf => cf.ByName("Setup later")).AsButton();
        private TextBox MailboxPasswordInput => Window.FindFirstDescendant(cf => cf.ByControlType(ControlType.Edit)).AsTextBox();
        private Button UnlockButton => Window.FindFirstDescendant(cf => cf.ByControlType(ControlType.Button).And(cf.ByName("Unlock"))).AsButton();
        private Button CancelSignIn => Window.FindFirstDescendant(cf => cf.ByControlType(ControlType.Button).And(cf.ByName("Cancel"))).AsButton();

        public LoginWindow SignIn(TestUserData user)
        {
            ClickStartSetupButton();
            EnterCredentials(user);
            WaitForAuthorizationToComplete(60);

            SetUpLater?.Click();

            return this;
        }

        public LoginWindow SignInMailbox(TestUserData user)
        {
            ClickStartSetupButton();
            EnterCredentials(user);
            Wait.UntilInputIsProcessed(TestData.TenSecondsTimeout);
            EnterMailboxPassword(user);
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

        public LoginWindow EnterMailboxPassword(TestUserData user)
        {
            MailboxPasswordInput.Text = user.MailboxPassword;
            UnlockButton.Click();
            return this;
        }

        public LoginWindow ClickCancelToSignIn () 
        {            
            CancelSignIn.Click();
            return this;
        }

        private void WaitForAuthorizationToComplete(int numOfSeconds)
        {
            TimeSpan timeout = TimeSpan.FromSeconds(numOfSeconds);
            Stopwatch stopwatch = Stopwatch.StartNew();

            
            while (stopwatch.Elapsed < timeout)
            {
                //if Signing in button is not visible authorization process is finished
                if (SigningInButton == null)
                {
                    return;
                }

                Wait.UntilInputIsProcessed();
                Thread.Sleep(500);
            }

        }
    }
}