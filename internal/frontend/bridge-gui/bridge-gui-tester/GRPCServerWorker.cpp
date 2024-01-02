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


#include "GRPCServerWorker.h"
#include "Cert.h"
#include "GRPCService.h"
#include <bridgepp/Exception/Exception.h>
#include <bridgepp/BridgeUtils.h>
#include <bridgepp/GRPC/GRPCUtils.h>
#include <bridgepp/GRPC/GRPCConfig.h>


using namespace bridgepp;
using namespace grpc;


//****************************************************************************************************************************************************
//
//****************************************************************************************************************************************************
GRPCServerWorker::GRPCServerWorker(QObject *parent)
    : Worker(parent) {
}


//****************************************************************************************************************************************************
//
//****************************************************************************************************************************************************
void GRPCServerWorker::run() {
    try {
        emit started();

        SslServerCredentialsOptions::PemKeyCertPair pair;
        pair.private_key = testTLSKey.toStdString();
        pair.cert_chain = testTLSCert.toStdString();
        SslServerCredentialsOptions ssl_opts;
        ssl_opts.pem_root_certs = "";
        ssl_opts.pem_key_cert_pairs.push_back(pair);
        std::shared_ptr<ServerCredentials> const credentials = SslServerCredentials(ssl_opts);

        GRPCConfig config;
        config.cert = testTLSCert;
        config.token = QUuid::createUuid().toString();
        processor_ = std::make_shared<GRPCMetadataProcessor>(config.token);
        credentials->SetAuthMetadataProcessor(processor_); // gRPC interceptors are still experimental in C++, so we use AuthMetadataProcessor
        ServerBuilder builder;
        int port = 0; // Port will not be known until ServerBuilder::BuildAndStart() is called
        if (useFileSocketForGRPC()) {
            QString const fileSocketPath = getAvailableFileSocketPath();
            if (fileSocketPath.isEmpty()) {
                throw Exception("Could not get an available file socket.");
            }
            builder.AddListeningPort(QString("unix://%1").arg(fileSocketPath).toStdString(), credentials);
            config.fileSocketPath = fileSocketPath;
        } else {
            builder.AddListeningPort("127.0.0.1:0", credentials, &port);
        }

        builder.RegisterService(&app().grpc());
        server_ = builder.BuildAndStart();

        if (!server_) {
            throw Exception("Could not create gRPC server.");
        }
        app().log().debug("gRPC Server started.");

        config.port = port;
        QString err;
        if (!config.save(grpcServerConfigPath(bridgepp::userConfigDir()), &err)) {
            throw Exception(QString("Could not save gRPC server config. %1").arg(err));
        }

        server_->Wait();
        emit finished();
        app().log().debug("gRPC Server exited.");
    }
    catch (Exception const &e) {
        if (server_) {
            server_->Shutdown();
        }

        emit error(e.qwhat());
    }
}


//****************************************************************************************************************************************************
//
//****************************************************************************************************************************************************
void GRPCServerWorker::stop() const {
    if (server_) {
        server_->Shutdown();
    }
}


