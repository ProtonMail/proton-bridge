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


#ifndef BRIDGE_GUI_EVENT_STREAM_WORKER_H
#define BRIDGE_GUI_EVENT_STREAM_WORKER_H


#include <bridgepp/Worker/Worker.h>


//****************************************************************************************************************************************************
/// \brief Stream reader class.
//****************************************************************************************************************************************************
class EventStreamReader : public bridgepp::Worker {
Q_OBJECT
public: // member functions
    explicit EventStreamReader(QObject *parent); ///< Default constructor.
    EventStreamReader(EventStreamReader const &) = delete; ///< Disabled copy-constructor.
    EventStreamReader(EventStreamReader &&) = delete; ///< Disabled assignment copy-constructor.
    ~EventStreamReader() override = default; ///< Destructor.
    EventStreamReader &operator=(EventStreamReader const &) = delete; ///< Disabled assignment operator.
    EventStreamReader &operator=(EventStreamReader &&) = delete; ///< Disabled move assignment operator.

public slots:
    void run() override; ///< Run the reader.
    void onStarted() const; ///< Slot for the 'started' signal.
    void onFinished() const; ///< Slot for the 'finished' signal.

signals:
    void eventReceived(QString eventString); ///< signal for events.
};


#endif //BRIDGE_GUI_EVENT_STREAM_WORKER_H
