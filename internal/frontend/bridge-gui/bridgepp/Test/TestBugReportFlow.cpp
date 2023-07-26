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

#include "TestBugReportFlow.h"
#include <bridgepp/BugReportFlow/BugReportFlow.h>


using namespace bridgepp;


//****************************************************************************************************************************************************
//
//****************************************************************************************************************************************************
BugReportFlowFixture::BugReportFlowFixture()
        : testing::Test()
        , flow_()
        , file_(){
}


//****************************************************************************************************************************************************
//
//****************************************************************************************************************************************************
void BugReportFlowFixture::SetUp() {
    Test::SetUp();
}


//****************************************************************************************************************************************************
//
//****************************************************************************************************************************************************
void BugReportFlowFixture::TearDown() {
    Test::TearDown();
}


//****************************************************************************************************************************************************
//
//****************************************************************************************************************************************************
void BugReportFlowFixture::feedTempFile(const QString& json) {
    QJsonDocument doc = QJsonDocument().fromJson(json.toUtf8());
    file_.open();
    file_.write(doc.toJson());
    file_.close();
}


//****************************************************************************************************************************************************
//
//****************************************************************************************************************************************************
TEST_F(BugReportFlowFixture, noFile) {
    EXPECT_FALSE(flow_.parse(""));
    EXPECT_EQ(flow_.categories(), QStringList());
    EXPECT_EQ(flow_.questions(), QVariantList());
}


//****************************************************************************************************************************************************
//
//****************************************************************************************************************************************************
TEST_F(BugReportFlowFixture, emptyFile) {
    feedTempFile("");
    EXPECT_TRUE(flow_.parse(file_.fileName()));
    EXPECT_EQ(flow_.categories(), QStringList());
    EXPECT_EQ(flow_.questions(), QVariantList());
}


//****************************************************************************************************************************************************
//
//****************************************************************************************************************************************************
TEST_F(BugReportFlowFixture, validFile) {
    feedTempFile("{"
                 "  \"metadata\": {"
                 "    \"version\": \"1.0.0\""
                 "  },"
                 "  \"data_v1.0.0\": {"
                 "    \"categories\": ["
                 "      {"
                 "        \"id\": 0,"
                 "        \"name\": \"I can't receive mail\","
                 "        \"questions\": [0]"
                 "      }"
                 "    ],"
                 "    \"questions\": ["
                 "      {"
                 "        \"id\": 0,"
                 "        \"text\": \"What happened?\","
                 "        \"tips\": \"Expected behavior\","
                 "        \"type\": 1"
                 "      }"
                 "    ]"
                 "  }"
                 "}");

    EXPECT_TRUE(flow_.parse(file_.fileName()));
    QStringList categories = flow_.categories();
    QVariantList questions = flow_.questions();
    EXPECT_EQ(categories.count(), 1);
    EXPECT_EQ(categories[0], "I can't receive mail");
    EXPECT_EQ(questions.count(), 1);
    QVariantMap q1 = questions[0].toMap();
    EXPECT_EQ(q1.value("id").toInt(), 0);
    EXPECT_EQ(q1.value("text").toString(), "What happened?");
    EXPECT_EQ(q1.value("tips").toString(), "Expected behavior");
    EXPECT_EQ(q1.value("type").toInt(), 1);

    QVariantList questionSet = flow_.questionSet(0);
    EXPECT_EQ(questionSet.count(), 1);
    EXPECT_EQ(questionSet[0].toInt(), 0);

    QVariantList questionSetBad = flow_.questionSet(1);
    EXPECT_EQ(questionSetBad.count(), 0);

    EXPECT_TRUE(flow_.setAnswer(0, "pwet"));
    EXPECT_FALSE(flow_.setAnswer(1, "pwet"));
    qDebug() << flow_.collectAnswers(0);
    EXPECT_EQ(flow_.collectAnswers(0), " - What happened? pwet\n\r");
    EXPECT_EQ(flow_.collectAnswers(1), "");
    flow_.clearAnswers();
    EXPECT_EQ(flow_.collectAnswers(0), " - What happened? \n\r");
}


//****************************************************************************************************************************************************
//
//****************************************************************************************************************************************************
TEST_F(BugReportFlowFixture, badVersionFile) {
    feedTempFile("{"
                 "  \"metadata\": {"
                 "    \"version\": \"1.0.1\""
                 "  },"
                 "  \"data_v1.0.1\": {"
                 "    \"categories\": ["
                 "      {"
                 "        \"id\": 0,"
                 "        \"name\": \"I can't receive mail\","
                 "        \"questions\": [0]"
                 "      }"
                 "    ],"
                 "    \"questions\": ["
                 "      {"
                 "        \"id\": 0,"
                 "        \"text\": \"What happened?\","
                 "        \"tips\": \"Expected behavior\","
                 "        \"type\": 1"
                 "      }"
                 "    ]"
                 "  }"
                 "}");

    EXPECT_TRUE(flow_.parse(file_.fileName()));
    EXPECT_EQ(flow_.categories(), QStringList());
    EXPECT_EQ(flow_.questions(), QVariantList());
}