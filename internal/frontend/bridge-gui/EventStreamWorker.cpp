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


#include "EventStreamWorker.h"
#include <bridgepp/GRPC/GRPCClient.h>
#include <bridgepp/Exception/Exception.h>
#include <bridgepp/Log/Log.h>


using namespace bridgepp;


//****************************************************************************************************************************************************
/// \param[in] parent The parent object.
//****************************************************************************************************************************************************
EventStreamReader::EventStreamReader(QObject *parent)
    : Worker(parent)
{
    connect(this, &EventStreamReader::started, [&]() { app().log().debug("EventStreamReader started");});
    connect(this, &EventStreamReader::finished, [&]() { app().log().debug("EventStreamReader finished");});
    connect(this, &EventStreamReader::error, &app().log(), &Log::error);
}


//****************************************************************************************************************************************************
//
//****************************************************************************************************************************************************
void EventStreamReader::run()
{
    try
    {
        emit started();

        grpc::Status const status = app().grpc().startEventStream();
        if (!status.ok())
            throw Exception(QString::fromStdString(status.error_message()));

        emit finished();
    }
    catch (Exception const &e)
    {
        emit error(e.qwhat());
    }
}
