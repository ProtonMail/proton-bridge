// Copyright (c) 2022 Proton AG
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


namespace
{

Exception const invalidFileException("The service configuration file is invalid"); // Exception for invalid config.
Exception const couldNotSaveException("The service configuration file could not be saved"); ///< Exception for write errors.
QString const keyPort = "port"; ///< The JSON key for the port.
QString const keyCert = "cert"; ///< The JSON key for the TLS certificate.
QString const keyToken = "token"; ///< The JSON key for the identification token.


//****************************************************************************************************************************************************
/// \brief read a string value from a JSON object.
///
/// This function throws an Exception in case of error
///
/// \param[in] object The JSON object containing the value.
/// \param[in] key The key under which the value is stored.
//****************************************************************************************************************************************************
QString jsonStringValue(QJsonObject const &object, QString const &key)
{
    QJsonValue const v = object[key];
    if (!v.isString())
        throw invalidFileException;
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
qint32 jsonIntValue(QJsonObject const &object, QString const &key)
{
    QJsonValue const v = object[key];
    if (!v.isDouble())
        throw invalidFileException;
    return v.toInt();
}


} // anonymous namespace


//****************************************************************************************************************************************************
/// \param[in] path The path of the file to load from.
/// \param[out] outError if not null and an error occurs, this variable contains a description of the error.
/// \return true iff the operation was successful.
//****************************************************************************************************************************************************
bool GRPCConfig::load(QString const &path, QString *outError)
{
    try
    {
        QFile file(path);
        if (!file.open(QIODevice::ReadOnly | QIODevice::Text))
            throw Exception("Could not open gRPC service config file.");

        QJsonDocument const doc = QJsonDocument::fromJson(file.readAll());
        QJsonObject const object = doc.object();
        port = jsonIntValue(object, keyPort);
        cert = jsonStringValue(object, keyCert);
        token = jsonStringValue(object, keyToken);

        return true;
    }
    catch (Exception const &e)
    {
        if (outError)
            *outError = e.qwhat();
        return false;
    }
}


//****************************************************************************************************************************************************
/// \param[in] path The path of the file to write to.
/// \param[out] outError if not null and an error occurs, this variable contains a description of the error.
/// \return true iff the operation was successful.
//****************************************************************************************************************************************************
bool GRPCConfig::save(QString const &path, QString *outError)
{
    try
    {
        QJsonObject const object;
        object[keyPort] = port;
        object[keyCert] = cert;
        object[keyToken] = token;

        QFile file(path);
        if (!file.open(QIODevice::WriteOnly | QIODevice::Text))
            throw couldNotSaveException;

        QByteArray const array = QJsonDocument(object).toJson();
        if (array.size() != file.write(array))
            throw couldNotSaveException;

        return true;
    }
    catch (Exception const &e)
    {
        if (outError)
            *outError = e.qwhat();
        return false;
    }
}
