using System;
using System.Collections.Generic;
using System.Linq;
using System.Text;
using System.Threading.Tasks;
using System;
using System.Collections.Generic;
using System.Linq;
using System.Threading.Tasks;
using FlaUI.Core.AutomationElements;
using FlaUI.Core.AutomationElements.Scrolling;
using FlaUI.Core.Definitions;
using FlaUI.Core.Input;
using FlaUI.Core.WindowsAPI;
using Microsoft.VisualBasic.Devices;
using NUnit.Framework.Legacy;
using ProtonMailBridge.UI.Tests.Results;
using ProtonMailBridge.UI.Tests.TestsHelper;
using Keyboard = FlaUI.Core.Input.Keyboard;
using Mouse = FlaUI.Core.Input.Mouse;
using static System.Windows.Forms.VisualStyles.VisualStyleElement.Window;
//using System.Windows.Forms;
using CheckBox = FlaUI.Core.AutomationElements.CheckBox;
using FlaUI.Core.Tools;
using System.Diagnostics;
using System.Drawing;
using System.Text.RegularExpressions;
using FlaUI.UIA3;

namespace ProtonMailBridge.UI.Tests.Windows
{
    public class SettingsMenuWindow : UIActions
    {
        private static Random random = new Random();
        private const int MinPort = 49152;
        private const int MaxPort = 65535;

