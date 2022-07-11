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


#include "Pch.h"
#include "Exception.h"


//****************************************************************************************************************************************************
/// \param[in] what A description of the exception
//****************************************************************************************************************************************************
Exception::Exception(QString what) noexcept
    : std::exception()
    , what_(std::move(what))
{
}


//****************************************************************************************************************************************************
/// \param[in] ref The Exception to copy from
//****************************************************************************************************************************************************
Exception::Exception(Exception const& ref) noexcept
    : std::exception(ref)
    , what_(ref.what_)
{
}


//****************************************************************************************************************************************************
/// \param[in] ref The Exception to copy from
//****************************************************************************************************************************************************
Exception::Exception(Exception&& ref) noexcept
    : std::exception(ref)
    , what_(ref.what_)
{
}


//****************************************************************************************************************************************************
/// \return a string describing the exception
//****************************************************************************************************************************************************
QString const& Exception::qwhat() const noexcept
{
    return what_;
}


//****************************************************************************************************************************************************
/// \return A pointer to the description string of the exception.
//****************************************************************************************************************************************************
const char* Exception::what() const noexcept
{
    return what_.toLocal8Bit().constData();
}
