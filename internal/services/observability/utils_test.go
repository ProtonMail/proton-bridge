// Copyright (c) 2024 Proton AG
//
// This file is part of Proton Mail Bridge.
//
// Proton Mail Bridge is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Proton Mail Bridge is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with Proton Mail Bridge.  If not, see <https://www.gnu.org/licenses/>.

package observability

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMatchUserAgent(t *testing.T) {
	type agentParseResult struct {
		agent  string
		result string
	}

	testCases := []agentParseResult{
		{
			agent:  "Microsoft Outlook/16.0.17928.20114 (Windows 11 Version 23H2)",
			result: emailAgentOutlook,
		},
		{
			agent:  "Mailbird/3.0.18.0 (Windows 11 Version 23H2)",
			result: emailAgentOther,
		},
		{
			agent:  "Microsoft Outlook/16.0.17830.20166 (Windows 11 Version 23H2)",
			result: emailAgentOutlook,
		},
		{
			agent:  "Mac OS X Mail/16.0-3776.700.51 (macOS 14.6)",
			result: emailAgentAppleMail,
		},
		{
			agent:  "/ (Windows 11 Version 23H2)",
			result: emailAgentOther,
		},
		{
			agent:  "Microsoft Outlook for Mac/16.88.0-BUILDDAY (macOS 14.6)",
			result: emailAgentOutlook,
		},
		{
			agent:  "/ (macOS 14.5)",
			result: emailAgentOther,
		},
		{
			agent:  "/ (Freedesktop SDK 23.08 (Flatpak runtime))",
			result: emailAgentOther,
		},
		{
			agent:  "Mac OS X Mail/16.0-3774.600.62 (macOS 14.5)",
			result: emailAgentAppleMail,
		},
		{
			agent:  "Mac OS X Notes/4.11-2817 (macOS 14.6)",
			result: emailAgentUnknown,
		},
		{
			agent:  "NoClient/0.0.1 (macOS 14.6)",
			result: emailAgentOther,
		},
		{
			agent:  "Thunderbird/115.15.0 (Ubuntu 20.04.6 LTS)",
			result: emailAgentThunderbird,
		},
		{
			agent:  "Thunderbird/115.14.0 (macOS 14.6)",
			result: emailAgentThunderbird,
		},
		{
			agent:  "Thunderbird/115.10.2 (Windows 11 Version 23H2)",
			result: emailAgentThunderbird,
		},
		{
			agent:  "Mac OS X Notes/4.9-1965 (macOS Monterey (12.0))",
			result: emailAgentUnknown,
		},
		{
			agent:  " Thunderbird/115.14.0 (macOS 14.6) ",
			result: emailAgentThunderbird,
		},
		{
			agent:  "",
			result: emailAgentUnknown,
		},
	}

	for _, testCase := range testCases {
		require.Equal(t, testCase.result, matchUserAgent(testCase.agent))
	}
}

func TestFormatBool(t *testing.T) {
	require.Equal(t, "false", formatBool(false))
	require.Equal(t, "true", formatBool(true))
}