        private AutomationElement[] InputFields => Window.FindAllDescendants(cf => cf.ByControlType(ControlType.Edit));
        private AutomationElement[] HomeButtons => Window.FindAllDescendants(cf => cf.ByControlType(ControlType.Button));
        private AutomationElement NotificationWindow => Window.FindFirstDescendant(cf => cf.ByControlType(ControlType.Window));
        private Button EnableBetaAccessButtonInPopUp => NotificationWindow.FindFirstDescendant(cf => cf.ByControlType(ControlType.Button).And(cf.ByName("Enable"))).AsButton();
        private AutomationElement[] ReportProblemPane => Window.FindAllDescendants(cf => cf.ByControlType(ControlType.Pane));
        private Button SettingsButton => HomeButtons[4].AsButton();
        private Button BackToAccountViewButton => Window.FindFirstDescendant(cf => cf.ByControlType(ControlType.Button).And(cf.ByName("Back"))).AsButton();
        private CheckBox AutomaticUpdates => Window.FindFirstDescendant(cf => cf.ByControlType(ControlType.CheckBox).And(cf.ByName("Automatic updates toggle"))).AsCheckBox();
        private CheckBox OpenOnStartUp => Window.FindFirstDescendant(cf => cf.ByControlType(ControlType.CheckBox).And(cf.ByName("Open on startup toggle"))).AsCheckBox();
        private CheckBox BetaAccess => Window.FindFirstDescendant(cf => cf.ByControlType(ControlType.CheckBox).And(cf.ByName("Beta access toggle"))).AsCheckBox();
        private TextBox AdvancedSettings => Window.FindFirstDescendant(cf => cf.ByControlType(ControlType.Text).And(cf.ByName("Advanced settings"))).AsTextBox();
        private TextBox AlternativeRoutingText => Window.FindFirstDescendant(cf => cf.ByControlType(ControlType.Text).And(cf.ByName("Alternative routing"))).AsTextBox();
        private CheckBox AlternativeRouting => Window.FindFirstDescendant(cf => cf.ByControlType(ControlType.CheckBox).And(cf.ByName("Alternative routing toggle"))).AsCheckBox();
        private CheckBox DarkMode => Window.FindFirstDescendant(cf => cf.ByControlType(ControlType.CheckBox).And(cf.ByName("Dark mode toggle"))).AsCheckBox();
        private CheckBox ShowAllMail => Window.FindFirstDescendant(cf => cf.ByControlType(ControlType.CheckBox).And(cf.ByName("Show All Mail toggle"))).AsCheckBox();
        private Button HideAllMailFolderInPopUp => NotificationWindow.FindFirstDescendant(cf => cf.ByControlType(ControlType.Button).And(cf.ByName("Hide All Mail folder"))).AsButton();
        private Button ShowAllMailFolderInPopUp => NotificationWindow.FindFirstDescendant(cf => cf.ByControlType(ControlType.Button).And(cf.ByName("Show All Mail folder"))).AsButton();
        private CheckBox CollectUsageDiagnostics => Window.FindFirstDescendant(cf => cf.ByControlType(ControlType.CheckBox).And(cf.ByName("Collect usage diagnostics toggle"))).AsCheckBox();
        private Button ChangeDefaultPortsButton => Window.FindFirstDescendant(cf => cf.ByControlType(ControlType.Button).And(cf.ByName("Default ports button"))).AsButton();
        private TextBox ImapPort => Window.FindFirstDescendant(cf => cf.ByControlType(ControlType.Edit).And(cf.ByName("IMAP port edit"))).AsTextBox();
        private TextBox SmtpPort => Window.FindFirstDescendant(cf => cf.ByControlType(ControlType.Edit).And(cf.ByName("SMTP port edit"))).AsTextBox();
        private Button SaveChangedPorts => Window.FindFirstDescendant(cf => cf.ByControlType(ControlType.Button).And(cf.ByName("Save"))).AsButton();
        private Button CancelDefaultPorts => Window.FindFirstDescendant(cf => cf.ByControlType(ControlType.Button).And(cf.ByName("Cancel"))).AsButton();
        private Button ChangeConnectionModeButton => Window.FindFirstDescendant(cf => cf.ByControlType(ControlType.Button).And(cf.ByName("Connection mode button"))).AsButton();
        private AutomationElement[] RadioButtons => Window.FindAllDescendants(cf => cf.ByControlType(ControlType.RadioButton));
        private RadioButton ImapStarttlsMode => RadioButtons[1].AsRadioButton();
        private RadioButton SmtpStarttlsMode => RadioButtons[3].AsRadioButton();
        private Button CancelChangeConnectionModeButton => Window.FindFirstDescendant(cf => cf.ByControlType(ControlType.Button).And(cf.ByName("Cancel"))).AsButton();
        private RadioButton ImapSslMode => RadioButtons[0].AsRadioButton();
        private RadioButton SmtpSslMode => RadioButtons[2].AsRadioButton();
        private Button SaveChangedConnectionMode => Window.FindFirstDescendant(cf => cf.ByControlType(ControlType.Button).And(cf.ByName("Save"))).AsButton();
        private Button ConfigureLocalCacheButton => Window.FindFirstDescendant(cf => cf.ByControlType(ControlType.Button).And(cf.ByName("Local cache button"))).AsButton();
        private AutomationElement[] TextFields => Window.FindAllDescendants(cf => cf.ByControlType(ControlType.Text));
        private Button Cancel => Window.FindFirstDescendant(cf => cf.ByControlType(ControlType.Button).And(cf.ByName("Cancel"))).AsButton();
        private Button ChangeLocalCacheLocationButton => Window.FindFirstDescendant(cf => cf.ByControlType(ControlType.Button).And(cf.ByName("Current cache location button"))).AsButton();
        private TextBox CacheLocation => Window.FindFirstDescendant(cf => cf.ByControlType(ControlType.Group).And(cf.ByName("Current cache location"))).FindAllDescendants(cf => cf.ByControlType(ControlType.Text))[1].AsTextBox();
        private Button ClickNewFolder => Window.FindFirstDescendant(cf => cf.ByControlType(ControlType.Button).And(cf.ByName("New folder"))).AsButton();
        private TextBox NewCreatedFolderTextBox => Window.FindFirstDescendant(cf => cf.ByControlType(ControlType.Pane).And(cf.ByName("Shell Folder View"))).FindFirstDescendant(cf => cf.ByControlType(ControlType.ListItem).And(cf.ByName("New folder"))).FindFirstDescendant(cf => cf.ByControlType(ControlType.Edit)).AsTextBox();
        private Button SelectFolderButton => Window.FindFirstDescendant(cf => cf.ByControlType(ControlType.Button).And(cf.ByName("Select Folder"))).AsButton();
        private Button SaveChangedCacheFolderLocation => Window.FindFirstDescendant(cf => cf.ByControlType(ControlType.Button).And(cf.ByName("Save"))).AsButton();
        private TextBox CacheLocationIsChangedNotification => NotificationWindow.FindFirstDescendant(cf => cf.ByControlType(ControlType.Text).And(cf.ByName("Cache location successfully changed"))).AsTextBox();
        private Button OkCacheLocationChangedNotification => NotificationWindow.FindFirstDescendant(cf => cf.ByControlType(ControlType.Button).And(cf.ByName("OK"))).AsButton();
        private Button Back => Window.FindFirstDescendant(cf => cf.ByControlType(ControlType.Button).And(cf.ByName("Back"))).AsButton();
        private Button UpArrowToGoBackToPreviousFolder => Window.FindFirstDescendant(cf => cf.ByControlType(ControlType.Pane).And(cf.ByClassName("UpBand"))).FindFirstDescendant(cf => cf.ByControlType(ControlType.Button)).AsButton();
        private Window SelectCacheLocationWindow => Window.FindFirstDescendant(cf => cf.ByControlType(ControlType.Window).And(cf.ByName("Select cache location"))).AsWindow();
        private Button ExportTLSCertificatesButton => Window.FindFirstDescendant(cf => cf.ByControlType(ControlType.Button).And(cf.ByName("Export TLS certificates button"))).AsButton();
        private Window SelectDirectoryWindow => Window.FindFirstDescendant(CF => CF.ByControlType(ControlType.Window).And(CF.ByName("Select directory"))).AsWindow();
        private Button RepairBridgeButton => Window.FindFirstDescendant(cf => cf.ByControlType(ControlType.Button).And(cf.ByName("Repair Bridge button"))).AsButton();
        private Button RepairButtonInPopUp => NotificationWindow.FindFirstDescendant(cf => cf.ByControlType(ControlType.Button).And(cf.ByName("Repair"))).AsButton();
        private Button ResetButton => Window.FindFirstDescendant(cf => cf.ByControlType(ControlType.Button).And(cf.ByName("Reset Bridge button"))).AsButton();
        private Button ResetAndRestartButtonInPopUp => NotificationWindow.FindFirstDescendant(cf => cf.ByControlType(ControlType.Button).And(cf.ByName("Reset and restart"))).AsButton();
        private Button StartSetUpButton => Window.FindFirstDescendant(cf => cf.ByControlType(ControlType.Button).And(cf.ByName("Start setup"))).AsButton();
       
