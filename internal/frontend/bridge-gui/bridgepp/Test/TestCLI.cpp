// Copyright (c) 2023 Proton AG
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
#include <gtest/gtest.h>


using namespace bridgepp;



//****************************************************************************************************************************************************
//
//****************************************************************************************************************************************************
TEST(CLI, stripStringParameterFromCommandLine) {
    struct Test {
        QStringList input;
        QStringList expectedOutput;
    };
    QList<Test> const tests = {
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

    for (Test const& test: tests) {
        EXPECT_EQ(stripStringParameterFromCommandLine("--string", test.input), test.expectedOutput);
    }
}
