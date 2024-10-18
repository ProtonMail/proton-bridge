using FlaUI.Core.AutomationElements;
using FlaUI.Core.Definitions;
using FlaUI.Core.Conditions;
using FlaUI.Core.Input;
using ProtonMailBridge.UI.Tests.TestsHelper;
using System;


namespace ProtonMailBridge.UI.Tests.Windows
{
    public class HomeWindow : UIActions
    {
        private AutomationElement[] AccountViewButtons => AccountView.FindAllChildren(cf => cf.ByControlType(ControlType.Button));
        private AutomationElement[] HomeButtons => Window.FindAllDescendants(cf => cf.ByControlType(ControlType.Button));
        private Button AddNewAccountButton => HomeButtons[6].AsButton();
        private Button RemoveAccountButton => AccountViewButtons[1].AsButton();
        private AutomationElement RemoveAccountConfirmModal => Window.FindFirstDescendant(cf => cf.ByControlType(ControlType.Window));
        private Button ConfirmRemoveAccountButton => RemoveAccountConfirmModal.FindFirstDescendant(cf => cf.ByControlType(ControlType.Button).And(cf.ByName("Remove this account"))).AsButton();
        private Button SignOutButton => AccountView.FindFirstDescendant(cf => cf.ByControlType(ControlType.Button).And(cf.ByName("Sign out"))).AsButton();
        private Button SignInButton => AccountView.FindFirstDescendant(cf => cf.ByControlType(ControlType.Button).And(cf.ByName("Sign in"))).AsButton();

        public HomeWindow RemoveAccount()
        {
            try
            {
                RemoveAccountButton.Click();
                ConfirmRemoveAccountButton.Click();
            }
            catch (System.NullReferenceException)
            {
                ClientCleanup();
            }
            return this;
        }

        public HomeWindow AddNewAccount ()
        {
            AddNewAccountButton.Click();
            return this;
        }
        
        public HomeWindow SignOutAccount()
        {
            SignOutButton.Click();
            return this;
        }
        public HomeWindow ClickSignInMainWindow()
        {
            SignInButton.Click();
            return this;
        }

    }
}
