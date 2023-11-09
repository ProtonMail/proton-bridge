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


#ifndef BRIDGE_PP_EXCEPTION_H
#define BRIDGE_PP_EXCEPTION_H


#include <stdexcept>



namespace bridgepp {


//****************************************************************************************************************************************************
/// \brief Exception class.
//****************************************************************************************************************************************************
class Exception : public std::exception {
public: // member functions
    explicit Exception(QString qwhat = QString(), QString details = QString(), QString function = QString(),
        QByteArray attachment = QByteArray(), bool showSupportLink = false) noexcept; ///< Constructor
    Exception(Exception const &ref) noexcept; ///< copy constructor
    Exception(Exception &&ref) noexcept; ///< copy constructor
    Exception &operator=(Exception const &) = delete; ///< Disabled assignment operator
    Exception &operator=(Exception &&) = delete; ///< Disabled assignment operator
    ~Exception() noexcept override = default; ///< Destructor
    QString qwhat() const noexcept; ///< Return the description of the exception as a QString
    const char *what() const noexcept override; ///< Return the description of the exception as C style string
    QString details() const noexcept; ///< Return the details for the exception
    QString function() const noexcept; ///< Return the function that threw the exception.
    QByteArray attachment() const noexcept; ///< Return the attachment for the exception.
    QString detailedWhat() const; ///< Return the detailed description of the message (i.e. including the function name and the details).
    bool showSupportLink() const; ///< Return the value for the 'Show support link' option.

public: // static data members
    static qsizetype const attachmentMaxLength {25 * 1024}; ///< The maximum length text attachment sent in Sentry reports, in bytes.

private: // data members
    QString const qwhat_; ///< The description of the exception.
    QByteArray const what_; ///< The c-string version of the qwhat message. Stored as a QByteArray for automatic lifetime management.
    QString const details_; ///< The optional details for the exception.
    QString const function_; ///< The name of the function that created the exception.
    QByteArray const attachment_; ///< The attachment to add to the exception.
    bool const showSupportLink_; ///< Should the GUI feedback include a link to support.
};


} // namespace bridgepp


#endif //BRIDGE_PP_EXCEPTION_H
