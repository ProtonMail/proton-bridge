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


#include "EventStreamWorker.h"
#include "SentryUtils.h"
#include <bridgepp/GRPC/GRPCClient.h>
#include <bridgepp/Exception/Exception.h>
#include <bridgepp/Log/Log.h>


using namespace bridgepp;


//****************************************************************************************************************************************************
/// \param[in] parent The parent object.
//****************************************************************************************************************************************************
EventStreamReader::EventStreamReader(QObject *parent)
    : Worker(parent) {
    connect(this, &EventStreamReader::started, this, &EventStreamReader::onStarted);
    connect(this, &EventStreamReader::finished, this, &EventStreamReader::onFinished);
    connect(this, &EventStreamReader::error, &app().log(), &Log::error);
}


//****************************************************************************************************************************************************
//
//****************************************************************************************************************************************************
void EventStreamReader::run() {
    try {
        emit started();
        // Status code for the call below is ignored. The event stream may have interrupted by system shutdown or OS user sign-out, and we do not
        // want this to generate a sentry report.
        app().grpc().runEventStreamReader();
        emit finished();
    }
    catch (Exception const &e) {
        reportSentryException("Error during event stream read", e);
        emit error(e.qwhat());
    }
}


//****************************************************************************************************************************************************
//
//****************************************************************************************************************************************************
void EventStreamReader::onStarted() const {
    app().log().debug("EventStreamReader started");
}


//****************************************************************************************************************************************************
//
//****************************************************************************************************************************************************
void EventStreamReader::onFinished() const {
    app().log().debug("EventStreamReader finished");
    if (!app().bridgeMonitor()) {
        // no bridge monitor means we are in a debug environment, running in attached mode. Event stream has terminated, so bridge is shutting
        // down. Because we're in attached mode, bridge-gui will not get notified that bridge is going down, so we shutdown manually here.
        qApp->exit(EXIT_SUCCESS);
    }
}
