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


#include "UserList.h"


using namespace bridgepp;


//****************************************************************************************************************************************************
/// \param[in] parent The parent object of the user list.
//****************************************************************************************************************************************************
UserList::UserList(QObject *parent)
    : QAbstractListModel(parent) {
    /// \todo use mutex to prevent concurrent access
}


//****************************************************************************************************************************************************
//
//****************************************************************************************************************************************************
void UserList::connectGRPCEvents() const {
    GRPCClient &client = app().grpc();
    connect(&client, &GRPCClient::userChanged, this, &UserList::onUserChanged);
    connect(&client, &GRPCClient::toggleSplitModeFinished, this, &UserList::onToggleSplitModeFinished);
    connect(&client, &GRPCClient::usedBytesChanged, this, &UserList::onUsedBytesChanged);
    connect(&client, &GRPCClient::syncStarted, this, &UserList::onSyncStarted);
    connect(&client, &GRPCClient::syncFinished, this, &UserList::onSyncFinished);
    connect(&client, &GRPCClient::syncProgress, this, &UserList::onSyncProgress);
}


//****************************************************************************************************************************************************
//
//****************************************************************************************************************************************************
int UserList::rowCount(QModelIndex const &) const {
    return users_.size();
}


//****************************************************************************************************************************************************
/// \param[in] index The index to retrieve data for.
/// \param[in] role The role to retrieve data for.
/// \return The data at the index for the given role.
//****************************************************************************************************************************************************
QVariant UserList::data(QModelIndex const &index, int role) const {
    /// This It does not seem to be used, but the method is required by the base class.
    /// From the original QtThe recipe QML backend User model, the User is always returned, regardless of the role.
    Q_UNUSED(role)
    int const row = index.row();
    if ((row < 0) || (row >= users_.size())) {
        return QVariant();
    }
    return QVariant::fromValue(users_[row].get());
}


//****************************************************************************************************************************************************
/// \param[in] userID The userID.
/// \return the row of the user.
/// \return -1 if the userID is not in the list
//****************************************************************************************************************************************************
int UserList::rowOfUserID(QString const &userID) const {
    for (qint32 row = 0; row < users_.count(); ++row) {
        if (userID == users_[row]->property("id")) {
            return row;
        }
    }
    return -1;
}


//****************************************************************************************************************************************************
/// \param[in] users The new user list.
//****************************************************************************************************************************************************
void UserList::reset(QList<SPUser> const &users) {
    this->beginResetModel();
    users_ = users;
    this->endResetModel();
    emit countChanged(users_.size());
}


//****************************************************************************************************************************************************
/// \param[in] user The user.
//****************************************************************************************************************************************************
void UserList::appendUser(SPUser const &user) {
    int const size = users_.size();
    this->beginInsertRows(QModelIndex(), size, size);
    users_.append(user);
    this->endInsertRows();
    emit countChanged(users_.size());
}


//****************************************************************************************************************************************************
/// \param[in] row The row.
//****************************************************************************************************************************************************
void UserList::removeUserAt(int row) {
    if ((row < 0) && (row >= users_.size())) {
        return;
    }
    this->beginRemoveRows(QModelIndex(), row, row);
    users_.removeAt(row);
    this->endRemoveRows();
    emit countChanged(users_.size());
}


//****************************************************************************************************************************************************
/// \param[in] row The row.
/// \param[in] user The user.
//****************************************************************************************************************************************************
void UserList::updateUserAtRow(int row, User const &user) {
    if ((row < 0) || (row >= users_.count())) {
        app().log().error(QString("invalid user at row %2 (user userCount = %2)").arg(row).arg(users_.count()));
        return;
    }

    users_[row]->update(user);

    QModelIndex modelIndex = this->index(row);
    emit dataChanged(modelIndex, modelIndex);
}


//****************************************************************************************************************************************************
/// \param[in] userID The userID.
/// \return The user with the given ID.
/// \return A null pointer if the user could not be found.
//****************************************************************************************************************************************************
bridgepp::SPUser UserList::getUserWithID(QString const &userID) const {
    QList<SPUser>::const_iterator it = std::find_if(users_.begin(), users_.end(), [userID](SPUser const &user) -> bool {
        return user && user->id() == userID;
    });
    return (it == users_.end()) ? nullptr : *it;
}


//****************************************************************************************************************************************************
/// \param[in] username The username or email.
/// \return The user with the given ID.
/// \return A null pointer if the user could not be found.
//****************************************************************************************************************************************************
bridgepp::SPUser UserList::getUserWithUsernameOrEmail(QString const &username) const {
    QList<SPUser>::const_iterator it = std::find_if(users_.begin(), users_.end(), [username](SPUser const &user) -> bool {
        return user && ((username.compare(user->username(), Qt::CaseInsensitive) == 0) || user->addresses().contains(username, Qt::CaseInsensitive));
    });
    return (it == users_.end()) ? nullptr : *it;
}


