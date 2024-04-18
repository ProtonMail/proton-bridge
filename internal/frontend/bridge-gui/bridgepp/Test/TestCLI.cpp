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
// along with Proton Mail Bridge. If not, see <https://www.gnu.org/licenses/>.


#include <bridgepp/CLI/CLIUtils.h>
#include <bridgepp/SessionID/SessionID.h>
#include <gtest/gtest.h>


using namespace bridgepp;


//****************************************************************************************************************************************************
//
//****************************************************************************************************************************************************
TEST(CLI, stripStringParameterFromCommandLine) {
    struct TestData {
        QStringList input;
        QStringList expectedOutput;
    };
    QList<TestData> const tests = {
        {{}, {}},
        {{ "--a", "-b", "--C" }, { "--a", "-b", "--C" } },
        {{ "--string", "value" }, {} },
        {{ "--string" }, {} },
        {{ "--string", "value", "-b", "--C" }, { "-b", "--C" } },
        {{ "--string", "value", "-b", "--string", "value", "--C" }, { "-b", "--C" } },
        {{ "--string", "value", "-b", "--string", "value", "--C" }, { "-b", "--C" } },
        {{ "--string", "value", "-b", "--string"}, { "-b" } },
        {{ "--string", "--string", "value", "-b", "--string"}, { "value", "-b" } },
    };

    for (TestData const& test: tests) {
        EXPECT_EQ(stripStringParameterFromCommandLine("--string", test.input), test.expectedOutput);
    }
}


TEST(CLI, parseGoCLIStringArgument) {
    struct TestData {
        QStringList args;
        QStringList params;
        QStringList expectedOutput;
    };

    QList<TestData> const tests = {
        { {}, {}, {} },
        { {"-param"}, {"param"}, {} },
        { {"--param", "1"}, {"param"}, { "1" } },
        { {"--param", "1","p", "-p", "2", "-flag", "-param=3", "--p=4"}, {"param", "p"}, { "1", "2", "3", "4" } },
        { {"--param", "--param", "1"}, {"param"}, { "--param" } },
    };

    for (TestData const& test: tests) {
        EXPECT_EQ(parseGoCLIStringArgument(test.args, test.params), test.expectedOutput);
    }
}

TEST(CLI, cliArgsToStringList) {
    int constexpr argc = 3;
    char *argv[] = { const_cast<char *>("1"), const_cast<char *>("2"), const_cast<char *>("3") };
    QStringList const strList { "1", "2", "3" };
    EXPECT_EQ(cliArgsToStringList(argc,argv), strList);
    EXPECT_EQ(cliArgsToStringList(0, nullptr), QStringList {});
}

TEST(CLI, mostRecentSessionID) {
    QStringList const sessionIDs { "20220411_155931148", "20230411_155931148", "20240411_155931148" };
    EXPECT_EQ(mostRecentSessionID({ hyphenatedSessionIDFlag, sessionIDs[0] }), sessionIDs[0]);
    EXPECT_EQ(mostRecentSessionID({ hyphenatedSessionIDFlag, sessionIDs[1], hyphenatedSessionIDFlag, sessionIDs[2] }), sessionIDs[2]);
    EXPECT_EQ(mostRecentSessionID({ hyphenatedSessionIDFlag, sessionIDs[2], hyphenatedSessionIDFlag, sessionIDs[1] }), sessionIDs[2]);
    EXPECT_EQ(mostRecentSessionID({ hyphenatedSessionIDFlag, sessionIDs[1], hyphenatedSessionIDFlag, sessionIDs[2], hyphenatedSessionIDFlag,
        sessionIDs[0] }), sessionIDs[2]);
}
