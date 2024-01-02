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


#include "GRPCUtils.h"
#include "GRPCConfig.h"
#include "../Exception/Exception.h"
#include "../BridgeUtils.h"


namespace bridgepp {


std::string const grpcMetadataServerTokenKey = "server-token";


namespace {


//****************************************************************************************************************************************************
/// \return the gRPC server config file name
//****************************************************************************************************************************************************
QString grpcServerConfigFilename() {
    return "grpcServerConfig.json";
}


//****************************************************************************************************************************************************
/// \return the gRPC client config file name
//****************************************************************************************************************************************************
QString grpcClientConfigBaseFilename() {
    return "grpcClientConfig_%1.json";
}


} // anonymous namespace


//****************************************************************************************************************************************************
/// return true if gRPC connection should use file socket instead of TCP socket.
//****************************************************************************************************************************************************
bool useFileSocketForGRPC() {
    return !onWindows();
}


//****************************************************************************************************************************************************
/// \param[in] configDir The folder containing the configuration files.
/// \return The absolute path of the service config path.
//****************************************************************************************************************************************************
QString grpcServerConfigPath(QString const &configDir) {
    return QDir(configDir).absoluteFilePath(grpcServerConfigFilename());
}


//****************************************************************************************************************************************************
/// \return The absolute path of the service config path.
//****************************************************************************************************************************************************
QString grpcClientConfigBasePath(QString const &configDir) {
    return QDir(configDir).absoluteFilePath(grpcClientConfigBaseFilename());
}


//****************************************************************************************************************************************************
/// \param[in] token The token to put in the file.
/// \param[out] outError if the function returns an empty string and this pointer is not null, the pointer variable holds a description of the error
/// on exit.
/// \return The path of the created file.
/// \return A null string if the file could not be saved.
//****************************************************************************************************************************************************
QString createClientConfigFile(QString const &configDir, QString const &token, QString *outError) {
    QString const basePath = grpcClientConfigBasePath(configDir);
    QString path, error;
    for (qint32 i = 0; i < 1000; ++i) // we try a decent amount of times
    {
        path = basePath.arg(i);
        if (!QFileInfo(path).exists()) {
            GRPCConfig config;
            config.token = token;

            if (!config.save(path, outError)) {
                return QString();
            }
            return path;
        }
    }

    if (outError)
        *outError = "no usable client configuration file name could be found.";
    return QString();
}


//****************************************************************************************************************************************************
/// \param[in] level The Log::Level.
/// \return The grpc::LogLevel.
//****************************************************************************************************************************************************
grpc::LogLevel logLevelToGRPC(Log::Level level) {
    switch (level) {
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
Log::Level logLevelFromGRPC(grpc::LogLevel level) {
    switch (level) {
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
/// \param[in] state The user state.
/// \return The  gRPC user state.
//****************************************************************************************************************************************************
grpc::UserState userStateToGRPC(UserState state) {
    switch (state) {
    case UserState::SignedOut:
        return grpc::UserState::SIGNED_OUT;
    case UserState::Locked:
        return grpc::UserState::LOCKED;
    case UserState::Connected:
        return grpc::UserState::CONNECTED;
    default:
        throw Exception(QString("unknown gRPC user state %1.").arg(qint32(state)));
    }
}


//****************************************************************************************************************************************************
/// \param[in] state The gRPC user state
/// \return the user state
//****************************************************************************************************************************************************
UserState userStateFromGRPC(grpc::UserState state) {
    switch (state) {
    case grpc::UserState::SIGNED_OUT:
        return UserState::SignedOut;
    case grpc::UserState::LOCKED:
        return UserState::Locked;
    case grpc::UserState::CONNECTED:
        return UserState::Connected;
    default:
        throw Exception(QString("unknown gRPC user state %1.").arg(qint32(state)));
    }
}


//****************************************************************************************************************************************************
/// \param[in] grpcUser The gRPC user.
/// \return user The user.
//****************************************************************************************************************************************************
SPUser userFromGRPC(grpc::User const &grpcUser) {
    SPUser user = User::newUser(nullptr);

    user->setID(QString::fromStdString(grpcUser.id()));
    user->setUsername(QString::fromStdString(grpcUser.username()));
    user->setPassword(QString::fromStdString(grpcUser.password()));
    QStringList addresses;
    for (int j = 0; j < grpcUser.addresses_size(); ++j) {
        addresses.append(QString::fromStdString(grpcUser.addresses(j)));
    }
    user->setAddresses(addresses);
    user->setAvatarText(QString::fromStdString(grpcUser.avatartext()));
    user->setState(userStateFromGRPC(grpcUser.state()));
    user->setSplitMode(grpcUser.splitmode());
    user->setUsedBytes(float(grpcUser.usedbytes()));
    user->setTotalBytes(float(grpcUser.totalbytes()));

    return user;
}


//****************************************************************************************************************************************************
/// \param[in] user the user.
/// \param[out] outGRPCUser The GRPC user.
//****************************************************************************************************************************************************
void userToGRPC(User const &user, grpc::User &outGRPCUser) {
    outGRPCUser.set_id(user.id().toStdString());
    outGRPCUser.set_username(user.username().toStdString());
    outGRPCUser.set_password(user.password().toStdString());
    outGRPCUser.clear_addresses();
    for (QString const &address: user.addresses()) {
        outGRPCUser.add_addresses(address.toStdString());
    }
    outGRPCUser.set_avatartext(user.avatarText().toStdString());
    outGRPCUser.set_state(userStateToGRPC(user.state()));
    outGRPCUser.set_splitmode(user.splitMode());
    outGRPCUser.set_usedbytes(qint64(user.usedBytes()));
    outGRPCUser.set_totalbytes(qint64(user.totalBytes()));
}


//****************************************************************************************************************************************************
/// \return The path to a file that can be used for a gRPC file socket.
/// \return A null string if no path could be found.
//****************************************************************************************************************************************************
QString getAvailableFileSocketPath() {
    QDir const tempDir(QStandardPaths::writableLocation(QStandardPaths::TempLocation));
    for (qint32 i = 0; i < 10000; i++) {
        QString const path = tempDir.absoluteFilePath(QString("bridge_%1.sock").arg(qint32(i), 4, 10, QChar('0')));
        QFile f(path);
        if ((!f.exists()) || f.remove()) {
            return path;
        }
    }

    return QString();
}


} // namespace bridgepp