        public SettingsMenuWindow ClickSettingsButton()
        {
            SettingsButton.Click();
            return this;
        }

        public SettingsMenuWindow ClickBackFromSettingsMenu()
        {
            BackToAccountViewButton.Click();
            return this;
        }

        public SettingsMenuWindow DisableAndEnableAutomaticUpdates()
        {
            AutomaticUpdates.Click();
            Assert.That(AutomaticUpdates.IsToggled, Is.False);
            Thread.Sleep(1000);
            AutomaticUpdates.Click();
            Assert.That(AutomaticUpdates.IsToggled, Is.True);
            return this;
        }
        public SettingsMenuWindow DisableAndEnableOpenOnStartUp()
        {
            OpenOnStartUp.Click();
            Assert.That(OpenOnStartUp.IsToggled, Is.False);
            Thread.Sleep(1000);
            OpenOnStartUp.Click();
            Assert.That(OpenOnStartUp.IsToggled, Is.True);
            return this;
        }

        public SettingsMenuWindow EnableAndDisableBetaAccess()
        {
            BetaAccess.Click();
            EnableBetaAccessButtonInPopUp.Click();
            Thread.Sleep(1000);
            Assert.That(BetaAccess.IsToggled, Is.True);
            BetaAccess.Click();
            Assert.That(BetaAccess.IsToggled, Is.False);
            return this;
        }
        public SettingsMenuWindow ExpandAdvancedSettings()
        {
            AdvancedSettings.Click();
            Thread.Sleep(1000);
            Assert.That(AlternativeRouting != null && AlternativeRouting.IsAvailable, Is.True);
            return this;
        }

