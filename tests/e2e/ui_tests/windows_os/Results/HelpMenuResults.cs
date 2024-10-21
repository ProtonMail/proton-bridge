using FlaUI.Core.AutomationElements;
using FlaUI.Core.Definitions;
using ProtonMailBridge.UI.Tests.TestsHelper;
using FlaUI.Core.Input;
using System.DirectoryServices;
using System.Net;

namespace ProtonMailBridge.UI.Tests.Results
{
    public class HelpMenuResult : UIActions
    {
        private AutomationElement NotificationWindow => Window.FindFirstDescendant(cf => cf.ByControlType(ControlType.Window));
        private AutomationElement[] TextFields => Window.FindAllDescendants(cf => cf.ByControlType(ControlType.Text));
        private TextBox HelpText => Window.FindFirstDescendant(cf => cf.ByControlType(ControlType.Text).And(cf.ByName("Help"))).AsTextBox();
        private TextBox BridgeIsUpToDate => NotificationWindow.FindFirstDescendant(cf => cf.ByControlType(ControlType.Text)).AsTextBox();
        private AutomationElement ChromeTab => ChromeWindow.FindFirstDescendant(cf => cf.ByControlType(ControlType.Document));
        private TextBox ChromeText => ChromeTab.FindFirstDescendant(cf => cf.ByControlType(ControlType.Text).And(cf.ByName("We can help you with every step of using Proton Mail Bridge."))).AsTextBox();

        //private AutomationElement AdressBar => FileExplorerWindow.FindFirstDescendant(cf => cf.ByControlType(ControlType.Group).And(cf.ByAutomationId("PART_BreadcrumbBar")));

        private AutomationElement AdressPane => FileExplorerWindow.FindFirstDescendant(cf => cf.ByControlType(ControlType.Pane).And(cf.ByClassName("Microsoft.UI.Content.DesktopChildSiteBridge")));
        private AutomationElement AdressBar => AdressPane.FindFirstDescendant(cf =>cf.ByControlType(ControlType.Group).And(cf.ByAutomationId("PART_BreadcrumbBar")));
        private AutomationElement[] Folders => AdressBar.FindAllDescendants(cf => cf.ByControlType(ControlType.SplitButton));

        private TextBox SendReportConfirmation => NotificationWindow.FindFirstDescendant(cf => cf.ByControlType(ControlType.Text).And(cf.ByName("Thank you for the report. We'll get back to you as soon as we can."))).AsTextBox();

        public HelpMenuResult CheckIfUserOpenedHelpMenu()
        {
            Assert.That(HelpText.IsAvailable, Is.True);
            return this;
        }

        public HelpMenuResult CheckBridgeIsUpToDateNotification()
        {
            Assert.That(BridgeIsUpToDate.IsAvailable, Is.True);
            return this;
        }

        public HelpMenuResult CheckHelpLinkIsOpen()
        {
            Assert.That(ChromeText.IsAvailable, Is.True);
            return this;
        }

        public HelpMenuResult CheckBridgeLogsAreOpen()
        {
            var adressName = "";
            foreach (var folder in Folders)
            {
                var folderName = folder.Name;
                adressName = System.IO.Path.Combine(adressName, folderName);
            }
            var expectedPath = "\\AppData\\Roaming\\protonmail\\bridge-v3\\logs";
            Assert.That(adressName.Contains(expectedPath), Is.True);
            return this;
        }

        public HelpMenuResult CheckIfProblemIsSuccReported()
        {
            Assert.That(SendReportConfirmation.IsAvailable, Is.True);
            return this;
        }
    }
}