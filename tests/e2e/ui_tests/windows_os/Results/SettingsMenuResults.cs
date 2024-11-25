using System;
using System.Collections.Generic;
using System.Linq;
using System.Text;
using System.Threading.Tasks;
using FlaUI.Core.AutomationElements;
using FlaUI.Core.Definitions;
using ProtonMailBridge.UI.Tests.TestsHelper;
using FlaUI.Core.Input;
using System.DirectoryServices;
using System.Net;
using System;
using System.Collections.Generic;
using System.Linq;
using System.Threading.Tasks;
using System;
using System.Collections.Generic;
using System.Linq;
using System.Threading.Tasks;
using FlaUI.Core.AutomationElements.Scrolling;
using FlaUI.Core.WindowsAPI;
using Microsoft.VisualBasic.Devices;
using NUnit.Framework.Legacy;
using ProtonMailBridge.UI.Tests.Results;
using Keyboard = FlaUI.Core.Input.Keyboard;
using Mouse = FlaUI.Core.Input.Mouse;
using static System.Windows.Forms.VisualStyles.VisualStyleElement.Window;
using ProtonMailBridge.UI.Tests.Windows;

namespace ProtonMailBridge.UI.Tests.Results
{
    public class SettingsMenuResults : UIActions
    {
        private AutomationElement[] TextFields => Window.FindAllDescendants(cf => cf.ByControlType(ControlType.Text));
        private AutomationElement Pane => Window.FindFirstDescendant(cf => cf.ByControlType(ControlType.Window));
        private CheckBox AutomaticUpdates => Window.FindFirstDescendant(cf => cf.ByControlType(ControlType.CheckBox).And(cf.ByName("Automatic updates toggle"))).AsCheckBox();
        private CheckBox OpenOnStartUp => Window.FindFirstDescendant(cf => cf.ByControlType(ControlType.CheckBox).And(cf.ByName("Open on startup toggle"))).AsCheckBox();
        private CheckBox BetaAccess => Window.FindFirstDescendant(cf => cf.ByControlType(ControlType.CheckBox).And(cf.ByName("Beta access toggle"))).AsCheckBox();
        private CheckBox AlternativeRouting => Window.FindFirstDescendant(cf => cf.ByControlType(ControlType.CheckBox).And(cf.ByName("Alternative routing toggle"))).AsCheckBox();
        private CheckBox DarkMode => Window.FindFirstDescendant(cf => cf.ByControlType(ControlType.CheckBox).And(cf.ByName("Dark mode toggle"))).AsCheckBox();
        private CheckBox ShowAllMail => Window.FindFirstDescendant(cf => cf.ByControlType(ControlType.CheckBox).And(cf.ByName("Show All Mail toggle"))).AsCheckBox();
        private CheckBox CollectUsageDiagnostics => Window.FindFirstDescendant(cf => cf.ByControlType(ControlType.CheckBox).And(cf.ByName("Collect usage diagnostics toggle"))).AsCheckBox();
        private TextBox ImapPort => Window.FindFirstDescendant(cf => cf.ByControlType(ControlType.Edit).And(cf.ByName("IMAP port edit"))).AsTextBox();
        private TextBox SmtpPort => Window.FindFirstDescendant(cf => cf.ByControlType(ControlType.Edit).And(cf.ByName("SMTP port edit"))).AsTextBox();
        private AutomationElement[] RadioButtons => Window.FindAllDescendants(cf => cf.ByControlType(ControlType.RadioButton));
        private RadioButton ImapStarttlsMode => RadioButtons[1].AsRadioButton();
        private RadioButton SmtpStarttlsMode => RadioButtons[3].AsRadioButton();
        private RadioButton ImapSslMode => RadioButtons[0].AsRadioButton();
        private RadioButton SmtpSslMode => RadioButtons[2].AsRadioButton();
        private TextBox CacheLocation => TextFields[9].AsTextBox();
        public SettingsMenuResults AutomaticUpdatesIsEnabledByDefault()
        {
            Assert.That(AutomaticUpdates.IsToggled, Is.True);
            return this;
        }

        public SettingsMenuResults OpenOnStartUpIsEnabledByDefault()
        {
            Assert.That(OpenOnStartUp.IsToggled, Is.True);
            return this;
        }

        public SettingsMenuResults BetaAccessIsDisabledByDefault()
        {
            Assert.That(BetaAccess.IsToggled, Is.False);
            return this;
        }

        public SettingsMenuResults AlternativeRoutingIsDisabledByDefault()
        {
            Assert.That(AlternativeRouting.IsToggled, Is.False);
            return this;
        }

        public SettingsMenuResults DarkModeIsDisabledByDefault()
        {
            Assert.That(DarkMode.IsToggled, Is.False);
            return this;
        }
        public SettingsMenuResults ShowAllMailIsEnabledByDefault()
        {
            Assert.That(ShowAllMail.IsToggled, Is.True);
            return this;
        }
        public SettingsMenuResults CollectUsageDiagnosticsIsEnabledByDefault()
        {
            Assert.That(CollectUsageDiagnostics.IsToggled, Is.True);
            return this;
        }

        public SettingsMenuResults VerifyDefaultPorts()
        {
            Assert.That(ImapPort.Patterns.Value.Pattern.Value, Is.AnyOf("1143", "1144", "1045"));
            Assert.That(SmtpPort.Patterns.Value.Pattern.Value, Is.AnyOf("1025", "1026", "1027"));
            return this;
        }
        public SettingsMenuResults VerifyDefaultConnectionMode()
        {
            Assert.That(ImapStarttlsMode.IsChecked, Is.True);
            Assert.That(SmtpStarttlsMode.IsChecked, Is.True);
            return this;
        }

        public SettingsMenuResults AssertTheChangedConnectionMode()
        {
            Assert.That(ImapSslMode.IsChecked, Is.True);
            Assert.That(SmtpSslMode.IsChecked, Is.True);
            return this;
        }

        public SettingsMenuResults DefaultCacheLocation()
        {
            string userProfilePath = Environment.GetEnvironmentVariable("USERPROFILE");
            Assert.That(CacheLocation.Name, Is.EqualTo(userProfilePath + "\\AppData\\Roaming\\protonmail\\bridge-v3\\gluon"));
            return this;
        }
    }
}
