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


#ifndef BRIDGE_PP_WORKER_H
#define BRIDGE_PP_WORKER_H


namespace bridgepp {


//****************************************************************************************************************************************************
/// \brief Pure virtual class for worker intended to perform a threaded operation.
//****************************************************************************************************************************************************
class Worker : public QObject {
Q_OBJECT
public: // member functions
    explicit Worker(QObject *parent)
        : QObject(parent) {} ///< Default constructor.
    Worker(Worker const &) = delete; ///< Disabled copy-constructor.
    Worker(Worker &&) = delete; ///< Disabled assignment copy-constructor.
    ~Worker() override = default; ///< Destructor.
    Worker &operator=(Worker const &) = delete; ///< Disabled assignment operator.
    Worker &operator=(Worker &&) = delete; ///< Disabled move assignment operator.

public slots:
    virtual void run() = 0; ///< run the worker.

signals:
    void started(); ///< Signal for the start of the worker
    void finished(); ///< Signal for the end of the worker
    void error(QString const &message); ///< Signal for errors. After an error, worker ends and finished is NOT emitted.
    void cancelled(); ///< Signal for the cancellation of the worker.
};


} // namespace bridgepp


#endif //BRIDGE_PP_WORKER_H
