using FlaUI.Core.AutomationElements;
using FlaUI.Core.Definitions;
using System;


namespace ProtonMailBridge.UI.Tests.Windows
{
    public class HomeWindow : UIActions
    {
        private AutomationElement[] AccountViewButtons => AccountView.FindAllChildren(cf => cf.ByControlType(ControlType.Button));
        private Button RemoveAccountButton => AccountViewButtons[1].AsButton();
        private AutomationElement RemoveAccountConfirmModal => Window.FindFirstDescendant(cf => cf.ByControlType(ControlType.Window));
        private Button ConfirmRemoveAccountButton => RemoveAccountConfirmModal.FindFirstDescendant(cf => cf.ByControlType(ControlType.Button).And(cf.ByName("Remove this account"))).AsButton();
        private Button SignOutButton => AccountView.FindFirstDescendant(cf => cf.ByControlType(ControlType.Button).And(cf.ByName("Sign out"))).AsButton();
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
        public HomeWindow SignOutAccount()
        {
            SignOutButton.Click();
            return this;
        }
    }
}
