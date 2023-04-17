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

#include <bridgepp/Exception/Exception.h>
#include <gtest/gtest.h>


using namespace bridgepp;


namespace {
    QString const testQWhat = "What";
    QString const testDetails = "Some details";
    QString const testFunction = "function";
    QByteArray const testAttachment = QString("Some data").toLocal8Bit();
    Exception const testException(testQWhat, testDetails, testFunction, testAttachment);
}


//****************************************************************************************************************************************************
//
//****************************************************************************************************************************************************
TEST(Exceptions, ExceptionConstructor) {
    // Default exception
    Exception const emptyException;
    EXPECT_TRUE(emptyException.qwhat().isEmpty());
    EXPECT_EQ(strlen(emptyException.what()), 0);
    EXPECT_EQ(emptyException.attachment().size(), 0);
    EXPECT_TRUE(emptyException.details().isEmpty());
    EXPECT_TRUE(emptyException.detailedWhat().isEmpty());

    // Fully detailed exception
    EXPECT_EQ(testException.qwhat(), testQWhat);
    EXPECT_EQ(QString::fromLocal8Bit(testException.what()), testQWhat);
    EXPECT_EQ(testException.details(), testDetails);
    EXPECT_EQ(testException.attachment(), testAttachment);
    QString const detailed = testException.detailedWhat();
    EXPECT_TRUE(detailed.contains(testQWhat));
    EXPECT_TRUE(detailed.contains(testFunction));
    EXPECT_TRUE(detailed.contains(testDetails));
}


//****************************************************************************************************************************************************
//
//****************************************************************************************************************************************************
TEST(Exceptions, ExceptionCopyMoveConstructors) {
    Exception const e(testQWhat, testDetails, testFunction, testAttachment);

    // Check copy-constructor
    Exception eCopied(e);
    EXPECT_EQ(eCopied.qwhat(), testQWhat);
    EXPECT_EQ(eCopied.details(), testDetails);
    EXPECT_EQ(eCopied.function(), testFunction);
    EXPECT_EQ(eCopied.attachment(), testAttachment);

    // Check move-constructor
    Exception eMoved(std::move(eCopied));
    EXPECT_EQ(eMoved.qwhat(), testQWhat);
    EXPECT_EQ(eMoved.details(), testDetails);
    EXPECT_EQ(eMoved.function(), testFunction);
    EXPECT_EQ(eMoved.attachment(), testAttachment);
}


//****************************************************************************************************************************************************
//
//****************************************************************************************************************************************************
TEST(Exceptions, ExceptionThrow) {
    std::function t = []() { throw testException; };
    EXPECT_THROW(t(), Exception);
    EXPECT_THROW(t(), std::exception);
    bool caught = false;
    try {
        t();
    } catch (Exception const &e) {
        caught = true;
        EXPECT_EQ(e.detailedWhat(), testException.detailedWhat());
    }
    EXPECT_TRUE(caught);
}
