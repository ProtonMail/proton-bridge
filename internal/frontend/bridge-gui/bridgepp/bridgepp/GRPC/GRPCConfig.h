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


#ifndef BRIDGE_PP_GRPC_CONFIG_H
#define BRIDGE_PP_GRPC_CONFIG_H


//****************************************************************************************************************************************************
/// Service configuration class.
//****************************************************************************************************************************************************
struct GRPCConfig {
public: // data members
    qint32 port; ///< The port.
    QString cert; ///< The server TLS certificate.
    QString token; ///< The identification token.
    QString fileSocketPath; ///< The path of the file socket.

    bool load(QString const &path, QString *outError = nullptr); ///< Load the service config from file
    bool save(QString const &path, QString *outError = nullptr); ///< Save the service config to file
};


#endif //BRIDGE_PP_GRPC_CONFIG_H
