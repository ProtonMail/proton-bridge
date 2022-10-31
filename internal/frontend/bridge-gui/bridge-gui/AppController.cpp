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


#include "AppController.h"
#include "QMLBackend.h"
#include <bridgepp/GRPC/GRPCClient.h>
#include <bridgepp/Exception/Exception.h>
#include <bridgepp/ProcessMonitor.h>
#include <bridgepp/Log/Log.h>


using namespace bridgepp;


//****************************************************************************************************************************************************
/// \return The AppController instance.
//****************************************************************************************************************************************************
AppController &app()
{
    static AppController app;
    return app;
}


//****************************************************************************************************************************************************
//
//****************************************************************************************************************************************************
AppController::AppController()
    : backend_(std::make_unique<QMLBackend>())
    , grpc_(std::make_unique<GRPCClient>())
    , log_(std::make_unique<Log>())
{
}


//****************************************************************************************************************************************************
/// \return The bridge worker, which can be null if the application was run in 'attach' mode (-a command-line switch).
//****************************************************************************************************************************************************
ProcessMonitor *AppController::bridgeMonitor() const
{
    if (!bridgeOverseer_)
        return nullptr;

    // null bridgeOverseer is OK, it means we run in 'attached' mode (app attached to an already runnning instance of Bridge).
    // but if bridgeOverseer is not null, its attached worker must be a valid ProcessMonitor instance.
    auto *monitor = dynamic_cast<ProcessMonitor*>(bridgeOverseer_->worker());
    if (!monitor)
        throw Exception("Could not retrieve bridge monitor");

    return monitor;
}


