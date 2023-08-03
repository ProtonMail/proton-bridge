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


#ifndef BRIDGE_GUI_BUG_REPORT_FLOW_H
#define BRIDGE_GUI_BUG_REPORT_FLOW_H

namespace bridgepp {


//****************************************************************************************************************************************************
/// \brief Bug Report Flow parser.
//****************************************************************************************************************************************************
class BugReportFlow {

public: // member functions.
    BugReportFlow(); ///< Default constructor.
    BugReportFlow(BugReportFlow const &) = delete; ///< Disabled copy-constructor.
    BugReportFlow(BugReportFlow &&) = delete; ///< Disabled assignment copy-constructor.
    ~BugReportFlow() = default; ///< Destructor.

    [[nodiscard]] bool parse(const QString& filepath); ///< Initialize the Bug Report Flow.

    [[nodiscard]] QVariantList categories() const; ///< Getter for the 'bugCategories' property.
    [[nodiscard]] QVariantList questions() const; ///< Getter for the 'bugQuestions' property.
    [[nodiscard]] QVariantList questionSet(quint8 categoryId) const; ///< Retrieve the set of question for a given bug category.
    [[nodiscard]] bool setAnswer(quint8 questionId, QString const &answer); ///< Feed an answer for a given question.
    [[nodiscard]] QString getCategory(quint8 categoryId) const;  ///< Get category name.
    [[nodiscard]] QString getAnswer(quint8 questionId) const; ///< Get answer for a given question.
    [[nodiscard]] QString collectAnswers(quint8 categoryId) const; ///< Collect answer for a given set of questions.
    void clearAnswers(); ///< Clear all collected answers.


private: // member functions
    bool parseFile(); ///< Parse the bug report flow description file.
    void reset(); ///< Reset all data.
    [[nodiscard]] QJsonObject getJsonRootObj() const; ///< Extract the JSON root object.
    [[nodiscard]] QJsonObject getJsonDataObj(const QJsonObject& root) const; ///< Extract the JSON data object.
    [[nodiscard]] QString getJsonVersion(const QJsonObject& root) const; ///< Extract the JSON version of the file.
    [[nodiscard]] QJsonObject migrateData(const QJsonObject& data, const QString& version) const; ///< Migrate data if needed/possible.

private: // data members
    QString filepath_; ///< The file path of the BugReportFlow description file.
    QVariantList categories_; ///< The list of Bug Category parsed from the description file.
    QVariantList questions_; ///< The list of Questions parsed from the description file.
    QList<QVariantList> questionsSet_; ///< Sets of questions per bug category.
    QMap<quint8, QString> answers_; ///< Map of QuestionId/Answer for the bug form.
};


} // namespace bridgepp

#endif // BRIDGE_GUI_BUG_REPORT_FLOW_H
