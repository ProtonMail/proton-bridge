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



#ifndef BRIDGE_GUI_BRIDGE_MONITOR_H
#define BRIDGE_GUI_BRIDGE_MONITOR_H


#include <bridgepp/Worker/Worker.h>


//**********************************************************************************************************************
/// \brief Bridge process launcher and monitor class.
//**********************************************************************************************************************
class BridgeMonitor: public bridgepp::Worker
{
    Q_OBJECT
public: // static member functions
    static QString locateBridgeExe(); ///< Try to find the bridge executable path.

    struct MonitorStatus {
        bool running = false;
        int returnCode = 0;
        qint64 pid = 0;
    };

public: // member functions.
    BridgeMonitor(QString const& exePath, QStringList const &args, QObject *parent); ///< Default constructor.
    BridgeMonitor(BridgeMonitor const&) = delete; ///< Disabled copy-constructor.
    BridgeMonitor(BridgeMonitor&&) = delete; ///< Disabled assignment copy-constructor.
    ~BridgeMonitor() override = default; ///< Destructor.
    BridgeMonitor& operator=(BridgeMonitor const&) = delete; ///< Disabled assignment operator.
    BridgeMonitor& operator=(BridgeMonitor&&) = delete; ///< Disabled move assignment operator.
    void run() override; ///< Run the worker.

    const MonitorStatus& getStatus();
signals:
    void processExited(int code); ///< Slot for the exiting of the process.

private: // data members
    QString const exePath_; ///< The path to the bridge executable.
    QStringList args_; ///< arguments to be passed to the brigde.
    MonitorStatus status_; ///< Status of the monitoring.
};



#endif // BRIDGE_GUI_BRIDGE_MONITOR_H
