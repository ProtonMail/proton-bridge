using System;
using System.Collections.Generic;
using System.Diagnostics;
using System.Linq;
using System.Text;
using System.Text.Encodings.Web;
using System.Text.Json.Nodes;
using System.Text.Json;
using System.Threading.Tasks;
using ProtonMailBridge.UI.Tests.TestsHelper;
using FlaUI.Core.Definitions;
using FlaUI.Core.AutomationElements;
using FlaUI.Core.Input;
// using System.Windows.Forms;

namespace ProtonMailBridge.UI.Tests.Windows
{
    public class ZeroPercentUpdateWindow : UIActions
    {

        private TextBox BridgeUpdateIsReady => Window.FindFirstDescendant(cf => cf.ByControlType(ControlType.Text).And(cf.ByName("Bridge update is ready"))).AsTextBox();
        private Button StartSetupButton => Window.FindFirstDescendant(cf => cf.ByControlType(ControlType.Button).And(cf.ByName("Start setup"))).AsButton();
        private Button CancelButton => Window.FindFirstDescendant(cf => cf.ByControlType(ControlType.Button).And(cf.ByName("Cancel"))).AsButton();

        private TextBox VersionBuildNumberTextBox => Window.FindFirstDescendant(cf => cf.ByControlType(ControlType.Pane)).FindAllDescendants(cf => cf.ByControlType(ControlType.Text))[13].AsTextBox();
        
        public void editTheVault()
        {
            try
            {
                string executablePath = Environment.GetEnvironmentVariable("UI_TEST_VAULT_EDITOR_EXE_PATH");

                if (string.IsNullOrEmpty(executablePath))
                {
                    Console.WriteLine("Executable path is not set in the environment variables.");
                    return;
                }

                if (!System.IO.File.Exists(executablePath))
                {
                    Console.WriteLine($"Executable not found at path: {executablePath}");
                    return;
                }
                string readArgument = "read";
                string writeArgument = "write";
                string vaultJson;

                // start the read process
                ProcessStartInfo readProcessInfo = new ProcessStartInfo
                {
                    FileName = executablePath,
                    Arguments = readArgument,
                    RedirectStandardOutput = true, // Capture the output
                    UseShellExecute = false, // Required to redirect output
                    CreateNoWindow = true // Optional: Run without showing a console window
                };

                //read the output
                using (Process readProcess = Process.Start(readProcessInfo))
                {
                    if (readProcess == null)
                    {
                        Assert.Fail("Failed to start the read process.");
                        return;
                    }

                    string output = readProcess.StandardOutput.ReadToEnd();
                    readProcess.WaitForExit();

                    // Write the output to a file
                    string desktopPath = Environment.GetFolderPath(Environment.SpecialFolder.Desktop);
                    vaultJson = Path.Combine(desktopPath, "vault.json");
                    File.WriteAllText(vaultJson, output);

                    Console.WriteLine($"Vault data has been written to: {vaultJson}");
                }

                //update the json file, UpdateChannel to early and UpdateRollout to 0
                string jsonContent = File.ReadAllText(vaultJson);
                JsonNode json = JsonNode.Parse(jsonContent);

                if (json == null)
                {
                    Assert.Fail("Can not parse the vault.json file");
                    return;
                }
                else
                {
                    var settingsNode = json["Settings"];

                    settingsNode["UpdateChannel"] = "early";
                    settingsNode["UpdateRollout"] = 0;
                    // must use UnsafeRelaxedJsonEscaping for the serialization of the + character in the timestamp
                    File.WriteAllText(vaultJson, json.ToJsonString(new JsonSerializerOptions { WriteIndented = true, Encoder = JavaScriptEncoder.UnsafeRelaxedJsonEscaping }));
                }

                // read the new content from the json file
                jsonContent = File.ReadAllText(vaultJson);
                // start the write process
                Process writeProcess = new Process
                {
                    StartInfo = new ProcessStartInfo
                    {
                        FileName = executablePath,
                        Arguments = writeArgument,
                        RedirectStandardInput = true, // Pass input via standard input
                        RedirectStandardOutput = true, // Capture standard output
                        RedirectStandardError = true,  // Capture error output
                        UseShellExecute = false, // Required for redirection
                        CreateNoWindow = true   // Run without showing a window
                    }
                };

                writeProcess.Start();

                // pass the json file content to the process
                using (StreamWriter writer = writeProcess.StandardInput)
                {
                    if (writer.BaseStream.CanWrite)
                    {
                        writer.Write(jsonContent);
                    }
                }

                // read the output and errors
                string writeOutput = writeProcess.StandardOutput.ReadToEnd();
                string error = writeProcess.StandardError.ReadToEnd();

                writeProcess.WaitForExit();

                if (!string.IsNullOrWhiteSpace(writeOutput))
                {
                    Console.WriteLine("Output from write process:");
                    Console.WriteLine(writeOutput);
                }

                if (!string.IsNullOrWhiteSpace(error))
                {
                    Console.WriteLine("Error from write process:");
                    Console.WriteLine(error);
                }

                // Check exit code for success/failure
                if (writeProcess.ExitCode != 0)
                {
                    Assert.Fail("Write process exited with error code: " + writeProcess.ExitCode);
                }
            }
            catch (Exception ex)
            {
                Assert.Fail(ex.Message);
            }

        }

