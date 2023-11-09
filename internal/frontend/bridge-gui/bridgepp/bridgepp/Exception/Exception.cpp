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
/// \param[in] function The name of the calling function.
/// \param[in] showSupportLink Should a link to the support web form be included in GUI message.
//****************************************************************************************************************************************************
Exception::Exception(QString qwhat, QString details, QString function, QByteArray attachment, bool showSupportLink) noexcept
    : std::exception()
    , qwhat_(std::move(qwhat))
    , what_(qwhat_.toLocal8Bit())
    , details_(std::move(details))
    , function_(std::move(function))
    , attachment_(std::move(attachment))
    , showSupportLink_(showSupportLink) {
}


//****************************************************************************************************************************************************
/// \param[in] ref The Exception to copy from
//****************************************************************************************************************************************************
Exception::Exception(Exception const &ref) noexcept
    : std::exception(ref)
    , qwhat_(ref.qwhat_)
    , what_(ref.what_)
    , details_(ref.details_)
    , function_(ref.function_)
    , attachment_(ref.attachment_)
    , showSupportLink_(ref.showSupportLink_) {
}


//****************************************************************************************************************************************************
/// \param[in] ref The Exception to copy from
//****************************************************************************************************************************************************
Exception::Exception(Exception &&ref) noexcept
    : std::exception(ref)
    , qwhat_(ref.qwhat_)
    , what_(ref.what_)
    , details_(ref.details_)
    , function_(ref.function_)
    , attachment_(ref.attachment_)
    , showSupportLink_(ref.showSupportLink_) {
}


//****************************************************************************************************************************************************
/// \return a string describing the exception
//****************************************************************************************************************************************************
QString Exception::qwhat() const noexcept {
    return qwhat_;
}


//****************************************************************************************************************************************************
/// \return A pointer to the description string of the exception.
//****************************************************************************************************************************************************
const char *Exception::what() const noexcept {
    return what_.constData();
}


//****************************************************************************************************************************************************
/// \return The details for the exception.
//****************************************************************************************************************************************************
QString Exception::details() const noexcept {
    return details_;
}


//****************************************************************************************************************************************************
/// \return The function that threw the exception.
//****************************************************************************************************************************************************
QString Exception::function() const noexcept {
    return function_;
}


//****************************************************************************************************************************************************
/// \return The attachment for the exception.
//****************************************************************************************************************************************************
QByteArray Exception::attachment() const noexcept {
    return attachment_;
}


//****************************************************************************************************************************************************
/// \return The details exception.
//****************************************************************************************************************************************************
QString Exception::detailedWhat() const {
    QString result = qwhat_;
    if (!function_.isEmpty()) {
        result = QString("%1(): %2").arg(function_, result);
    }
    if (!details_.isEmpty()) {
        result += "\n\nDetails:\n" + details_;
    }
    return result;
}


//****************************************************************************************************************************************************
/// \return true iff A link to the support page should shown in the GUI message box.
//****************************************************************************************************************************************************
bool Exception::showSupportLink() const {
    return showSupportLink_;
}


} // namespace bridgepp
