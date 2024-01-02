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


#include "GRPCConfig.h"
#include "../Exception/Exception.h"


using namespace bridgepp;


namespace {

Exception const invalidFileException("The content of the service configuration file is invalid"); // Exception for invalid config.
QString const keyPort = "port"; ///< The JSON key for the port.
QString const keyCert = "cert"; ///< The JSON key for the TLS certificate.
QString const keyToken = "token"; ///< The JSON key for the identification token.
QString const keyFileSocketPath = "fileSocketPath"; ///< The JSON key for the file socket path.


//****************************************************************************************************************************************************
/// \brief read a string value from a JSON object.
///
/// This function throws an Exception in case of error
///
/// \param[in] object The JSON object containing the value.
/// \param[in] key The key under which the value is stored.
//****************************************************************************************************************************************************
QString jsonStringValue(QJsonObject const &object, QString const &key) {
    QJsonValue const v = object[key];
    if (!v.isString()) {
        throw invalidFileException;
    }
    return v.toString();
}


//****************************************************************************************************************************************************
/// \brief read a string value from a JSON object.
///
/// This function throws an Exception in case of error.
///
/// \param[in] object The JSON object containing the value.
/// \param[in] key The key under which the value is stored.
//****************************************************************************************************************************************************
qint32 jsonIntValue(QJsonObject const &object, QString const &key) {
    QJsonValue const v = object[key];
    if (!v.isDouble()) {
        throw invalidFileException;
    }
    return v.toInt();
}


} // anonymous namespace


//****************************************************************************************************************************************************
/// \param[in] path The path of the file to load from.
/// \param[out] outError if not null and an error occurs, this variable contains a description of the error.
/// \return true iff the operation was successful.
//****************************************************************************************************************************************************
bool GRPCConfig::load(QString const &path, QString *outError) {
    try {
        QFile file(path);
        if (!file.exists())
            throw Exception("The gRPC service configuration file does not exist.");

        if (!file.open(QIODevice::ReadOnly | QIODevice::Text)) {
            QThread::msleep(500); // we wait a bit and retry once, just in case server is not done writing/moving the config file.
            if (!file.open(QIODevice::ReadOnly | QIODevice::Text)) {
                throw Exception("The gRPC service configuration file exists but cannot be opened.");
            }
        }

        QJsonDocument const doc = QJsonDocument::fromJson(file.readAll());
        QJsonObject const object = doc.object();
        port = jsonIntValue(object, keyPort);
        cert = jsonStringValue(object, keyCert);
        token = jsonStringValue(object, keyToken);
        fileSocketPath = jsonStringValue(object, keyFileSocketPath);

        return true;
    }
    catch (Exception const &e) {
        if (outError) {
            *outError = QString("Error loading gRPC service configuration file '%1'.\n%2").arg(QFileInfo(path).absoluteFilePath(), e.qwhat());
        }
        return false;
    }
}


//****************************************************************************************************************************************************
/// \param[in] path The path of the file to write to.
/// \param[out] outError if not null and an error occurs, this variable contains a description of the error.
/// \return true iff the operation was successful.
//****************************************************************************************************************************************************
bool GRPCConfig::save(QString const &path, QString *outError) {
    try {
        QJsonObject object;
        object.insert(keyPort, port);
        object.insert(keyCert, cert);
        object.insert(keyToken, token);
        object.insert(keyFileSocketPath, fileSocketPath);

        QFile file(path);
        if (!file.open(QIODevice::WriteOnly | QIODevice::Text)) {
            throw Exception("The file could not be opened for writing.");
        }

        QByteArray const array = QJsonDocument(object).toJson();
        if (array.size() != file.write(array)) {
            throw Exception("An error occurred while writing to the file.");
        }

        return true;
    }
    catch (Exception const &e) {
        if (outError) {
            *outError = QString("Error saving gRPC service configuration file '%1'.\n%2").arg(QFileInfo(path).absoluteFilePath(), e.qwhat());
        }
        return false;
    }
}
