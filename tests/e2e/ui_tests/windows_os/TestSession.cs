using System;
using System.Threading;
using FlaUI.Core.AutomationElements;
using FlaUI.Core;
using FlaUI.UIA3;
using ProtonMailBridge.UI.Tests.TestsHelper;
using FlaUI.Core.Input;
using System.Diagnostics;

namespace ProtonMailBridge.UI.Tests
{
    public class TestSession
    {

        public static Application App;
        protected static Application Service;
        protected static Window Window;
        protected static Window ChromeWindow;
        protected static Window FileExplorerWindow;

        protected static void ClientCleanup()
        {
            App.Kill();
            App.Dispose();
            // Give some time to properly exit the app
            Thread.Sleep(10000);
        }

        public static void switchToFileExplorerWindow()
        {
            var _automation = new UIA3Automation();
            var desktop = _automation.GetDesktop();

            var _explorerWindow = desktop.FindFirstDescendant(cf => cf.ByClassName("CabinetWClass"));

            // If the File Explorer window is not found, fail the test
            if (_explorerWindow == null)
            {
                throw new Exception("File Explorer window not found.");
            }

            // Cast the found element to a Window object
            FileExplorerWindow = _explorerWindow.AsWindow();

            // Focus on the File Explorer window
            FileExplorerWindow.Focus();
        }

        public static void switchToChromeWindow()
        {
            var _automation = new UIA3Automation();
            var desktop = _automation.GetDesktop();

            var _chromeWindow = desktop.FindFirstDescendant(cf => cf.ByClassName("Chrome_WidgetWin_1"));

            // If the Chrome window is not found, fail the test
            if (_chromeWindow == null)
            {
                throw new Exception("Google Chrome window not found.");
            }

            // Cast the found element to a Window object
             ChromeWindow = _chromeWindow.AsWindow();

            // Focus on the Chrome window
            ChromeWindow.Focus();
        }


        public static void LaunchApp()
        {
            string appExecutable = TestData.AppExecutable;
            Application.Launch(appExecutable);
            Wait.UntilInputIsProcessed(TestData.FiveSecondsTimeout);
            App = Application.Attach("bridge-gui.exe");

            try
            {
                Window = App.GetMainWindow(new UIA3Automation(), TestData.ThirtySecondsTimeout);
            }
            catch (System.TimeoutException)
            {
                Assert.Fail("Failed to get window of application!");
            }
        }
    }
}