        public SettingsMenuWindow CollapseAdvancedSettings()
        {
            AdvancedSettings.Click();
            return this;
        }
        public SettingsMenuWindow EnableAndDisableAlternativeRouting()
        {
            AlternativeRouting.Click();
            Assert.That(AlternativeRouting.IsToggled, Is.True);
            Thread.Sleep(1000);
            AlternativeRouting.Click();
            Assert.That(AlternativeRouting?.IsToggled, Is.False);
            return this;
        }

        public SettingsMenuWindow CheckEnableAndDisableDarkMode()
        {
            DarkMode.Click();
            Assert.That(DarkMode.IsToggled, Is.True);
            Thread.Sleep(1000);
            DarkMode.Click();
            Assert.That(DarkMode.IsToggled, Is.False);
            return this;
        }
        public SettingsMenuWindow DisableAndEnableShowAllMail()
        {
            ShowAllMail.Click();
            HideAllMailFolderInPopUp.Click();
            Assert.That(ShowAllMail.IsToggled, Is.False);
            Thread.Sleep(1000);
            ShowAllMail.Click();
            Thread.Sleep(1000);
            ShowAllMailFolderInPopUp.Click();
            Assert.That(ShowAllMail?.IsToggled, Is.True);
            return this;
        }

        public SettingsMenuWindow DisableAndEnableCollectUsageDiagnostics()
        {
            CollectUsageDiagnostics.Click();
            Thread.Sleep(3000);
            Assert.That(CollectUsageDiagnostics.IsToggled, Is.False);
            Thread.Sleep(1000);
            CollectUsageDiagnostics.Click();
            Thread.Sleep(1000);
            Assert.That(CollectUsageDiagnostics?.IsToggled, Is.True);
            return this;
        }

        public SettingsMenuWindow OpenChangeDefaultPorts()
        {
            ChangeDefaultPortsButton.Click();
            return this;
        }

        public SettingsMenuWindow CancelChangingDefaultPorts()
        {
            CancelDefaultPorts.Click();
            return this;
        }
        private int GenerateUniqueRandomPort()
        {
            return random.Next(MinPort, MaxPort +1);
        }
        public SettingsMenuWindow ChangeDefaultPorts()
        {
            ChangeDefaultPortsButton.Click();
            Thread.Sleep(2000);
            ImapPort.Click();
            int imapPort = GenerateUniqueRandomPort();
            int smtpPort;

            do
            {
                smtpPort = GenerateUniqueRandomPort();
            } while (smtpPort == imapPort);

            ImapPort.Patterns.Value.Pattern.SetValue("");
            ImapPort.Patterns.Value.Pattern.SetValue(imapPort.ToString());
            SmtpPort.Click();
            SmtpPort.Patterns.Value.Pattern.SetValue("");
            SmtpPort.Patterns.Value.Pattern.SetValue(smtpPort.ToString());
            Thread.Sleep(2000);
            SaveChangedPorts.Click();
            return this;
        }

        public SettingsMenuWindow SwitchBackToDefaultPorts()
        {
            ChangeDefaultPortsButton.Click();
            Thread.Sleep(2000);
            ImapPort.Click();
            ImapPort.Patterns.Value.Pattern.SetValue("");
            ImapPort.Patterns.Value.Pattern.SetValue("1143");
            SmtpPort.Click();
            SmtpPort.Patterns.Value.Pattern.SetValue("");
            SmtpPort.Patterns.Value.Pattern.SetValue("1025");
            Thread.Sleep(2000);
            SaveChangedPorts.Click();
            return this;
        }

