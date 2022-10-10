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


#include "GRPCUtils.h"
#include "GRPCConfig.h"
#include "../Exception/Exception.h"
#include "../BridgeUtils.h"


namespace bridgepp
{


std::string const grpcMetadataServerTokenKey = "server-token";


namespace
{


//****************************************************************************************************************************************************
/// \return the gRPC server config file name
//****************************************************************************************************************************************************
QString grpcServerConfigFilename()
{
    return "grpcServerConfig.json";
}


//****************************************************************************************************************************************************
/// \return the gRPC client config file name
//****************************************************************************************************************************************************
QString grpcClientConfigBaseFilename()
{
    return "grpcClientConfig_%1.json";
}


//****************************************************************************************************************************************************
/// \return The server certificate file name
//****************************************************************************************************************************************************
QString serverCertificateFilename()
{
    return "cert.pem";
}


//****************************************************************************************************************************************************
//
//****************************************************************************************************************************************************
QString serverKeyFilename()
{
    return "key.pem";
}


}


//****************************************************************************************************************************************************
/// \return The absolute path of the service config path.
//****************************************************************************************************************************************************
QString grpcServerConfigPath()
{
    return QDir(userConfigDir()).absoluteFilePath(grpcServerConfigFilename());
}


//****************************************************************************************************************************************************
/// \return The absolute path of the service config path.
//****************************************************************************************************************************************************
QString grpcClientConfigBasePath()
{
    return QDir(userConfigDir()).absoluteFilePath(grpcClientConfigBaseFilename());
}


//****************************************************************************************************************************************************
/// \return The absolute path of the server certificate.
//****************************************************************************************************************************************************
QString serverCertificatePath()
{
    return QDir(userConfigDir()).absoluteFilePath(serverCertificateFilename());
}


//****************************************************************************************************************************************************
/// \return The absolute path of the server key.
//****************************************************************************************************************************************************
QString serverKeyPath()
{

    return QDir(userConfigDir()).absoluteFilePath(serverKeyFilename());
}


//****************************************************************************************************************************************************
/// \param[in] token The token to put in the file.
/// \return The path of the created file.
/// \return A null string if the file could not be saved..
//****************************************************************************************************************************************************
QString createClientConfigFile(QString const &token)
{
    QString const basePath = grpcClientConfigBasePath();
    QString path, error;
    for (qint32 i = 0; i < 1000; ++i) // we try a decent amount of times
    {
        path = basePath.arg(i);
        if (!QFileInfo(path).exists())
        {
            GRPCConfig config;
            config.token = token;
            if (!config.save(path))
                return QString();
            return path;
        }
    }

    return QString();
}


//****************************************************************************************************************************************************
/// \param[in] level The Log::Level.
/// \return The grpc::LogLevel.
//****************************************************************************************************************************************************
grpc::LogLevel logLevelToGRPC(Log::Level level)
{
    switch (level)
    {
    case Log::Level::Panic:
        return grpc::LogLevel::LOG_PANIC;
    case Log::Level::Fatal:
        return grpc::LogLevel::LOG_FATAL;
    case Log::Level::Error:
        return grpc::LogLevel::LOG_ERROR;
    case Log::Level::Warn:
        return grpc::LogLevel::LOG_WARN;
    case Log::Level::Info:
        return grpc::LogLevel::LOG_INFO;
    case Log::Level::Debug:
        return grpc::LogLevel::LOG_DEBUG;
    case Log::Level::Trace:
        return grpc::LogLevel::LOG_TRACE;
    default:
        throw Exception(QString("unknown log level %1.").arg(qint32(level)));
    }
}


//****************************************************************************************************************************************************
/// \param[in] level The level::LogLevel.
/// \return The Log::Level.
//****************************************************************************************************************************************************
Log::Level logLevelFromGRPC(grpc::LogLevel level)
{
    switch (level)
    {
    case grpc::LOG_PANIC:
        return Log::Level::Panic;
    case grpc::LOG_FATAL:
        return Log::Level::Fatal;
    case grpc::LOG_ERROR:
        return Log::Level::Error;
    case grpc::LOG_WARN:
        return Log::Level::Warn;
    case grpc::LOG_INFO:
        return Log::Level::Info;
    case grpc::LOG_DEBUG:
        return Log::Level::Debug;
    case grpc::LOG_TRACE:
        return Log::Level::Trace;
    default:
        throw Exception(QString("unknown log level %1.").arg(qint32(level)));
    }
}


//****************************************************************************************************************************************************
/// \param[in] grpcUser The gRPC user.
/// \return user The user.
//****************************************************************************************************************************************************
SPUser userFromGRPC(grpc::User const &grpcUser)
{
    SPUser user = User::newUser(nullptr);

    user->setID(QString::fromStdString(grpcUser.id()));
    user->setUsername(QString::fromStdString(grpcUser.username()));
    user->setPassword(QString::fromStdString(grpcUser.password()));
    QStringList addresses;
    for (int j = 0; j < grpcUser.addresses_size(); ++j)
        addresses.append(QString::fromStdString(grpcUser.addresses(j)));
    user->setAddresses(addresses);
    user->setAvatarText(QString::fromStdString(grpcUser.avatartext()));
    user->setLoggedIn(grpcUser.loggedin());
    user->setSplitMode(grpcUser.splitmode());
    user->setSetupGuideSeen(grpcUser.setupguideseen());
    user->setUsedBytes(float(grpcUser.usedbytes()));
    user->setTotalBytes(float(grpcUser.totalbytes()));

    return user;
}


//****************************************************************************************************************************************************
/// \param[in] user the user.
/// \param[out] outGRPCUser The GRPC user.
//****************************************************************************************************************************************************
void userToGRPC(User const &user, grpc::User &outGRPCUser)
{
    outGRPCUser.set_id(user.id().toStdString());
    outGRPCUser.set_username(user.username().toStdString());
    outGRPCUser.set_password(user.password().toStdString());
    outGRPCUser.clear_addresses();
    for (QString const& address: user.addresses())
        outGRPCUser.add_addresses(address.toStdString());
    outGRPCUser.set_avatartext(user.avatarText().toStdString());
    outGRPCUser.set_loggedin(user.loggedIn());
    outGRPCUser.set_splitmode(user.splitMode());
    outGRPCUser.set_setupguideseen(user.setupGuideSeen());
    outGRPCUser.set_usedbytes(qint64(user.usedBytes()));
    outGRPCUser.set_totalbytes(qint64(user.totalBytes()));
}


} // namespace bridgepp