        public ZeroPercentUpdateWindow VerifyBetaAccessIsEnabled()
        {
            CheckBox BetaAccess = Window.FindFirstDescendant(cf => cf.ByControlType(ControlType.CheckBox).And(cf.ByName("Beta access toggle"))).AsCheckBox();
            Assert.That(BetaAccess.IsToggled, Is.True);
            return this;
        }

        public ZeroPercentUpdateWindow RestartBridgeNotification()
        {
            WaitUntilElementIsVisible(() => BridgeUpdateIsReady, 60);
            Assert.That(BridgeUpdateIsReady.IsAvailable, Is.True);
            Button RestartBridge = Window.FindFirstDescendant(cf => cf.ByControlType(ControlType.Button).And(cf.ByName("Restart Bridge"))).AsButton();
            RestartBridge.Click();
            Thread.Sleep(5000);
            LaunchApp();
            Window.Focus();
            Thread.Sleep(5000);
            return this;
        }

        private static int[] PreviousVersionNumbers = null;
        private static string PreviousTagNumber = null;

        private int[] GetVersionNumbers()
        {
            string text = VersionBuildNumberTextBox.Name;

            string versionString = text.Substring(text.IndexOf("v") + 1).Split(' ')[0];
            return Array.ConvertAll(versionString.Split('.'), int.Parse);
        }
        private string GetTagNumber()
        {
            string text = VersionBuildNumberTextBox.Name;
            return text.Substring(text.IndexOf("(br-") + 4).Split(')')[0];
        }
        public void SaveCurrentVersionAndTagNumber()
        {
            PreviousVersionNumbers = GetVersionNumbers();
            PreviousTagNumber = GetTagNumber();
        }
        public void VerifyVersionAndTagNumberOnRelaunch()
        {
            int[] newVersionNumbers = GetVersionNumbers();
            string newTagNumber = GetTagNumber();

            if (PreviousVersionNumbers == null || PreviousTagNumber == null)
            {
                throw new Exception("Previous version or tag number is not set.");
            }

            bool isVersionGreater = false;
            for (int i = 0; i < PreviousVersionNumbers.Length; i++)
            {
                if (newVersionNumbers[i] > PreviousVersionNumbers[i])
                {
                    isVersionGreater = true;
                    break;
                }
                else if (newVersionNumbers[i] < PreviousVersionNumbers[i])
                {
                    throw new Exception($"Current version {string.Join('.', newVersionNumbers)} is not greater than previous version {string.Join('.', PreviousVersionNumbers)}.");
                }
            }

            if (!isVersionGreater)
            {
                throw new Exception($"Current version {string.Join('.', newVersionNumbers)} is not greater than previous version {string.Join('.', PreviousVersionNumbers)}.");
            }

            if (newTagNumber == PreviousTagNumber)
            {
                throw new Exception($"Current tag number (br-{newTagNumber}) is the same as the previous tag number (br-{PreviousTagNumber}).");
            }
        }

        public ZeroPercentUpdateWindow ClickStartSetupButton()
        {
            StartSetupButton.Click();
            return this;
        }
        public ZeroPercentUpdateWindow CLickCancelButton()
        {
            CancelButton.Click();
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
    }
}
