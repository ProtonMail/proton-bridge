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


#include <gtest/gtest.h>
#include <bridgepp/BridgeUtils.h>


using namespace bridgepp;


//****************************************************************************************************************************************************
//
//****************************************************************************************************************************************************
TEST(BridgeUtils, OS) {
#ifdef Q_OS_MACOS
    EXPECT_EQ(os(), OS::MacOS);
    EXPECT_FALSE(onLinux());
    EXPECT_TRUE(onMacOS());
    EXPECT_FALSE(onWindows());
    EXPECT_EQ(goos(), "darwin");
    return;
#endif

#ifdef Q_OS_WIN
    EXPECT_EQ(os(), OS::Windows);
    EXPECT_FALSE(onLinux());
    EXPECT_FALSE(onMacOS());
    EXPECT_TRUE(onWindows());
    EXPECT_EQ(goos(), "windows");
    return;
#endif

#ifdef Q_OS_LINUX
    EXPECT_EQ(os(), OS::Linux);
    EXPECT_TRUE(onLinux());
    EXPECT_FALSE(onMacOS());
    EXPECT_FALSE(onWindows());
    EXPECT_EQ(goos(), "linux");
    return;
#endif

    EXPECT_TRUE(false); // should be unreachable.
}

//****************************************************************************************************************************************************
//
//****************************************************************************************************************************************************
TEST(BridgeUtils, UserFolders) {
    typedef QString (*dirFunction)();
    QList<dirFunction> functions = { userConfigDir, userCacheDir, userDataDir, sentryCacheDir };
    QString path;
    for (dirFunction f: functions) {
        EXPECT_NO_THROW(path = f());
        EXPECT_FALSE(path.isEmpty());
        EXPECT_TRUE(QDir(path).exists());
    }
}


//****************************************************************************************************************************************************
//
//****************************************************************************************************************************************************
TEST(BridgeUtils, Random) {
    qint32 repeatCount = 1000;
    qint32 const maxValue = 5;
    for (qint32 i = 0; i < repeatCount; ++i) {
        qint64 n = 0;
        EXPECT_NO_THROW(n = randN(maxValue));
        EXPECT_TRUE((n >= 0) && (n < maxValue));
        QString name;
        EXPECT_NO_THROW(name = randomFirstName());
        EXPECT_FALSE(name.isEmpty());
        EXPECT_NO_THROW(name = randomLastName());
        EXPECT_FALSE(name.isEmpty());
        EXPECT_NO_THROW(randomUser());
    }
}


//****************************************************************************************************************************************************
//
//****************************************************************************************************************************************************
TEST(BridgeUtils, ElideLongString) {
    std::function const test = [](QString const &input, qint32 maxLength, QString const &expected) -> bool {
        QString output;
        EXPECT_NO_THROW(output = elideLongString(input, maxLength));
        return output == expected;
    };
    
    EXPECT_TRUE(test( "", 0, ""));
    EXPECT_TRUE(test("1234", 4, "1234"));
    EXPECT_TRUE(test("123", 2, "..."));
    EXPECT_TRUE(test("1234567890", 8, "12...90"));
    EXPECT_TRUE(test("1234567890", 10, "1234567890"));
    EXPECT_TRUE(test("1234567890", 100, "1234567890"));
}
