using System;
using System.Collections.Generic;
using System.Linq;
using System.Text;
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

namespace ProtonMailBridge.UI.Tests.Windows
{
    public class HelpMenuWindow : UIActions
    {
        private AutomationElement[] InputFields => Window.FindAllDescendants(cf => cf.ByControlType(ControlType.Edit));
        private AutomationElement[] HomeButtons => Window.FindAllDescendants(cf => cf.ByControlType(ControlType.Button));
        private AutomationElement NotificationWindow => Window.FindFirstDescendant(cf => cf.ByControlType(ControlType.Window));
        private AutomationElement[] ReportProblemPane => Window.FindAllDescendants(cf => cf.ByControlType(ControlType.Pane));

        private Button HelpButton => HomeButtons[3].AsButton();
        private Button BackToAccountViewButton => HomeButtons[11].AsButton();
        private Button GoToHelpTopics => HomeButtons[7].AsButton();
        private Button CheckNow => HomeButtons[8].AsButton();
        private Button ConfirmNotificationButton => NotificationWindow.FindFirstDescendant(cf => cf.ByControlType(ControlType.Button).And(cf.ByName("OK"))).AsButton();
        private Button LogsButton => HomeButtons[9].AsButton();
        private Button ReportProblemButton => HomeButtons[10].AsButton();
        private Button ICannotFindEmailInClient => HomeButtons[7].AsButton();
        private TextBox DescriptionOnWhatHappened => Window.FindFirstDescendant(cf => cf.ByControlType(ControlType.Edit)).AsTextBox();
        private RadioButton MissingEmails => ReportProblemPane[0].FindFirstDescendant(cf => cf.ByControlType(ControlType.RadioButton).And(cf.ByName("Old emails are missing"))).AsRadioButton();
        private RadioButton FindEmails => ReportProblemPane[0].FindFirstDescendant(cf => cf.ByControlType(ControlType.RadioButton).And(cf.ByName("Yes"))).AsRadioButton();
        private CheckBox VPNSoftware => ReportProblemPane[0].FindFirstDescendant(cf => cf.ByControlType(ControlType.CheckBox).And(cf.ByName("VPN"))).AsCheckBox();
        private CheckBox FirewallSoftware => ReportProblemPane[0].FindFirstDescendant(cf => cf.ByControlType(ControlType.CheckBox).And(cf.ByName("Firewall"))).AsCheckBox();
        private Button ContinueToReportButton => ReportProblemPane[0].FindFirstDescendant(cf => cf.ByControlType(ControlType.Button).And(cf.ByName("Continue"))).AsButton();
        private AutomationElement ScrollBarFirstPane => ReportProblemPane[0].FindFirstDescendant(cf => cf.ByControlType(ControlType.ScrollBar));
        private Button SendReportButton => ReportProblemPane[0].FindFirstDescendant(cf => cf.ByControlType(ControlType.Button).And(cf.ByName("Send"))).AsButton();
        private Button NotAbleToSendEmailsButton => HomeButtons[8].AsButton();
        private Button BridgeIsNotStartingCorrectlyButton => HomeButtons[9].AsButton();
        private TextBox StepByStepActionsForBridgeIsNotStartingCorrectly => ReportProblemPane[2].AsTextBox();
        private TextBox IssuesLastOccurence => ReportProblemPane[3].AsTextBox();
        private TextBox QuestionFocusWhenWasLastOccurence => ReportProblemPane[0].FindFirstDescendant(cf => cf.ByControlType(ControlType.Text).And(cf.ByName("When did the issue last occur? Is it repeating?"))).AsTextBox();
        private TextBox QuestionFocusOnStepByStepActions => ReportProblemPane[0].FindFirstDescendant(cf => cf.ByControlType(ControlType.Text).And(cf.ByName("What were the step-by-step actions you took that led to this happening?"))).AsTextBox();
        private Button BridgeIsRunningSlowButton => HomeButtons[10].AsButton();
        private TextBox StepByStepActionsForBridgeIsSlow => ReportProblemPane[2].FindFirstDescendant(cf => cf.ByControlType(ControlType.Edit)).AsTextBox();
        private CheckBox ExperiencingIssues => ReportProblemPane[0].FindFirstDescendant(cf => cf.ByControlType(ControlType.CheckBox).And(cf.ByName("Emails arrive with a delay"))).AsCheckBox();
        private Button SomethingElseButton => HomeButtons[11].AsButton();
        private TextBox StepByStepActionsForSomethingElse => ReportProblemPane[3].AsTextBox();
        private TextBox IssuesLastOccurenceInSomethingElseProblemSection => ReportProblemPane[4].AsTextBox();
        private TextBox OverviewOfReportProblemDetails => ReportProblemPane[1].FindFirstDescendant(cf => cf.ByControlType(ControlType.Edit)).AsTextBox();
        private TextBox ContactEmailInReportProblem => ReportProblemPane[0].FindAllDescendants(cf => cf.ByControlType(ControlType.Edit)).ToList()[1].AsTextBox();
        public CheckBox IncludeLogs => ReportProblemPane[0].FindFirstDescendant(cf => cf.ByControlType(ControlType.CheckBox)).AsCheckBox();
        public HelpMenuWindow ClickHelpButton()
        {
            HelpButton.Click();

            return this;
        }

