// Copyright (c) 2023 Proton AG
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


#include "Exception.h"


namespace bridgepp {


//****************************************************************************************************************************************************
/// \param[in] what A description of the exception.
/// \param[in] details The optional details for the exception.
//****************************************************************************************************************************************************
Exception::Exception(QString what, QString details) noexcept
    : std::exception()
    , what_(std::move(what))
    , details_(std::move(details)) {
}


//****************************************************************************************************************************************************
/// \param[in] ref The Exception to copy from
//****************************************************************************************************************************************************
Exception::Exception(Exception const &ref) noexcept
    : std::exception(ref)
    , what_(ref.what_)
    , details_(ref.details_) {
}


//****************************************************************************************************************************************************
/// \param[in] ref The Exception to copy from
//****************************************************************************************************************************************************
Exception::Exception(Exception &&ref) noexcept
    : std::exception(ref)
    , what_(ref.what_)
    , details_(ref.details_) {
}


//****************************************************************************************************************************************************
/// \return a string describing the exception
//****************************************************************************************************************************************************
QString Exception::qwhat() const noexcept {
    return what_;
}


//****************************************************************************************************************************************************
/// \return A pointer to the description string of the exception.
//****************************************************************************************************************************************************
const char *Exception::what() const noexcept {
    return what_.toLocal8Bit().constData();
}


//****************************************************************************************************************************************************
/// \return The details for the exception.
//****************************************************************************************************************************************************
QString Exception::details() const noexcept {
    return details_;
}


} // namespace bridgepp
