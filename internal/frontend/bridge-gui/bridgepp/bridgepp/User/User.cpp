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


#include "User.h"


namespace bridgepp {


//****************************************************************************************************************************************************
/// \param[in] parent The parent object of the user.
//****************************************************************************************************************************************************
SPUser User::newUser(QObject *parent) {
    return SPUser(new User(parent));
}


//****************************************************************************************************************************************************
/// \param[in] parent The parent object.
//****************************************************************************************************************************************************
User::User(QObject *parent)
    : QObject(parent)
    , imapFailureCooldownEndTime_(QDateTime::currentDateTime()) {

}


//****************************************************************************************************************************************************
/// \param[in] user The user to copy from
//****************************************************************************************************************************************************
void User::update(User const &user) {
    this->setID(user.id());
    this->setUsername(user.username());
    this->setPassword(user.password());
    this->setAddresses(user.addresses());
    this->setAvatarText(user.avatarText());
    this->setState(user.state());
    this->setSplitMode(user.splitMode());
    this->setUsedBytes(user.usedBytes());
    this->setTotalBytes(user.totalBytes());
}


//****************************************************************************************************************************************************
/// \return The user's primary email. If not known, return turn username
//****************************************************************************************************************************************************
QString User::primaryEmailOrUsername() const {
    return addresses_.isEmpty() ? username_ : addresses_.front();
}


//****************************************************************************************************************************************************
/// \param[in] makeItActive Should split mode be made active.
//****************************************************************************************************************************************************
void User::toggleSplitMode(bool makeItActive) {
    emit toggleSplitModeForUser(id_, makeItActive);
}


//****************************************************************************************************************************************************
//
//****************************************************************************************************************************************************
void User::logout() {
    emit logoutUser(id_);
}


//****************************************************************************************************************************************************
//
//****************************************************************************************************************************************************
void User::remove() {
    emit removeUser(id_);
}


//****************************************************************************************************************************************************
/// \param[in] address The email address to configure Apple Mail for.
//****************************************************************************************************************************************************
void User::configureAppleMail(QString const &address) {
    emit configureAppleMailForUser(id_, address);
}


//****************************************************************************************************************************************************
// The only purpose of this call is to forward to the QML application the toggleSplitModeFinished(userID) event
// that was received by the UserList model.
//****************************************************************************************************************************************************
void User::emitToggleSplitModeFinished() {
    emit toggleSplitModeFinished();
}


//****************************************************************************************************************************************************
/// \return The userID.
//****************************************************************************************************************************************************
QString User::id() const {
    return id_;
}


//****************************************************************************************************************************************************
/// \param[in] id The userID.
//****************************************************************************************************************************************************
void User::setID(QString const &id) {
    if (id == id_) {
        return;
    }

    id_ = id;
    emit idChanged(id_);
}


//****************************************************************************************************************************************************
/// \return The username.
//****************************************************************************************************************************************************
QString User::username() const {
    return username_;
}


//****************************************************************************************************************************************************
/// \param[in] username The username.
//****************************************************************************************************************************************************
void User::setUsername(QString const &username) {
    if (username == username_) {
        return;
    }

    username_ = username;
    emit usernameChanged(username_);
}


//****************************************************************************************************************************************************
/// \return The password.
//****************************************************************************************************************************************************
QString User::password() const {
    return password_;
}


//****************************************************************************************************************************************************
/// \param[in] password The password.
//****************************************************************************************************************************************************
void User::setPassword(QString const &password) {
    if (password == password_) {
        return;
    }

    password_ = password;
    emit passwordChanged(password_);
}


//****************************************************************************************************************************************************
/// \return The addresses.
//****************************************************************************************************************************************************
QStringList User::addresses() const {
    return addresses_;
}


//****************************************************************************************************************************************************
/// \param[in] addresses The addresses.
//****************************************************************************************************************************************************
void User::setAddresses(QStringList const &addresses) {
    if (addresses == addresses_) {
        return;
    }

    addresses_ = addresses;
    emit addressesChanged(addresses_);
}


//****************************************************************************************************************************************************
/// \return The avatar text.
//****************************************************************************************************************************************************
QString User::avatarText() const {
    return avatarText_;
}


//****************************************************************************************************************************************************
/// \param[in] avatarText The avatar text.
//****************************************************************************************************************************************************
void User::setAvatarText(QString const &avatarText) {
    if (avatarText == avatarText_) {
        return;
    }

    avatarText_ = avatarText;
    emit usernameChanged(avatarText_);
}


//****************************************************************************************************************************************************
/// \return The user state.
//****************************************************************************************************************************************************
UserState User::state() const {
    return state_;
}


//****************************************************************************************************************************************************
/// \param[in] state The user state.
//****************************************************************************************************************************************************
void User::setState(UserState state) {
    if (state_ == state) {
        return;
    }

    state_ = state;
    emit stateChanged(state);
}


//****************************************************************************************************************************************************
/// \return The split mode status.
//****************************************************************************************************************************************************
bool User::splitMode() const {
    return splitMode_;
}


//****************************************************************************************************************************************************
/// \param[in] splitMode The split mode status.
//****************************************************************************************************************************************************
void User::setSplitMode(bool splitMode) {
    if (splitMode == splitMode_) {
        return;
    }

    splitMode_ = splitMode;
    emit splitModeChanged(splitMode_);
}


//****************************************************************************************************************************************************
/// \return The used bytes.
//****************************************************************************************************************************************************
float User::usedBytes() const {
    return usedBytes_;
}


//****************************************************************************************************************************************************
/// \param[in] usedBytes The used bytes.
//****************************************************************************************************************************************************
void User::setUsedBytes(float usedBytes) {
    if (usedBytes == usedBytes_) {
        return;
    }

    usedBytes_ = usedBytes;
    emit usedBytesChanged(usedBytes_);
}


//****************************************************************************************************************************************************
/// \return The total bytes.
//****************************************************************************************************************************************************
float User::totalBytes() const {
    return totalBytes_;
}


//****************************************************************************************************************************************************
/// \param[in] totalBytes The total bytes.
//****************************************************************************************************************************************************
void User::setTotalBytes(float totalBytes) {
    if (totalBytes == totalBytes_) {
        return;
    }

    totalBytes_ = totalBytes;
    emit totalBytesChanged(totalBytes_);
}


//****************************************************************************************************************************************************
/// \param[in] state The user state.
/// \return A string describing the state.
//****************************************************************************************************************************************************
QString User::stateToString(UserState state) {
    switch (state) {
    case UserState::SignedOut:
        return "Signed out";
    case UserState::Locked:
        return "Locked";
    case UserState::Connected:
        return "Connected";
    default:
        return "Unknown";
    }
}


//****************************************************************************************************************************************************
/// We display a notification and pop the application window if an IMAP client tries to connect to a signed out account, but we do not want to
/// do it repeatedly, as it's an intrusive action. This function let's you define a period of time during which the notification should not be
/// displayed.
///
/// \param durationMSecs The duration of the period in milliseconds.
//****************************************************************************************************************************************************
void User::startImapLoginFailureCooldown(qint64 durationMSecs) {
    imapFailureCooldownEndTime_ = QDateTime::currentDateTime().addMSecs(durationMSecs);
}


//****************************************************************************************************************************************************
/// \return true if we currently are in a cooldown period for the notification
//****************************************************************************************************************************************************
bool User::isInIMAPLoginFailureCooldown() const {
    return QDateTime::currentDateTime() < imapFailureCooldownEndTime_;
}


} // namespace bridgepp
