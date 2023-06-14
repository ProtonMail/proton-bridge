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


#ifndef BRIDGE_GUI_APP_CONTROLLER_H
#define BRIDGE_GUI_APP_CONTROLLER_H


//@formatter:off
class QMLBackend;
class Settings;
namespace bridgepp {
class Log;
class Overseer;
class GRPCClient;
class ProcessMonitor;
class Exception;
}
//@formatter:on


//****************************************************************************************************************************************************
/// \brief App controller class.
//****************************************************************************************************************************************************
class AppController : public QObject {
    Q_OBJECT
    friend AppController &app();

public: // member functions.
    AppController(AppController const &) = delete; ///< Disabled copy-constructor.
    AppController(AppController &&) = delete; ///< Disabled assignment copy-constructor.
    ~AppController() override; ///< Destructor.
    AppController &operator=(AppController const &) = delete; ///< Disabled assignment operator.
    AppController &operator=(AppController &&) = delete; ///< Disabled move assignment operator.
    QMLBackend &backend() { return *backend_; } ///< Return a reference to the backend.
    bridgepp::GRPCClient &grpc() { return *grpc_; } ///< Return a reference to the GRPC client.
    bridgepp::Log &log() { return *log_; } ///< Return a reference to the log.
    std::unique_ptr<bridgepp::Overseer> &bridgeOverseer() { return bridgeOverseer_; }; ///< Returns a reference the bridge overseer
    bridgepp::ProcessMonitor *bridgeMonitor() const; ///< Return the bridge worker.
    Settings &settings();; ///< Return the application settings.
    void setLauncherArgs(const QString &launcher, const QStringList &args); ///< Set the launcher arguments.
    void setSessionID(QString const &sessionID); ///< Set the sessionID.
    QString sessionID(); ///< Get the sessionID.

public slots:
    void onFatalError(bridgepp::Exception const &e); ///< Handle fatal errors.

private: // member functions
    AppController(); ///< Default constructor.
    void restart(bool isCrashing = false); ///< Restart the app.

private: // data members
    std::unique_ptr<QMLBackend> backend_; ///< The backend.
    std::unique_ptr<bridgepp::GRPCClient> grpc_; ///< The RPC client.
    std::unique_ptr<bridgepp::Log> log_; ///< The log.
    std::unique_ptr<bridgepp::Overseer> bridgeOverseer_; ///< The overseer for the bridge monitor worker.
    std::unique_ptr<Settings> settings_; ///< The application settings.
    QString launcher_; ///< The launcher.
    QStringList launcherArgs_; ///< The launcher arguments.
    QString sessionID_; ///<  The sessionID.
};


AppController &app(); ///< Return a reference to the app controller.


#endif // BRIDGE_GUI_APP_CONTROLLER_H
