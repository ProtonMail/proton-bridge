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


#ifndef BRIDGE_PP_PROCESS_MONITOR_H
#define BRIDGE_PP_PROCESS_MONITOR_H


#include "Worker/Worker.h"


namespace bridgepp {


//**********************************************************************************************************************
/// \brief Process launcher and monitor class.
//**********************************************************************************************************************
class ProcessMonitor : public Worker {
Q_OBJECT
public: // static member functions
    struct MonitorStatus {
        bool ended = false;
        int returnCode = 0;
        qint64 pid = 0;
    };

public: // member functions.
    ProcessMonitor(QString const &exePath, QStringList const &args, QObject *parent); ///< Default constructor.
    ProcessMonitor(ProcessMonitor const &) = delete; ///< Disabled copy-constructor.
    ProcessMonitor(ProcessMonitor &&) = delete; ///< Disabled assignment copy-constructor.
    ~ProcessMonitor() override = default; ///< Destructor.
    ProcessMonitor &operator=(ProcessMonitor const &) = delete; ///< Disabled assignment operator.
    ProcessMonitor &operator=(ProcessMonitor &&) = delete; ///< Disabled move assignment operator.
    void run() override; ///< Run the worker.
    MonitorStatus const getStatus(); ///< Retrieve the current status of the process.

signals:
    void processExited(int code); ///< Slot for the exiting of the process.

private: // member functions
    void forwardProcessOutput(QProcess &p); ///< Forward the standard output and error from the process to this application standard output and error.

private: // data members
    QMutex statusMutex_; ///< The status mutex.
    QString const exePath_; ///< The path to the executable.
    QStringList args_; ///< arguments to be passed to Bridge.
    MonitorStatus status_; ///< Status of the monitoring.
    QTextStream out_; ///< The standard output stream.
    QTextStream err_; ///< The standard error stream.
};


}


#endif //BRIDGE_PP_PROCESS_MONITOR_H