        public HelpMenuWindow ClickBackFromHelpMenu()
        {
            BackToAccountViewButton.Click();
            return this;
        }

        public HelpMenuWindow ClickGoToHelpTopics()
        {
            GoToHelpTopics.Click();
            return this;
        }
        public HelpMenuWindow ClickCheckNowButton()
        {
            CheckNow.Click();
            return this;
        }

        public HelpMenuWindow ConfirmNotification()
        {
            ConfirmNotificationButton.Click();
            return this;
        }
        public HelpMenuWindow ClickLogsButton()
        {
            LogsButton.Click();
            return this;
        }

        public HelpMenuWindow ClickReportProblemButton()
        {
            ReportProblemButton.Click();
            return this;
        }

        public HelpMenuWindow ClickICannotFindEmailsInEmailClient()
        {
            ICannotFindEmailInClient.Click();
            return this;
        }

        public HelpMenuWindow ClickToFillQuestionDescription()
        {
            DescriptionOnWhatHappened.Click();
            return this;
        }

        public HelpMenuWindow ClickNotAbleToSendEmails()
        {
            NotAbleToSendEmailsButton.Click();
            return this;
        }
        public HelpMenuWindow ClickBridgeIsNotStartingCorrectly()
        {
            BridgeIsNotStartingCorrectlyButton.Click();
            return this;
        }
        public HelpMenuWindow ClickBridgeIsRunningSlow()
        {
            BridgeIsRunningSlowButton.Click();
            return this;
        }

        public HelpMenuWindow ClickSomethingElse()
        {
            SomethingElseButton.Click();
            return this;
        }

        public HelpMenuWindow EnterMissingEmailsProblemDetails()
        {
            DescriptionOnWhatHappened.Enter("I am missing emails in my email client.");
            MissingEmails.Click();
            FindEmails.Click();
            VPNSoftware.IsChecked = true;
            FirewallSoftware.IsChecked = true;
            Mouse.Scroll(-20);
            Wait.UntilInputIsProcessed(TimeSpan.FromSeconds(1));
            ContinueToReportButton.Click();
            VerifyOverviewOfMissingEmailsDetails();
            VerifyContactEmail();
            VerifyIncludeLogsIsChecked();
            SendReportButton.Click();
            Wait.UntilInputIsProcessed(TimeSpan.FromSeconds(5));
            return this;
        }

        public HelpMenuWindow EnterNotAbleToSendEmailProblemDetails()
        {
            DescriptionOnWhatHappened.Enter("I am not able to send emails.");
            StepByStepActionsForBridgeIsNotStartingCorrectly.Enter("I compose a message, I click Send and I get an error that the message cannot be sent.");
            IssuesLastOccurence.Enter("It happened this morning for the first time.");
            QuestionFocusWhenWasLastOccurence.Click();
            VPNSoftware.IsChecked = true;
            FirewallSoftware.IsChecked = true;
            Mouse.Scroll(-20);
            Wait.UntilInputIsProcessed(TimeSpan.FromSeconds(1));
            ContinueToReportButton.Click();
            VerifyOverviewOfNotAbleToSendEmailsDetails();
            VerifyContactEmail();
            VerifyIncludeLogsIsChecked();
            SendReportButton.Click();
            Wait.UntilInputIsProcessed(TimeSpan.FromSeconds(5));
            return this;
        }
        public HelpMenuWindow EnterBridgeIsNotStartingCorrectlyProblemDetails()
        {
            DescriptionOnWhatHappened.Enter("Bridge is not starting correctly.");
            StepByStepActionsForBridgeIsNotStartingCorrectly.Enter("I turned on my device, and Bridge couldn't launch, I received an error.");
            IssuesLastOccurence.Enter("It occured today for the first time and I cannot fix it.");
            QuestionFocusWhenWasLastOccurence.Click();
            VPNSoftware.IsChecked = true;
            FirewallSoftware.IsChecked = true;
            Mouse.Scroll(-20);
            Wait.UntilInputIsProcessed(TimeSpan.FromSeconds(1));
            ContinueToReportButton.Click();
            VerifyOverviewOfBridgeNotStartingCorrectlyDetails();
            VerifyContactEmail();
            VerifyIncludeLogsIsChecked();
            SendReportButton.Click();
            Wait.UntilInputIsProcessed(TimeSpan.FromSeconds(5));
            return this;
        }