        public SettingsMenuWindow OpenChangeConnectionMode()
        {
            ChangeConnectionModeButton.Click();
            return this;
        }
        public SettingsMenuWindow CancelChangeConnectionMode()
        {
            CancelChangeConnectionModeButton.Click();
            return this;
        }
        public SettingsMenuWindow ChangeConnectionMode()
        {
            ImapSslMode.Click();
            SmtpSslMode.Click();
            Thread.Sleep(2000);
            SaveChangedConnectionMode.Click();
            return this;
        }
        public SettingsMenuWindow SwitchBackToDefaultConnectionMode()
        {
            ImapStarttlsMode.Click();
            SmtpStarttlsMode.Click();
            Thread.Sleep(2000);
            SaveChangedConnectionMode.Click();
            return this;
        }

        public SettingsMenuWindow ConfigureLocalCache()
        {
            ConfigureLocalCacheButton.Click();
            return this;
        }
        public SettingsMenuWindow CancelToConfigureLocalCache()
        {
            Cancel.Click();
            return this;
        }

        public void FocusOnSelectCacheLocationWindow()
        {
            if (SelectCacheLocationWindow != null)
            {
                SelectCacheLocationWindow.Focus();
                Console.WriteLine("Focused and interacted with 'Select cache location' window.");
            }
            else
            {
                Console.WriteLine("The 'Select cache location' window was not found.");
            }
        }
        public void AssertOldCachefolderIsDeleted()
        {
            string? userProfilePath = Environment.GetEnvironmentVariable("USERPROFILE");
            if (string.IsNullOrEmpty(userProfilePath))
            {
                Console.WriteLine("User profile path not found.");
                return;
            }
            string folderPath = Path.Combine(userProfilePath, "AppData", "Roaming", "protonmail", "bridge-v3", "gluon", "NewCacheFolder");
            try
            {
                if (Directory.Exists(folderPath))
                {
                    Directory.Delete(folderPath, recursive: true);
                    Console.WriteLine($"Folder '{folderPath}' deleted successfully.");
                }
                else
                {
                    Console.WriteLine($"Folder '{folderPath}' does not exist.");
                }
            }
            catch (Exception ex)
            {
                Console.WriteLine($"An error occurred while deleting the folder: {ex.Message}");
            }
        }

        public SettingsMenuWindow ChangeAndSwitchBackLocalCacheLocation()
        {
            string? userProfilePath = Environment.GetEnvironmentVariable("USERPROFILE");
            ChangeLocalCacheLocationButton.Click();
            Thread.Sleep(2000);
            FocusOnSelectCacheLocationWindow();
            ClickNewFolder.Click();
            Wait.UntilInputIsProcessed(TimeSpan.FromMilliseconds(2000));
            Keyboard.TypeVirtualKeyCode(0x0D);
            AutomationElement pane = Window.FindFirstDescendant(cf => cf.ByControlType(ControlType.Pane));
            AutomationElement pane2 =  pane.FindFirstDescendant(cf => cf.ByControlType(ControlType.Pane).And(cf.ByName("Shell Folder View")));
            AutomationElement list = pane2.FindFirstDescendant(cf => cf.ByControlType(ControlType.List).And(cf.ByName("Items View")));
            AutomationElement listItem = list.FindFirstDescendant(cf => cf.ByControlType(ControlType.ListItem).And(cf.ByName("New folder")));
            TextBox folderName = listItem.FindFirstDescendant(cf => cf.ByControlType(ControlType.Edit)).AsTextBox();
            folderName.Text = "NewCacheFolder";
            Keyboard.TypeVirtualKeyCode(0x0D); //press Enter
            SelectFolderButton.Click();
            Assert.That(CacheLocation.Name, Is.EqualTo(userProfilePath + "\\AppData\\Roaming\\protonmail\\bridge-v3\\gluon\\NewCacheFolder"));
            SaveChangedCacheFolderLocation.Click();
            WaitUntilElementIsVisible(() => Window.FindFirstDescendant(cf => cf.ByControlType(ControlType.Window)), 60);
            Assert.That(CacheLocationIsChangedNotification.IsAvailable, Is.True);
            OkCacheLocationChangedNotification.Click();
            Back.Click();
            Thread.Sleep(1000);
            ConfigureLocalCacheButton.Click();
            ChangeLocalCacheLocationButton.Click();
            FocusOnSelectCacheLocationWindow();
            Wait.UntilInputIsProcessed(TimeSpan.FromSeconds(1));
            UpArrowToGoBackToPreviousFolder.Click();
            UpArrowToGoBackToPreviousFolder.Click();
            UpArrowToGoBackToPreviousFolder.Click();
            Wait.UntilInputIsProcessed(TimeSpan.FromSeconds(1));
            SelectFolderButton.Click();
            SaveChangedCacheFolderLocation.Click();
            WaitUntilElementIsVisible(() => Window.FindFirstDescendant(cf => cf.ByControlType(ControlType.Window)), 60);
            OkCacheLocationChangedNotification.Click();
            Back.Click();
            Thread.Sleep(2000);
            AssertOldCachefolderIsDeleted();
            Thread.Sleep(1000);
            return this;
        }
        public void FocusOnSelectTLSCertificatesWindow()
        {
            if (SelectDirectoryWindow != null)
            {
                SelectDirectoryWindow.Focus();
                Console.WriteLine("Focused and interacted with 'Directory' window.");
            }
            else
            {
                Console.WriteLine("The 'Directory' window was not found.");
            }
        }

