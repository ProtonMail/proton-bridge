using System;

namespace ProtonMailBridge.UI.Tests.TestsHelper
{
    public class TestUserData
    {
        public string Username { get; set; }
        public string Password { get; set; }

        public TestUserData(string username, string password)
        {
            Username = username;
            Password = password;
        }

        public static TestUserData GetFreeUser()
        {
            (string username, string password) = GetusernameAndPassword("BRIDGE_FLAUI_FREE_USER");
            return new TestUserData(username, password);
        }

        public static TestUserData GetPaidUser()
        {
            (string username, string password) = GetusernameAndPassword("BRIDGE_FLAUI_PAID_USER");
            return new TestUserData(username, password);
        }

        public static TestUserData GetIncorrectCredentialsUser()
        {
            return new TestUserData("IncorrectUsername", "IncorrectPass");
        }

        private static (string, string) GetusernameAndPassword(string userType)
        {
            // Get the environment variable for the user and check if missing
            // When changing or adding an environment variable, you must restart Visual Studio
            // if you have it open while doing this
            string? str = Environment.GetEnvironmentVariable(userType);
            if (string.IsNullOrEmpty(str))
            {
                throw new Exception($"Missing environment variable: {userType}");
            }

            // Check if the environment variable contains only one ':'
            // The ':' character must be between the username/email and password
            string ch = ":";
            if ((str.IndexOf(ch) != str.LastIndexOf(ch)) | (str.IndexOf(ch) == -1))
            {
                throw new Exception(
                    $"Environment variable {str} must contain one ':' and it must be between username and password!"
                    );
            }

            string[] split = str.Split(':');
            return (split[0], split[1]);
        }
    }
}
