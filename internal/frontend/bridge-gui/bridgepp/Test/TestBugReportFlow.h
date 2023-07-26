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


#ifndef BRIDGEPP_TESTBUGREPORTFLOW_H
#define BRIDGEPP_TESTBUGREPORTFLOW_H


#include <bridgepp/BugReportFlow/BugReportFlow.h>
#include <gtest/gtest.h>


//****************************************************************************************************************************************************
/// \brief Fixture class for BugReportFlow tests.
//****************************************************************************************************************************************************
class BugReportFlowFixture : public testing::Test {
public: // member functions.
    BugReportFlowFixture(); ///< Default constructor.
    BugReportFlowFixture(BugReportFlowFixture const &) = delete; ///< Disabled copy-constructor.
    BugReportFlowFixture(BugReportFlowFixture &&) = delete; ///< Disabled assignment copy-constructor.
    ~BugReportFlowFixture() = default; ///< Destructor.
    BugReportFlowFixture &operator=(BugReportFlowFixture const &) = delete; ///< Disabled assignment operator.
    BugReportFlowFixture &operator=(BugReportFlowFixture &&) = delete; ///< Disabled move assignment operator.

protected: // member functions.
    void SetUp() override; ///< Setup the fixture.
    void TearDown() override; ///< Tear down the fixture.
    void feedTempFile(const QString& json); ///< Feed the temp file with raw JSON.

protected: // data members
    bridgepp::BugReportFlow flow_; ///< The BugReportFlow.
    QTemporaryFile file_; ///< The file to be feed and parsed.
};

#endif //BRIDGEPP_TESTBUGREPORTFLOW_H