        public void AssertCertificatesAreExported()
        {
            string? userProfilePath = Environment.GetEnvironmentVariable("USERPROFILE");
            string folderPath = Path.Combine(userProfilePath, "TLSCertificates");
            if (string.IsNullOrEmpty(userProfilePath))
            {
                Console.WriteLine("User profile path not found.");
                return;
            }
            string certFilePath = Path.Combine(folderPath, "cert.pem");
            string keyFilePath = Path.Combine(folderPath, "key.pem");
            if (Directory.Exists(folderPath))
            {
                Console.WriteLine("The TLSCertificates folder exists.");
                if (File.Exists(certFilePath))
                {
                    Console.WriteLine("The cert.pem file exists.");
                }
                else
                {
                    Console.WriteLine("The cert.pem file does not exist.");
                }
                if (File.Exists(keyFilePath))
                {
                    Console.WriteLine("The key.pem file exists.");
                }
                else
                {
                    Console.WriteLine("The key.pem file does not exist.");
                }
            }
            else
            {
                Console.WriteLine("The TLSCertificates folder does not exist.");
            }
        }
        public SettingsMenuWindow ExportAssertDeleteTLSCertificates()
        {
            ExportTLSCertificatesButton.Click();
            Thread.Sleep(2000);
            ClickNewFolder.Click();
            Wait.UntilInputIsProcessed(TimeSpan.FromMilliseconds(2000));
            Keyboard.TypeVirtualKeyCode(0x0D);
            AutomationElement pane = Window.FindFirstDescendant(cf => cf.ByControlType(ControlType.Pane));
            AutomationElement pane2 = pane.FindFirstDescendant(cf => cf.ByControlType(ControlType.Pane).And(cf.ByName("Shell Folder View")));
            AutomationElement list = pane2.FindFirstDescendant(cf => cf.ByControlType(ControlType.List).And(cf.ByName("Items View")));
            AutomationElement listItem = list.FindFirstDescendant(cf => cf.ByControlType(ControlType.ListItem).And(cf.ByName("New folder")));
            TextBox folderName = listItem.FindFirstDescendant(cf => cf.ByControlType(ControlType.Edit)).AsTextBox();
            folderName.Text = "TLSCertificates";
            Keyboard.TypeVirtualKeyCode(0x0D); //press Enter
            Wait.UntilInputIsProcessed(TimeSpan.FromSeconds(1));
            SelectFolderButton.Click();
            Thread.Sleep(5000);
            AssertCertificatesAreExported();
            Thread.Sleep(10000);
            ExportTLSCertificatesButton.Click();
            FocusOnSelectTLSCertificatesWindow();
            Thread.Sleep(3000);
            pane = Window.FindFirstDescendant(cf => cf.ByControlType(ControlType.Pane));
            pane2 = pane.FindFirstDescendant(cf => cf.ByControlType(ControlType.Pane).And(cf.ByName("Shell Folder View")));
            list = pane2.FindFirstDescendant(cf => cf.ByControlType(ControlType.List).And(cf.ByName("Items View")));
            var TLSFolder = list.FindFirstDescendant(cf => cf.ByControlType(ControlType.ListItem).And(cf.ByName("TLSCertificates")));
            Assert.That(TLSFolder.IsAvailable, Is.True);
            TLSFolder.Focus(); // Ensure the folder is selected
            var boundingRectangle = TLSFolder.Properties.BoundingRectangle.Value.X;
            Mouse.MoveTo(new Point(TLSFolder.Properties.BoundingRectangle.Value.X, TLSFolder.Properties.BoundingRectangle.Value.Y));
            Mouse.Click(); // Click to ensure selection
            Thread.Sleep(5000);
            Keyboard.TypeVirtualKeyCode(0x2E); // Press the Delete key (0x2E is the virtual key code for Delete)
            Wait.UntilInputIsProcessed(TimeSpan.FromMilliseconds(1000)); // Wait for the delete action to complete
            pane = Window.FindFirstDescendant(cf => cf.ByControlType(ControlType.Pane));
            pane2 = pane.FindFirstDescendant(cf => cf.ByControlType(ControlType.Pane).And(cf.ByName("Shell Folder View")));
            list = pane2.FindFirstDescendant(cf => cf.ByControlType(ControlType.List).And(cf.ByName("Items View")));
            var deletedFolder = list.FindFirstDescendant(cf => cf.ByControlType(ControlType.ListItem).And(cf.ByName("TLSCertificates")));
            Assert.That(deletedFolder, Is.Null, "The folder 'TLSCertificates' was not deleted successfully.");
            Cancel.Click();
            return this;
        }

