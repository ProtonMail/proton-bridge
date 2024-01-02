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


#include "AppController.h"
#include "GRPCService.h"
#include <bridgepp/Exception/Exception.h>
#include "MainWindow.h"
#include <bridgepp/Log/Log.h>


using namespace bridgepp;


//****************************************************************************************************************************************************
/// \return A reference to the application controller.
//****************************************************************************************************************************************************
AppController &app() {
    static AppController app;
    return app;
}


//****************************************************************************************************************************************************
//
//****************************************************************************************************************************************************
AppController::AppController()
    : log_(std::make_unique<Log>())
    , bridgeGUILog_(std::make_unique<Log>())
    , grpc_(std::make_unique<GRPCService>()) {

}


//****************************************************************************************************************************************************
//
//****************************************************************************************************************************************************
AppController::~AppController() // NOLINT(modernize-use-equals-default): implementation in cpp file is required because of forward declaration of Log in header
{

}


//****************************************************************************************************************************************************
/// \param[in] mainWindow The main window.
//****************************************************************************************************************************************************
void AppController::setMainWindow(MainWindow *mainWindow) {
    mainWindow_ = mainWindow;
    grpc_->connectProxySignals();
}


//****************************************************************************************************************************************************
/// \return The main window.
//****************************************************************************************************************************************************
MainWindow &AppController::mainWindow() const {
    if (!mainWindow_) {
        throw Exception("mainWindow has not yet been registered.");
    }
    return *mainWindow_;
}


//****************************************************************************************************************************************************
/// \return A reference to the log.
//****************************************************************************************************************************************************
bridgepp::Log &AppController::log() const {
    return *log_;
}


//****************************************************************************************************************************************************
/// \return A reference to the bridge-gui log.
//****************************************************************************************************************************************************
bridgepp::Log &AppController::bridgeGUILog() const {
    return *bridgeGUILog_;
}


//****************************************************************************************************************************************************
/// \return A reference to the gRPC service.
//****************************************************************************************************************************************************
GRPCService &AppController::grpc() const {
    return *grpc_;
}