//****************************************************************************************************************************************************
/// \param[in] row The row.
//****************************************************************************************************************************************************
User *UserList::get(int row) const {
    if ((row < 0) || (row >= users_.count())) {
        return nullptr;
    }

    app().log().trace(QString("Retrieving user at row %1 (user userCount = %2)").arg(row).arg(users_.count()));
    return users_[row].get();
}


//****************************************************************************************************************************************************
/// \param[in] userID The userID.
/// \return The primary email address (or if unknown the username) of the user.
/// \return An empty string if the user cannot be found.
//****************************************************************************************************************************************************
QString UserList::primaryEmailOrUsername(QString const &userID) const {
    SPUser const user = this->getUserWithID(userID);
    return user ? user->primaryEmailOrUsername() : QString();
}


//****************************************************************************************************************************************************
/// \param[in] userID The userID.
//****************************************************************************************************************************************************
void UserList::onUserChanged(QString const &userID) {
    int const index = this->rowOfUserID(userID);
    SPUser user;
    grpc::Status status = app().grpc().getUser(userID, user);
    QQmlEngine::setObjectOwnership(user.get(), QQmlEngine::CppOwnership);

    if ((!user) || (!status.ok())) {
        if (index >= 0) {  // user exists here but not in the go backend. we delete it.
            app().log().trace(QString("Removing user from user list: %1").arg(userID));
            this->removeUserAt(index);
        }
        return;
    }

    if (index < 0) {
        app().log().trace(QString("Adding user in user list: %1").arg(userID));
        this->appendUser(user);
        return;
    }

    app().log().trace(QString("Updating user in user list: %1").arg(userID));
    this->updateUserAtRow(index, *user);
}


//****************************************************************************************************************************************************
/// The only purpose of this function is to forward the toggleSplitModeFinished event received from gRPC to the
/// appropriate user.
///
/// \param[in] userID the userID.
//****************************************************************************************************************************************************
void UserList::onToggleSplitModeFinished(QString const &userID) {
    int const index = this->rowOfUserID(userID);
    if (index < 0) {
        app().log().error(QString("Received toggleSplitModeFinished event for unknown userID %1").arg(userID));
        return;
    }

    users_[index]->emitToggleSplitModeFinished();
}


//****************************************************************************************************************************************************
/// \return THe number of items in the list.
//****************************************************************************************************************************************************
int UserList::count() const {
    return users_.size();
}


//****************************************************************************************************************************************************
/// \param[in] userID The userID.
/// \param[in] usedBytes The used space, in bytes.
//****************************************************************************************************************************************************
void UserList::onUsedBytesChanged(QString const &userID, qint64 usedBytes) {
    int const index = this->rowOfUserID(userID);
    if (index < 0) {
        app().log().error(QString("Received usedBytesChanged event for unknown userID %1").arg(userID));
        return;
    }
    users_[index]->setUsedBytes(usedBytes);
}


//****************************************************************************************************************************************************
/// \param[in] userID The userID.
//****************************************************************************************************************************************************
void UserList::onSyncStarted(QString const &userID) {
    int const index = this->rowOfUserID(userID);
    if (index < 0) {
        app().log().error(QString("Received syncStarted event for unknown userID %1").arg(userID));
        return;
    }
    users_[index]->setIsSyncing(true);
}


//****************************************************************************************************************************************************
/// \param[in] userID The userID.
//****************************************************************************************************************************************************
void UserList::onSyncFinished(QString const &userID) {
    int const index = this->rowOfUserID(userID);
    if (index < 0) {
        app().log().error(QString("Received syncFinished event for unknown userID %1").arg(userID));
        return;
    }
    users_[index]->setIsSyncing(false);
}


//****************************************************************************************************************************************************
/// \param[in] userID The userID.
/// \param[in] progress The sync progress ratio.
/// \param[in] elapsedMs The elapsed sync time in milliseconds.
/// \param[in] remainingMs The remaining sync time in milliseconds.
//****************************************************************************************************************************************************
void UserList::onSyncProgress(QString const &userID, double progress, float elapsedMs, float remainingMs) {
    Q_UNUSED(elapsedMs)
    Q_UNUSED(remainingMs)
    int const index = this->rowOfUserID(userID);
    if (index < 0) {
        app().log().error(QString("Received syncProgress event for unknown userID %1").arg(userID));
        return;
    }
    users_[index]->setSyncProgress(progress);
}
