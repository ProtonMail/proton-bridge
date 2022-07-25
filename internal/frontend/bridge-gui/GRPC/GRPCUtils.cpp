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


#include "Pch.h"
#include "Exception.h"
#include "GRPCUtils.h"


//****************************************************************************************************************************************************
/// \param[in] status The status
/// \param[in] callName The call name.
//****************************************************************************************************************************************************
void logGRPCCallStatus(grpc::Status const& status, QString const &callName)
{
    if (status.ok())
        app().log().debug(QString("%1()").arg(callName));
    else
        app().log().error(QString("%1() FAILED").arg(callName));
}


//****************************************************************************************************************************************************
/// \param[in] grpcUser the gRPC user struct
/// \return a user.
//****************************************************************************************************************************************************
SPUser parsegrpcUser(grpc::User const &grpcUser)
{
    // As we want to use shared pointers here, we do not want to use the Qt ownership system, so we set parent to nil.
    // But: From https://doc.qt.io/qt-5/qtqml-cppintegration-data.html:
    // " When data is transferred from C++ to QML, the ownership of the data always remains with C++. The exception to this rule
    // is when a QObject is returned from an explicit C++ method call: in this case, the QML engine assumes ownership of the object. "
    // This is the case here, so we explicitely indicate that the object is owned by C++.
    SPUser user = std::make_shared<User>(nullptr);
    QQmlEngine::setObjectOwnership(user.get(), QQmlEngine::CppOwnership);
    user->setProperty("username", QString::fromStdString(grpcUser.username()));
    user->setProperty("avatarText", QString::fromStdString(grpcUser.avatartext()));
    user->setProperty("loggedIn", grpcUser.loggedin());
    user->setProperty("splitMode", grpcUser.splitmode());
    user->setProperty("setupGuideSeen", grpcUser.setupguideseen());
    user->setProperty("usedBytes", float(grpcUser.usedbytes()));
    user->setProperty("totalBytes", float(grpcUser.totalbytes()));
    user->setProperty("password", QString::fromStdString(grpcUser.password()));
    QStringList addresses;
    for (int j = 0; j < grpcUser.addresses_size(); ++j)
        addresses.append(QString::fromStdString(grpcUser.addresses(j)));
    user->setProperty("addresses", addresses);
    user->setProperty("id", QString::fromStdString(grpcUser.id()));
    return user;
}


//****************************************************************************************************************************************************
/// \param[in] level The log level
//****************************************************************************************************************************************************
grpc::LogLevel logLevelToGRPC(Log::Level level)
{
    switch (level)
    {
    case Log::Level::Panic: return grpc::LogLevel::PANIC;
    case Log::Level::Fatal: return grpc::LogLevel::FATAL;
    case Log::Level::Error: return grpc::LogLevel::ERROR;
    case Log::Level::Warn: return grpc::LogLevel::WARN;
    case Log::Level::Info: return grpc::LogLevel::INFO;
    case Log::Level::Debug: return grpc::LogLevel::DEBUG;
    case Log::Level::Trace: return grpc::LogLevel::TRACE;
    default:
        throw Exception(QString("unknown log level %1.").arg(qint32(level)));
    }
}
