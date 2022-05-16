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


#ifndef BRIDGE_QT6_APP_CONTROLLER_H
#define BRIDGE_QT6_APP_CONTROLLER_H


class QMLBackend;
class GRPCClient;
class Log;


//****************************************************************************************************************************************************
/// \brief App controller class.
//****************************************************************************************************************************************************
class AppController: public QObject
{
    Q_OBJECT
    friend AppController& app();

public: // member functions.
    AppController(AppController const&) = delete; ///< Disabled copy-constructor.
    AppController(AppController&&) = delete; ///< Disabled assignment copy-constructor.
    ~AppController() override = default; ///< Destructor.
    AppController& operator=(AppController const&) = delete; ///< Disabled assignment operator.
    AppController& operator=(AppController&&) = delete; ///< Disabled move assignment operator.
    QMLBackend& backend() { return *backend_; } ///< Return a reference to the backend.
    GRPCClient& grpc() { return *grpc_; } ///< Return a reference to the GRPC client.
    Log& log() { return *log_; } ///< Return a reference to the log.

private: // member functions
    AppController(); ///< Default constructor.

private: // data members
    std::unique_ptr<QMLBackend> backend_; ///< The backend.
    std::unique_ptr<GRPCClient> grpc_; ///< The RPC client.
    std::unique_ptr<Log> log_; ///< The log.
};


AppController& app(); ///< Return a reference to the app controller.


#endif // BRIDGE_QT6_APP_CONTROLLER_H
