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


#include "GRPCErrors.h"


using namespace grpc;


namespace bridgepp {


namespace {


ErrorInfo const unknownError { UNKNOWN_ERROR, QObject::tr("Unknown error"), QObject::tr("An unknown error occurred.") }; // Unknown error


QList<ErrorInfo> const errorList {
    unknownError,
    { TLS_CERT_EXPORT_ERROR, QObject::tr("Export error"), QObject::tr("The TLS certificate could not be exported.") },
    { TLS_KEY_EXPORT_ERROR, QObject::tr("Export error"), QObject::tr("The TLS private key could not be exported.") },
};


} //< anonymous namespace


//****************************************************************************************************************************************************
/// \param[in] error
//****************************************************************************************************************************************************
ErrorInfo errorInfo(grpc::ErrorCode code) {
    QList<ErrorInfo>::const_iterator it = std::find_if(errorList.begin(), errorList.end(), [code](ErrorInfo info) -> bool { return code == info.code; });
    return errorList.end() == it ? unknownError : *it;
}


} // namespace bridgepp