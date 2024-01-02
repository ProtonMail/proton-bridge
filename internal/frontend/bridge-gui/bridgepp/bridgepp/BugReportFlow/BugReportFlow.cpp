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

#include "BugReportFlow.h"


namespace {


    QString const currentFormatVersion = "1.0.0";


}


namespace bridgepp {


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
QVariantList BugReportFlow::categories() const {
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
/// \param[in] categoryId The id of the question.
/// \return answer the given question.
//****************************************************************************************************************************************************
QString BugReportFlow::getCategory(quint8 categoryId) const {
    QString category;
    if (categoryId <= categories_.count() - 1) {
        category = categories_[categoryId].toMap()["name"].toString();
    }
    return category;
}


//****************************************************************************************************************************************************
/// \param[in] questionId The id of the question.
/// \return answer the given question.
//****************************************************************************************************************************************************
QString BugReportFlow::getAnswer(quint8 questionId) const {
    QString answer;
    if (questionId <= questions_.count() - 1) {
        answer = answers_[questionId];
    }
    return answer;
}


//****************************************************************************************************************************************************
/// \param[in] categoryId The id of the question set.
/// \return concatenate answers for set of questions.
//****************************************************************************************************************************************************
QString BugReportFlow::collectAnswers(quint8 categoryId) const {
    QString answers;
    if (categoryId > categories_.count() - 1)
        return answers;

    QVariantList sets = this->questionSet(categoryId);
    for (QVariant const &var: sets) {
        const QString answer = getAnswer(var.toInt());
        if (answer.isEmpty())
            continue;
        answers += "#### " + questions_[var.toInt()].toMap()["text"].toString() + "\n\r";
        for (const QString& line : answer.split("\n"))
            answers += "> " + line + "\n\r";
    }
    return answers;
}


//****************************************************************************************************************************************************
/// \param[in] categoryId The id of the question set.
//****************************************************************************************************************************************************
QString BugReportFlow::collectUserInput(quint8 categoryId) const {
    if (categoryId > categories_.count() - 1)
        return {};

    QString input = this->getCategory(categoryId);
    for (QVariant const &var: this->questionSet(categoryId)) {
        QString const answer = getAnswer(var.toInt());
        if (!answer.isEmpty())
            input += " " + answer;
    }

    return input;
}


//****************************************************************************************************************************************************
//
//****************************************************************************************************************************************************
void BugReportFlow::clearAnswers() {
    answers_.clear();
}


//****************************************************************************************************************************************************
/// \return true iff parsing succeed.
//****************************************************************************************************************************************************
bool BugReportFlow::parseFile() {
    reset();

    QJsonObject data = getJsonDataObj(getJsonRootObj());
    QJsonArray categoriesJson = data.value("categories").toArray();
    for (const QJsonValueRef &v : categoriesJson) {
        QVariantMap cat;
        cat["name"] = v.toObject()["name"].toString();
        cat["hint"] = v.toObject()["hint"].toString();
        categories_.append(cat);
        questionsSet_.append(v.toObject()["questions"].toArray().toVariantList());
    }
    questions_ = data.value("questions").toArray().toVariantList();
    return true;
}


//****************************************************************************************************************************************************
//
//****************************************************************************************************************************************************
void BugReportFlow::reset() {
    categories_.clear();
    questions_.clear();
    questionsSet_.clear();
    answers_.clear();
}


//****************************************************************************************************************************************************
//
//****************************************************************************************************************************************************
QJsonObject BugReportFlow::getJsonRootObj() const {
    QFile file(filepath_);
    if (!file.open(QIODevice::ReadOnly | QIODevice::Text))
        return QJsonObject();

    const QString& val = file.readAll();
    file.close();
    QJsonDocument d = QJsonDocument::fromJson(val.toUtf8());
    return d.object();
}


//****************************************************************************************************************************************************
//
//****************************************************************************************************************************************************
QJsonObject BugReportFlow::getJsonDataObj(const QJsonObject& root) const {
    QString version = getJsonVersion(root);
    if (version.isEmpty())
        return QJsonObject();

    QJsonValue data = root.value(QString("data_v%1").arg(version));
    if (data == QJsonValue::Undefined || !data.isObject())
        return QJsonObject();

    return migrateData(data.toObject(), version);
}


//****************************************************************************************************************************************************
//
//****************************************************************************************************************************************************
QString BugReportFlow::getJsonVersion(const QJsonObject& root) const{
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


//****************************************************************************************************************************************************
//
//****************************************************************************************************************************************************
QJsonObject BugReportFlow::migrateData(const QJsonObject& data, const QString& version) const{
    if (version != currentFormatVersion)
        return QJsonObject();

    // nothing to migrate now but migration should be done here.
    return data;
}

} // namespace bridgepp