        private void WaitUntilElementIsVisible(Func<AutomationElement> findElementFunc, int numOfSeconds)
        {
            TimeSpan timeout = TimeSpan.FromSeconds(numOfSeconds);
            Stopwatch stopwatch = Stopwatch.StartNew();


            while (stopwatch.Elapsed < timeout)
            {
                //if element is visible the processing is completed
                var element = findElementFunc();
                if (element != null)
                {
                    return;
                }

                Wait.UntilInputIsProcessed();
                Thread.Sleep(500);
            }

        }
        public SettingsMenuWindow VerifyRepairRestartsSync()
        {
            RepairBridgeButton.Click();
            RepairButtonInPopUp.Click();
            bool syncRestarted = WaitForCondition(() =>
            {
                string syncStatus = GetSyncStatus(Window);
                return !string.IsNullOrEmpty(syncStatus) && syncStatus.Contains("Synchronizing (0%)");
            }, TimeSpan.FromSeconds(30)); // Adjust timeout as needed

            Assert.That(syncRestarted, Is.True, "Sync did not restart after repair.");
            return this;
        }
        private string GetSyncStatus(AutomationElement window)
        {
            var syncStatusElement = window.FindAllDescendants(cf => cf.ByControlType(ControlType.Text)).FirstOrDefault(el =>
                                          {
                                              string name = el.Name;
                                              return !string.IsNullOrEmpty(name) &&
                                                     name.StartsWith("Synchronizing (") &&
                                                     name.EndsWith("%)");
                                          });

            return syncStatusElement?.AsLabel()?.Text ?? string.Empty;
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

        public SettingsMenuWindow VerifyResetAndRestartBridge()
        {
            ResetButton.Click();
            ResetAndRestartButtonInPopUp.Click();
            Thread.Sleep(5000);
            LaunchApp();
            Window.Focus();
            Thread.Sleep(5000);
            Assert.That(StartSetUpButton.IsAvailable, Is.True);
            StartSetUpButton.Click();
            return this;
        }
    }
}