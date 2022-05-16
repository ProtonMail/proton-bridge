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
#include "User.h"
#include "GRPC/GRPCUtils.h"
#include "GRPC/GRPCClient.h"


//****************************************************************************************************************************************************
/// \param[in] parent The parent object.
//****************************************************************************************************************************************************
User::User(QObject *parent)
    : QObject(parent)
{

}


//****************************************************************************************************************************************************
/// \param[in] user The user to copy from
//****************************************************************************************************************************************************
void User::update(User const &user)
{
    this->setProperty("username", user.username_);
    this->setProperty("avatarText", user.avatarText_);
    this->setProperty("loggedIn", user.loggedIn_);
    this->setProperty("splitMode", user.splitMode_);
    this->setProperty("setupGuideSeen", user.setupGuideSeen_);
    this->setProperty("usedBytes", user.usedBytes_);
    this->setProperty("totalBytes", user.totalBytes_);
    this->setProperty("password", user.password_);
    this->setProperty("addresses", user.addresses_);
    this->setProperty("id", user.id_);
}

//****************************************************************************************************************************************************
/// \param[in] makeItActive Should split mode be made active.
//****************************************************************************************************************************************************
void User::toggleSplitMode(bool makeItActive)
{
    logGRPCCallStatus(app().grpc().setUserSplitMode(id_, makeItActive), "toggleSplitMode");
}

//****************************************************************************************************************************************************
//
//****************************************************************************************************************************************************
void User::logout()
{
    logGRPCCallStatus(app().grpc().logoutUser(id_), "logoutUser");
}


//****************************************************************************************************************************************************
//
//****************************************************************************************************************************************************
void User::remove()
{
    logGRPCCallStatus(app().grpc().removeUser(id_), "removeUser");
}


//****************************************************************************************************************************************************
/// \param[in] address The email address to configure Apple Mail for.
//****************************************************************************************************************************************************
void User::configureAppleMail(QString const &address)
{
    logGRPCCallStatus(app().grpc().configureAppleMail(id_, address), "configureAppleMail");
}


//****************************************************************************************************************************************************
// The only purpose of this call is to forward to the QML application the toggleSplitModeFinished(userID) event
// that was received by the UserList model.
//****************************************************************************************************************************************************
void User::emitToggleSplitModeFinished()
{
    this->setProperty("splitMode", QVariant::fromValue(!this->property("splitMode").toBool()));
    emit toggleSplitModeFinished();
}
