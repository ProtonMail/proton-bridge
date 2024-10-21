using System;
using System.Linq;
using System.IO;

namespace ProtonMailBridge.UI.Tests.TestsHelper
{
    public static class TestData
    {
        public static TimeSpan FiveSecondsTimeout => TimeSpan.FromSeconds(5);
        public static TimeSpan TenSecondsTimeout => TimeSpan.FromSeconds(10);
        public static TimeSpan ThirtySecondsTimeout => TimeSpan.FromSeconds(30);
        public static TimeSpan OneMinuteTimeout => TimeSpan.FromSeconds(60);
        public static TimeSpan RetryInterval => TimeSpan.FromMilliseconds(1000);
        public static string AppExecutable => "C:\\Program Files\\Proton AG\\Proton Mail Bridge\\proton-bridge.exe";
    }
}
