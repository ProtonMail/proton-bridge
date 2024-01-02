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



#ifndef BRIDGE_GUI_TESTER_USER_TABLE_H
#define BRIDGE_GUI_TESTER_USER_TABLE_H


#include <bridgepp/User/User.h>

//****************************************************************************************************************************************************
/// \brief User table model class
//****************************************************************************************************************************************************

class UserTable : public QAbstractTableModel {
Q_OBJECT
public: // member functions.
    explicit UserTable(QObject *parent); ///< Default constructor.
    UserTable(UserTable const &) = delete; ///< Disabled copy-constructor.
    UserTable(UserTable &&) = delete; ///< Disabled assignment copy-constructor.
    ~UserTable() override = default; ///< Destructor.
    UserTable &operator=(UserTable const &) = delete; ///< Disabled assignment operator.
    UserTable &operator=(UserTable &&) = delete; ///< Disabled move assignment operator.
    qint32 userCount() const; ///< Return the number of users in the table.
    void append(bridgepp::SPUser const &user); ///< Append a user.
    bridgepp::SPUser userAtIndex(qint32 index); ///< Return the user at the given index.
    bridgepp::SPUser userWithID(QString const &userID); ///< Return the user with a given id.
    bridgepp::SPUser userWithUsernameOrEmail(QString const &username); ///< Return the user with a given username.
    qint32 indexOfUser(QString const &userID); ///< Return the index of a given User.
    void touch(qint32 index); ///< touch the user at a given index (indicates it has been modified).
    void touch(QString const& userID); ///< touch the user with the given userID (indicates it has been modified).
    void remove(qint32 index); ///< Remove the user at a given index.
    QList<bridgepp::SPUser> users() const; ///< Return a copy of the user list.

private: // data members.
    int rowCount(QModelIndex const &parent) const override; ///< Get the number of rows in the table.
    int columnCount(QModelIndex const &parent) const override; ///< Get the number of columns in the table.
    QVariant data(QModelIndex const &index, int role) const override; ///< Get the data for a role at a given index.
    QVariant headerData(int section, Qt::Orientation orientation, int role) const override; ///< Get header data.
    bool isIndexValid(qint32 index) const; ///< return true iff the index is valid.

public:
    QList<bridgepp::SPUser> users_;
};


#endif //BRIDGE_GUI_TESTER_USER_TABLE_H
