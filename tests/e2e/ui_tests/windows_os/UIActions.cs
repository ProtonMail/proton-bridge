using System;
using FlaUI.Core.AutomationElements;
using FlaUI.Core.Definitions;
using FlaUI.Core.Input;
using FlaUI.Core.Tools;
using NUnit.Framework;

namespace ProtonMailBridge.UI.Tests
{
    public class UIActions : TestSession
    {
        public AutomationElement AccountView => Window.FindFirstDescendant(cf => cf.ByControlType(ControlType.Pane));
    }
}