        public HelpMenuWindow EnterBridgeIsRunningSlowProblemDetails()
        {
            DescriptionOnWhatHappened.Enter("Bridge is really slow.");
            StepByStepActionsForBridgeIsSlow.Enter("I started Bridge, added an account and the sync takes forever.");
            ExperiencingIssues.IsChecked = true;
            VPNSoftware.IsChecked = true;
            FirewallSoftware.IsChecked = true;
            Mouse.Scroll(-20);
            Wait.UntilInputIsProcessed(TimeSpan.FromSeconds(1));
            ContinueToReportButton.Click();
            VerifyOverviewOfBridgeIsRunningSlowDetails();
            VerifyContactEmail();
            VerifyIncludeLogsIsChecked();
            SendReportButton.Click();
            Wait.UntilInputIsProcessed(TimeSpan.FromSeconds(5));
            return this;
        }

        public HelpMenuWindow EnterSomethingElseProblemDetails()
        {
            DescriptionOnWhatHappened.Enter("I don't receive emails.");
            StepByStepActionsForBridgeIsSlow.Enter("I am expecting an email that is sent, but it hasn't arrived in my Inbox.");
            StepByStepActionsForSomethingElse.Enter("I click Get messages, but the emails that are sent to me do not arrive.");
            QuestionFocusOnStepByStepActions.Click();
            Mouse.Scroll(-20);
            IssuesLastOccurenceInSomethingElseProblemSection.Enter("Issue started happening today.");
            QuestionFocusWhenWasLastOccurence.Click();
            ContinueToReportButton.Click();
            VerifyOverviewOfSomethingElseDetails();
            VerifyContactEmail();
            VerifyIncludeLogsIsChecked();
            SendReportButton.Click();
            Wait.UntilInputIsProcessed(TimeSpan.FromSeconds(5));
            return this;
        }
        public HelpMenuWindow VerifyOverviewOfMissingEmailsDetails()
        {
            Assert.That(OverviewOfReportProblemDetails.Text, Does.Contain("Please describe what happened and include any error messages.\nI am missing emails in my email client.\nAre you missing emails from the email client or not receiving new ones?\nOld emails are missing\nCan you find the emails in the web/mobile application?\nYes\nAre you running any of these software? Select all that apply.\nVPN, Firewall"));
            return this;
        }

        public HelpMenuWindow VerifyOverviewOfNotAbleToSendEmailsDetails()
        {
            Assert.That(OverviewOfReportProblemDetails.Text, Does.Contain("Please describe what happened and include any error messages.\nI am not able to send emails.\nWhat were the step-by-step actions you took that led to this happening?\nI compose a message, I click Send and I get an error that the message cannot be sent.\nWhen did the issue last occur? Is it repeating?\nIt happened this morning for the first time.\nAre you running any of these software? Select all that apply.\nVPN, Firewall"));
            return this;
        }
        public HelpMenuWindow VerifyOverviewOfBridgeNotStartingCorrectlyDetails()
        {
            Assert.That(OverviewOfReportProblemDetails.Text, Does.Contain("Please describe what happened and include any error messages.\nBridge is not starting correctly.\nWhat were the step-by-step actions you took that led to this happening?\nI turned on my device, and Bridge couldn't launch, I received an error.\nWhen did the issue last occur? Is it repeating?\nIt occured today for the first time and I cannot fix it.\nAre you running any of these software? Select all that apply.\nVPN, Firewall"));
            return this;
        }
        public HelpMenuWindow VerifyOverviewOfBridgeIsRunningSlowDetails()
        {
            Assert.That(OverviewOfReportProblemDetails.Text, Does.Contain("Please describe what happened and include any error messages.\nBridge is really slow.\nWhat were the step-by-step actions you took that led to this happening?\nI started Bridge, added an account and the sync takes forever.\nWhich of these issues are you experiencing?\nEmails arrive with a delay\nAre you running any of these software? Select all that apply.\nVPN, Firewall"));
            return this;
        }
        public HelpMenuWindow VerifyOverviewOfSomethingElseDetails()
        {
            Assert.That(OverviewOfReportProblemDetails.Text, Does.Contain("Please describe what happened and include any error messages.\nI don't receive emails.\nWhat did you want or expect to happen?\nI am expecting an email that is sent, but it hasn't arrived in my Inbox.\nWhat were the step-by-step actions you took that led to this happening?\nI click Get messages, but the emails that are sent to me do not arrive.\nWhen did the issue last occur? Is it repeating?\nIssue started happening today."));
            return this;
        }
        public HelpMenuWindow VerifyContactEmail()
        {
            Assert.That(ContactEmailInReportProblem.Text, Is.EqualTo(TestUserData.GetPaidUser().Username));
            return this;
        }
        public HelpMenuWindow VerifyIncludeLogsIsChecked()
        {
            Assert.That(IncludeLogs.IsChecked, Is.True);
            return this;
        }
    }
}