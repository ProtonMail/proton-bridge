using System;
using System.Threading;
using FlaUI.Core.AutomationElements;
using FlaUI.Core;
using FlaUI.UIA3;
using ProtonMailBridge.UI.Tests.TestsHelper;
using FlaUI.Core.Input;

namespace ProtonMailBridge.UI.Tests
{
    public class TestSession
    {

        public static Application App;
        protected static Application Service;
        protected static Window Window;

        protected static void ClientCleanup()
        {
            App.Kill();
            App.Dispose();
            // Give some time to properly exit the app
            Thread.Sleep(2000);
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