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


#include "GRPCServerWorker.h"
#include "GRPCService.h"
#include <bridgepp/Exception/Exception.h>
#include <bridgepp/GRPC/GRPCUtils.h>


using namespace bridgepp;
using namespace grpc;


namespace
{


//****************************************************************************************************************************************************
/// \return The content of the file.
//****************************************************************************************************************************************************
QString loadAsciiTextFile(QString const &path) {
    QFile file(path);
    return file.open(QIODevice::ReadOnly | QIODevice::Text) ? QString::fromLocal8Bit(file.readAll()) : QString();
}


}


//****************************************************************************************************************************************************
//
//****************************************************************************************************************************************************
GRPCServerWorker::GRPCServerWorker(QObject *parent)
    : Worker(parent)
{

}


//****************************************************************************************************************************************************
//
//****************************************************************************************************************************************************
void GRPCServerWorker::run()
{
    try
    {
        emit started();

        QString cert = loadAsciiTextFile(serverCertificatePath());
        if (cert.isEmpty())
            throw Exception("Could not locate server certificate. Make sure to launch bridge once before starting this application");

        QString key = loadAsciiTextFile(serverKeyPath());
        if (key.isEmpty())
            throw Exception("Could not locate server key. Make sure to launch bridge once before starting this application");

        SslServerCredentialsOptions::PemKeyCertPair pair;
        pair.private_key = key.toStdString();
        pair.cert_chain = cert.toStdString();
        SslServerCredentialsOptions ssl_opts;
        ssl_opts.pem_root_certs="";
        ssl_opts.pem_key_cert_pairs.push_back(pair);
        std::shared_ptr<ServerCredentials> credentials  = grpc::SslServerCredentials(ssl_opts);

        ServerBuilder builder;
        builder.AddListeningPort("127.0.0.1:9292", credentials);
        builder.RegisterService(&app().grpc());
        server_ = builder.BuildAndStart();
        if (!server_)
            throw Exception("Could not create gRPC server.");
        app().log().debug("gRPC Server started.");

        server_->Wait();
        emit finished();
        app().log().debug("gRPC Server exited.");
    }
    catch (Exception const &e)
    {
        if (server_)
            server_->Shutdown();

        emit error(e.qwhat());
    }
}


//****************************************************************************************************************************************************
//
//****************************************************************************************************************************************************
void GRPCServerWorker::stop()
{
    if (server_)
        server_->Shutdown();
}
