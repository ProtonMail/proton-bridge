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

#include "BugReportFlow.h"


namespace {


    QString const currentFormatVersion = "1.0.0";


}


namespace bridgepp {


//****************************************************************************************************************************************************
//
//****************************************************************************************************************************************************
BugReportFlow::BugReportFlow() {
}


//****************************************************************************************************************************************************
/// \param[in] filepath The path of the file to parse.
/// \return True iff the file can be properly parsed.
//****************************************************************************************************************************************************
bool BugReportFlow::parse(const QString& filepath) {
    if (!QFile(filepath).exists())
        return false;

    this->filepath_ = filepath;
    return parseFile();
}


//****************************************************************************************************************************************************
/// \return The value for the 'bugCategories' property.
//****************************************************************************************************************************************************
    QStringList BugReportFlow::categories() const {
        return categories_;
    }


//****************************************************************************************************************************************************
/// \return The value for the 'bugQuestions' property.
//****************************************************************************************************************************************************
    QVariantList BugReportFlow::questions() const {
        return questions_;
    }


//****************************************************************************************************************************************************
/// \param[in] categoryId The id of the bug category.
/// \return Set of question for this category.
//****************************************************************************************************************************************************
QVariantList BugReportFlow::questionSet(quint8 categoryId) const {
    if (categoryId > questionsSet_.count() - 1)
        return QVariantList();
    return questionsSet_[categoryId];
};


//****************************************************************************************************************************************************
/// \param[in] questionId The id of the question.
/// \param[in] answer     The answer to that question.
/// \return true iff questionId match an existing question.
//****************************************************************************************************************************************************
bool BugReportFlow::setAnswer(quint8 questionId, QString const &answer) {
    if (questionId > questions_.count() - 1)
        return false;

    this->answers_[questionId] = answer;
    return true;
}


//****************************************************************************************************************************************************
/// \param[in] categoryId The id of the question set.
/// \return concatenate answers for set of questions.
//****************************************************************************************************************************************************
QString BugReportFlow::collectAnswers(quint8 categoryId) const {
    QString answers;
    if (categoryId > categories_.count() - 1)
        return answers;

    answers += "Category: " + categories_[categoryId] + "\n\r";
    QVariantList sets = this->questionSet(categoryId);
    for (QVariant const &var: sets) {
        const QString& answer = answers_[var.toInt()];
        if (answer.isEmpty())
            continue;
        answers += " - " + questions_[var.toInt()].toMap()["text"].toString() + "\n\r";
        answers += answer + "\n\r";
    }
    return answers;
}


//****************************************************************************************************************************************************
//
//****************************************************************************************************************************************************
void BugReportFlow::clearAnswers() {
    answers_.clear();
}


//****************************************************************************************************************************************************
//
//****************************************************************************************************************************************************
bool BugReportFlow::parseFile() {
    categories_.clear();
    questions_.clear();
    questionsSet_.clear();

    QJsonObject data = getJsonDataObj(getJsonRootObj());

    QJsonArray categoriesJson = data.value("categories").toArray();
    for (const QJsonValueRef &v : categoriesJson) {
        categories_.append(v.toObject()["name"].toString());
        questionsSet_.append(v.toObject()["questions"].toArray().toVariantList());
    }
    questions_ = data.value("questions").toArray().toVariantList();
    return true;
}

QJsonObject BugReportFlow::getJsonRootObj() {
    QFile file(filepath_);
    if (!file.open(QIODevice::ReadOnly | QIODevice::Text))
        return QJsonObject();

    const QString& val = file.readAll();
    file.close();
    QJsonDocument d = QJsonDocument::fromJson(val.toUtf8());
    return d.object();
}


QJsonObject BugReportFlow::getJsonDataObj(const QJsonObject& root) {
    QString version = getJsonVersion(root);
    if (version.isEmpty())
        return QJsonObject();

    QJsonValue data = root.value(QString("data_v%1").arg(version));
    if (data == QJsonValue::Undefined || !data.isObject())
        return QJsonObject();
    QJsonObject dataObj = data.toObject();

    return migrateData(dataObj, version);
}


QString BugReportFlow::getJsonVersion(const QJsonObject& root) {
    QJsonValue metadata = root.value("metadata");
    if (metadata == QJsonValue::Undefined || !metadata.isObject()) {
        return QString();
    }
    QJsonValue version = metadata.toObject().value("version");
    if (version == QJsonValue::Undefined || !version.isString()) {
        return QString();
    }
    return version.toString();
}


QJsonObject BugReportFlow::migrateData(const QJsonObject& data, const QString& version) {
    if (version != currentFormatVersion)
        return QJsonObject();
    // nothing to migrate now but migration should be done here.
    return data;
}

} // namespace bridgepp