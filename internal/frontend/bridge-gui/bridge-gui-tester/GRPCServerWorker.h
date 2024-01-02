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


#ifndef BRIDGE_GUI_TESTER_SERVER_WORKER_H
#define BRIDGE_GUI_TESTER_SERVER_WORKER_H


#include <bridgepp/Worker/Worker.h>
#include "GRPCMetaDataProcessor.h"
#include "GRPCService.h"
#include <grpcpp/grpcpp.h>


//**********************************************************************************************************************
/// \brief gRPC server worker
//**********************************************************************************************************************
class GRPCServerWorker : public bridgepp::Worker {
Q_OBJECT
public: // member functions.
    explicit GRPCServerWorker(QObject *parent); ///< Default constructor.
    GRPCServerWorker(GRPCServerWorker const &) = delete; ///< Disabled copy-constructor.
    GRPCServerWorker(GRPCServerWorker &&) = delete; ///< Disabled assignment copy-constructor.
    ~GRPCServerWorker() override = default; ///< Destructor.
    GRPCServerWorker &operator=(GRPCServerWorker const &) = delete; ///< Disabled assignment operator.
    GRPCServerWorker &operator=(GRPCServerWorker &&) = delete; ///< Disabled move assignment operator.

    void run() override; ///< Run the worker.
    void stop() const;  ///< Stop the gRPC service.

private: // data members
    std::unique_ptr<grpc::Server> server_ { nullptr }; ///< The gRPC server.
    SPMetadataProcessor processor_;
};


#endif // BRIDGE_GUI_TESTER_SERVER_WORKER_H
