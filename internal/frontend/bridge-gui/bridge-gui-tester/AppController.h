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


#ifndef BRIDGE_GUI_TESTER_APP_CONTROLLER_H
#define BRIDGE_GUI_TESTER_APP_CONTROLLER_H


class MainWindow;


class GRPCService;
namespace grpc { class StreamEvent; }
namespace bridgepp { class Log; }


//**********************************************************************************************************************
/// \brief Application controller class
//**********************************************************************************************************************
class AppController : public QObject {
Q_OBJECT
public: // member functions.
    friend AppController &app();

    AppController(AppController const &) = delete; ///< Disabled copy-constructor.
    AppController(AppController &&) = delete; ///< Disabled assignment copy-constructor.
    ~AppController() override; ///< Destructor.
    AppController &operator=(AppController const &) = delete; ///< Disabled assignment operator.
    AppController &operator=(AppController &&) = delete; ///< Disabled move assignment operator.
    void setMainWindow(MainWindow *mainWindow); ///< Set the main window.
    MainWindow &mainWindow() const; ///< Return the main window.
    bridgepp::Log &log() const; ///< Return a reference to the log.
    bridgepp::Log &bridgeGUILog() const; ///< Return a reference to the bridge-gui log.
    GRPCService &grpc() const; ///< Return a reference to the gRPC service.

private: // member functions.
    AppController(); ///< Default constructor.

private: // data members.
    MainWindow *mainWindow_ { nullptr }; ///< The main window.
    std::unique_ptr<bridgepp::Log> log_; ///< The log.
    std::unique_ptr<bridgepp::Log> bridgeGUILog_; ///< The bridge-gui log.
    std::unique_ptr<GRPCService> grpc_; ///< The gRPC service.
};


AppController &app(); ///< Return a reference to the app controller.


#endif // BRIDGE_GUI_TESTER_APP_CONTROLLER_H
