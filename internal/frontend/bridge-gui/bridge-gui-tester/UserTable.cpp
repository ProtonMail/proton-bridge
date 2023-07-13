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


#include "UserTable.h"


using namespace bridgepp;


//****************************************************************************************************************************************************
/// \param[in] parent The parent object of the class
//****************************************************************************************************************************************************
UserTable::UserTable(QObject *parent)
    : QAbstractTableModel(parent) {

}


//****************************************************************************************************************************************************
/// \return The number of rows in the table.
//****************************************************************************************************************************************************
int UserTable::rowCount(QModelIndex const &) const {
    return users_.size();
}


//****************************************************************************************************************************************************
/// \return The number of columns in the table.
//****************************************************************************************************************************************************
int UserTable::columnCount(QModelIndex const &) const {
    return 4;
}


//****************************************************************************************************************************************************
/// \param[in] index The model index.
/// \param[in] role The role to retrieve data for.
/// \return The data for the role at the given index.
//****************************************************************************************************************************************************
QVariant UserTable::data(QModelIndex const &index, int role) const {
    int const row = index.row();
    if ((row < 0) || (row >= users_.size()) || (Qt::DisplayRole != role)) {
        return QVariant();
    }

    SPUser const user = users_[row];
    if (!user) {
        return QVariant();
    }

    switch (index.column()) {
    case 0:
        return user->property("username");
    case 1:
        return user->property("addresses").toStringList().join(" ");
    case 2:
        return User::stateToString(user->state());
    case 3:
        return user->property("id");
    default:
        return QVariant();
    }
}


//****************************************************************************************************************************************************
/// \param[in] section The section (column).
/// \param[in] orientation The orientation.
/// \param[in] role The role to retrieve data
//****************************************************************************************************************************************************
QVariant UserTable::headerData(int section, Qt::Orientation orientation, int role) const {
    if (Qt::DisplayRole != role) {
        return QAbstractTableModel::headerData(section, orientation, role);
    }

    if (Qt::Horizontal != orientation) {
        return QString();
    }

    switch (section) {
    case 0:
        return "UserName";
    case 1:
        return "Addresses";
    case 2:
        return "State";
    case 3:
        return "UserID";
    default:
        return QString();
    }
}


//****************************************************************************************************************************************************
/// \param[in] user The user to add.
//****************************************************************************************************************************************************
void UserTable::append(SPUser const &user) {
    qint32 const count = users_.size();
    this->beginInsertRows(QModelIndex(), count, count);
    users_.append(user);
    this->endInsertRows();
}


//****************************************************************************************************************************************************
/// \return The number of users in the table.
//****************************************************************************************************************************************************
qint32 UserTable::userCount() const {
    return users_.count();
}


//****************************************************************************************************************************************************
/// \param[in] index The index of the user in the list.
/// \return the user at the given index.
/// \return null if the index is out of bounds.
//****************************************************************************************************************************************************
bridgepp::SPUser UserTable::userAtIndex(qint32 index) {
    return isIndexValid(index) ? users_[index] : nullptr;
}


//****************************************************************************************************************************************************
/// \return The user with the given userID.
/// \return A null pointer if the user is not in the list.
//****************************************************************************************************************************************************
bridgepp::SPUser UserTable::userWithID(QString const &userID) {
    QList<SPUser>::const_iterator it = std::find_if(users_.constBegin(), users_.constEnd(), [&userID](SPUser const &user) -> bool {
        return user->id() == userID;
    });

    return it == users_.end() ? nullptr : *it;
}


//****************************************************************************************************************************************************
/// \param[in] username The username, or any email address attached to the account.
/// \return The user with the given username.
/// \return A null pointer if the user is not in the list.
//****************************************************************************************************************************************************
bridgepp::SPUser UserTable::userWithUsernameOrEmail(QString const &username) {
    QList<SPUser>::const_iterator it = std::find_if(users_.constBegin(), users_.constEnd(), [&username](SPUser const &user) -> bool {
        if (user->username().compare(username, Qt::CaseInsensitive) == 0) {
            return true;
        }
        return user->addresses().contains(username, Qt::CaseInsensitive);
    });

    return it == users_.end() ? nullptr : *it;
}


//****************************************************************************************************************************************************
/// \param[in] userID The userID.
/// \return the index of the user.
/// \return -1 if the user could not be found.
//****************************************************************************************************************************************************
qint32 UserTable::indexOfUser(QString const &userID) {
    QList<SPUser>::const_iterator it = std::find_if(users_.constBegin(), users_.constEnd(), [&userID](SPUser const &user) -> bool {
        return user->id() == userID;
    });

    return it == users_.end() ? -1 : it - users_.constBegin();
}


//****************************************************************************************************************************************************
/// \param[in] index The index of the user in the list.
//****************************************************************************************************************************************************
void UserTable::touch(qint32 index) {
    if (isIndexValid(index))
        emit { dataChanged(this->index(index, 0), this->index(index, this->columnCount(QModelIndex()) - 1)); }
}


//****************************************************************************************************************************************************
/// \param[in] userID The userID.
//****************************************************************************************************************************************************
void UserTable::touch(QString const &userID) {
    this->touch(this->indexOfUser(userID));
}


//****************************************************************************************************************************************************
/// \param[in] index The index of the user in the list.
//****************************************************************************************************************************************************
void UserTable::remove(qint32 index) {
    if (!isIndexValid(index)) {
        return;
    }

    this->beginRemoveRows(QModelIndex(), index, index);
    users_.removeAt(index);
    this->endRemoveRows();
}


//****************************************************************************************************************************************************
/// \return true iff the index is valid.
//****************************************************************************************************************************************************
bool UserTable::isIndexValid(qint32 index) const {
    return (index >= 0) && (index < users_.count());
}


//****************************************************************************************************************************************************
/// \return The user list.
//****************************************************************************************************************************************************
QList<bridgepp::SPUser> UserTable::users() const {
    return users_;
}

