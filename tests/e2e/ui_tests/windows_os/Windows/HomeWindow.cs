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
        private CheckBox SplitAddressesToggle => AccountView.FindFirstDescendant(cf => cf.ByControlType(ControlType.CheckBox).And(cf.ByName("Split addresses toggle"))).AsCheckBox();
        private Button EnableSplitAddressButton => Window.FindFirstDescendant(cf => cf.ByControlType(ControlType.Button).And(cf.ByName("Enable split mode"))).AsButton();

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

        public HomeWindow AddNewAccount()
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
        public HomeWindow EnableSplitAddress()
        {
            SplitAddressesToggle.Click();
            EnableSplitAddressButton.Click();
            Thread.Sleep(5000);
            bool syncRestarted = WaitForCondition(() =>
            {
                return IsStatusLabelSyncing(Window);
            }, TimeSpan.FromSeconds(30));

            Assert.That(syncRestarted, Is.True, "Sync did not restart after Split Address mode was enabled.");
            Assert.That(SplitAddressesToggle.IsToggled, Is.True);
            return this;
        }

        public HomeWindow DisableSplitAddress()
        {
            SplitAddressesToggle.Click();
            Thread.Sleep(5000);
            bool syncRestarted = WaitForCondition(() =>
            {
                return IsStatusLabelSyncing(Window);
            }, TimeSpan.FromSeconds(30));

            Assert.That(syncRestarted, Is.True, "Sync did not restart after Split Address mode was disabled.");
            Assert.That(SplitAddressesToggle.IsToggled, Is.False);
            return this;
        }

        private bool IsStatusLabelSyncing(AutomationElement window)
        {
            var syncStatusElement = window.FindAllDescendants(cf => cf.ByControlType(ControlType.Text)).FirstOrDefault(el =>
            {
                string name = el.Name;
                return !string.IsNullOrEmpty(name) &&
                       name.StartsWith("Synchronizing (") &&
                       name.EndsWith("%)");
            });
            return syncStatusElement != null && syncStatusElement.Name.Contains("Synchronizing");
        }

        private bool WaitForCondition(Func<bool> condition, TimeSpan timeout, int pollingIntervalMs = 500)
        {
            var endTime = DateTime.Now + timeout;
            while (DateTime.Now < endTime)
            {
                if (condition())
                    return true;
                Thread.Sleep(pollingIntervalMs);
            }
            return false;
        }
    }
}
