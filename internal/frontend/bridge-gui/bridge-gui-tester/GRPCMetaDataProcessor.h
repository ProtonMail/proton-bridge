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


#ifndef BRIDGE_GUI_TESTER_GRPC_METADATA_PROCESSOR_H
#define BRIDGE_GUI_TESTER_GRPC_METADATA_PROCESSOR_H


#include <grpc++/grpc++.h>


//****************************************************************************************************************************************************
/// \brief Metadata processor class in charge of checking the server token provided by client calls and stream requests.
//****************************************************************************************************************************************************
class GRPCMetadataProcessor : public grpc::AuthMetadataProcessor {
public: // member functions.
    GRPCMetadataProcessor(QString const &serverToken); ///< Default constructor.
    GRPCMetadataProcessor(GRPCMetadataProcessor const &) = delete; ///< Disabled copy-constructor.
    GRPCMetadataProcessor(GRPCMetadataProcessor &&) = delete; ///< Disabled assignment copy-constructor.
    ~GRPCMetadataProcessor() = default; ///< Destructor.
    GRPCMetadataProcessor &operator=(GRPCMetadataProcessor const &) = delete; ///< Disabled assignment operator.
    GRPCMetadataProcessor &operator=(GRPCMetadataProcessor &&) = delete; ///< Disabled move assignment operator.
    bool IsBlocking() const override; ///< Is the processor blocking?
    grpc::Status Process(InputMetadata const &authMetadata, grpc::AuthContext *context, OutputMetadata *consumed_auth_metadata,
        OutputMetadata *response_metadata) override; ///< Process the metadata

private:
    std::string serverToken_; ///< The server token, as a std::string.
};


typedef std::shared_ptr<GRPCMetadataProcessor> SPMetadataProcessor; ///< Type definition for shared pointer to MetadataProcessor.


#endif //BRIDGE_GUI_TESTER_GRPC_METADATA_PROCESSOR_H
