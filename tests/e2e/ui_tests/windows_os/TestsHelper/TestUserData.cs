using System;
using System.Diagnostics;

namespace ProtonMailBridge.UI.Tests.TestsHelper
{
    public class TestUserData
    {
        public string Username { get; set; }
        public string Password { get; set; }
        public string MailboxPassword { get; set; }

        public TestUserData(string username, string password, string mailboxPassword ="")
        {
            Username = username;
            Password = password;
            MailboxPassword = mailboxPassword;
        }

        public static TestUserData GetFreeUser()
        {
            (string username, string password) = GetUsernameAndPassword("BRIDGE_FLAUI_FREE_USER");
            return new TestUserData(username, password);
        }

        public static TestUserData GetPaidUser()
        {
            (string username, string password) = GetUsernameAndPassword("BRIDGE_FLAUI_PAID_USER");
            return new TestUserData(username, password);
        }

        public static TestUserData GetMailboxUser()
        {
            (string username, string password, string mailboxPassword) = GetUsernameAndPasswordAndMailbox("BRIDGE_FLAUI_MAILBOX_USER");
            return new TestUserData(username, password, mailboxPassword);
        }
        
        public static TestUserData GetDisabledUser()
        {
            (string username, string password) = GetUsernameAndPassword("BRIDGE_FLAUI_DISABLED_USER");
            return new TestUserData(username, password);
        }

        public static TestUserData GetDeliquentUser()
        {
            (string username, string password) = GetUsernameAndPassword("BRIDGE_FLAUI_DELIQUENT_USER");
            return new TestUserData(username, password);
        }
        public static TestUserData GetIncorrectCredentialsUser()
        {
            return new TestUserData("IncorrectUsername", "IncorrectPass");
        }

        public static TestUserData GetEmptyCredentialsUser()
        {
            return new TestUserData("", "");
        }

        public static TestUserData GetAliasUser()
        {
            (string username, string password) = GetUsernameAndPassword("BRIDGE_FLAUI_ALIAS_USER");
            return new TestUserData(username, password);
        }
        private static (string, string) GetUsernameAndPassword(string userType)
        {
            // Get the environment variable for the user and check if missing
            // When changing or adding an environment variable, you must restart Visual Studio
            // if you have it open while doing this
            string? str = Environment.GetEnvironmentVariable(userType);
            if (string.IsNullOrEmpty(str))
            {
                throw new Exception($"Environment variable {userType} must contain one ':' and it must be between username and password!");
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
        private static (string, string, string) GetUsernameAndPasswordAndMailbox(string userType)
        {
            // Get the environment variable for the user type and check if missing
            string? str = Environment.GetEnvironmentVariable(userType);
            if (string.IsNullOrEmpty(str))
            {
                throw new Exception($"Missing environment variable: {userType}");
            }

            // Check if the environment variable contains exactly two ':'
            // The first part is the username, second part is the password, third is the mailbox
            string separator = ":";
            string[] parts = str.Split(separator);

            if (parts.Length != 3)
            {
                throw new Exception(
                $"Environment variable {userType} must contain exactly two ':' characters, separating username, password, and mailbox!"
                );
            }

            string username = parts[0];
            string password = parts[1];
            string mailbox = parts[2];

            // Return the username, password, and mailbox as a tuple
            return (username, password, mailbox);
            
        }
    }
}
