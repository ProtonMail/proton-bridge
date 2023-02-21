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


#ifndef BRIDGE_GUI_USER_LIST_H
#define BRIDGE_GUI_USER_LIST_H


#include <bridgepp/User/User.h>
#include <bridgepp/Log/Log.h>
#include <bridgepp/GRPC/GRPCClient.h>


//****************************************************************************************************************************************************
/// \brief User list class.
//****************************************************************************************************************************************************
class UserList : public QAbstractListModel {
Q_OBJECT
public: // member functions.
    UserList(QObject *parent); ///< Default constructor.
    UserList(UserList const &other) = delete; ///< Disabled copy-constructor.
    UserList &operator=(UserList const &other) = delete; ///< Disabled assignment operator.
    ~UserList() override = default; ///< Destructor
    void connectGRPCEvents() const; ///< Connects gRPC event to the model.
    int rowCount(QModelIndex const &parent) const override; ///< Return the number of row in the model
    QVariant data(QModelIndex const &index, int role) const override; ///< Retrieve model data.
    void reset(QList<bridgepp::SPUser> const &users); ///< Replace the user list.
    int rowOfUserID(QString const &userID) const;
    void removeUserAt(int row); ///< Remove the user at a given row
    void appendUser(bridgepp::SPUser const &user); ///< Add a new user.
    void updateUserAtRow(int row, bridgepp::User const &user); ///< Update the user at given row.
    bridgepp::SPUser getUserWithID(QString const &userID) const; ///< Retrieve the user with the given ID.
    bridgepp::SPUser getUserWithUsernameOrEmail(QString const& username) const; ///< Retrieve the user with the given primary email address or username

    // the userCount property.
    Q_PROPERTY(int count READ count NOTIFY countChanged)
    int count() const; ///< The userCount property getter.

signals:
    void countChanged(int count); ///< Signal for the userCount property.

public:
    Q_INVOKABLE bridgepp::User *get(int row) const;
    Q_INVOKABLE QString primaryEmailOrUsername(QString const& userID) const; ///< Return the primary email or username of a user

public slots: ///< handler for signals coming from the gRPC service
    void onUserChanged(QString const &userID);
    void onToggleSplitModeFinished(QString const &userID);
    void onUsedBytesChanged(QString const &userID, qint64 usedBytes); ///< Slot for usedBytesChanged events.
    void onSyncStarted(QString const &userID); ///< Slot for syncStarted events.
    void onSyncFinished(QString const &userID); ///< Slot for syncFinished events.
    void onSyncProgress(QString const &userID, double progress, float elapsedMs, float remainingMs); ///< Slot for syncFinished events.
    
private: // data members
    QList<bridgepp::SPUser> users_; ///< The user list.
};


#endif // BRIDGE_GUI_USER_LIST_H
