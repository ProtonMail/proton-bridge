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


#ifndef BRIDGE_GUI_USER_LIST_H
#define BRIDGE_GUI_USER_LIST_H

#include "User.h"

//****************************************************************************************************************************************************
/// \brief User list class.
//****************************************************************************************************************************************************
class UserList: public QAbstractListModel
{
    Q_OBJECT
public: // member functions.
    explicit UserList(QObject *parent = nullptr); ///< Default constructor.
    UserList(UserList const &other) = delete ; ///< Disabled copy-constructor.
    UserList& operator=(UserList const& other) = delete; ///< Disabled assignment operator.
    ~UserList() override = default; ///< Destructor
    void connectGRPCEvents() const; ///< Connects gRPC event to the model.
    int rowCount(QModelIndex const &parent) const override; ///< Return the number of row in the model
    QVariant data(QModelIndex const &index, int role) const override; ///< Retrieve model data.
    void reset(); ///< Reset the user list.
    void reset(QList<SPUser> const &users); ///< Replace the user list.
    int rowOfUserID(QString const &userID) const;
    void removeUserAt(int row); ///< Remove the user at a given row
    void appendUser(SPUser const& user); ///< Add a new user.
    void updateUserAtRow(int row, User const& user); ///< Update the user at given row.

    // the count property.
    Q_PROPERTY(int count READ count NOTIFY countChanged)
    int count() const; ///< The count property getter.

signals:
    void countChanged(int count); ///< Signal for the count property.

public:
    Q_INVOKABLE User* get(int row) const;

public slots: ///< handler for signals coming from the gRPC service
    void onUserChanged(QString const &userID);
    void onToggleSplitModeFinished(QString const &userID);

private: // data members
    QList<SPUser> users_; ///< The user list.
};


#endif // BRIDGE_GUI_USER_LIST_H
