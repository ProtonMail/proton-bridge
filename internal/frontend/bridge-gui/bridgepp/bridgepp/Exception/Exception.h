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


#ifndef BRIDGE_PP_EXCEPTION_H
#define BRIDGE_PP_EXCEPTION_H


#include <stdexcept>


namespace bridgepp
{


//****************************************************************************************************************************************************
/// \brief Exception class.
//****************************************************************************************************************************************************
class Exception : public std::exception
{
public: // member functions
    explicit Exception(QString what = QString()) noexcept; ///< Constructor
    Exception(Exception const &ref) noexcept; ///< copy constructor
    Exception(Exception &&ref) noexcept; ///< copy constructor
    Exception &operator=(Exception const &) = delete; ///< Disabled assignment operator
    Exception &operator=(Exception &&) = delete; ///< Disabled assignment operator
    ~Exception() noexcept override = default; ///< Destructor
    QString const &qwhat() const noexcept; ///< Return the description of the exception as a QString
    const char *what() const noexcept override; ///< Return the description of the exception as C style string

private: // data members
    QString const what_; ///< The description of the exception
};


} // namespace bridgepp


#endif //BRIDGE_PP_EXCEPTION_H
