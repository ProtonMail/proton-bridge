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


#ifndef BRIDGE_GUI_TESTER_USERS_TAB_H
#define BRIDGE_GUI_TESTER_USERS_TAB_H


#include "Tabs/ui_UsersTab.h"
#include "UserTable.h"


//****************************************************************************************************************************************************
/// \brief The 'Users' tab of the main window.
//****************************************************************************************************************************************************
class UsersTab : public QWidget
{
Q_OBJECT
public: // member functions.
    explicit UsersTab(QWidget *parent = nullptr); ///< Default constructor.
    UsersTab(UsersTab const &) = delete; ///< Disabled copy-constructor.
    UsersTab(UsersTab &&) = delete; ///< Disabled assignment copy-constructor.
    ~UsersTab() override = default; ///< Destructor.
    UsersTab &operator=(UsersTab const &) = delete; ///< Disabled assignment operator.
    UsersTab &operator=(UsersTab &&) = delete; ///< Disabled move assignment operator.
    UserTable &userTable(); ///< Returns a reference to the user table.
    bridgepp::SPUser userWithID(QString const &userID); ///< Get the user with the given ID.
    bool nextUserUsernamePasswordError() const; ///< Check if next user login should trigger a username/password error.
    bool nextUserFreeUserError() const; ///< Check if next user login should trigger a Free user error.
    bool nextUserTFARequired() const; ///< Check if next user login should requires 2FA.
    bool nextUserTFAError() const; ///< Check if next user login should trigger 2FA error
    bool nextUserTFAAbort() const; ///< Check if next user login should trigger 2FA abort.
    bool nextUserTwoPasswordsRequired() const; ///< Check if next user login requires 2nd password
    bool nextUserTwoPasswordsError() const; ///< Check if next user login should trigger 2nd password error.
    bool nextUserTwoPasswordsAbort() const; ///< Check if next user login should trigger 2nd password abort.
    bool nextUserAlreadyLoggedIn() const; ///< Check if next user login should report user as already logged in.

public slots:
    void setUserSplitMode(QString const &userID, bool makeItActive); ///< Slot for the split mode.
    void logoutUser(QString const &userID); ///< slot for the logging out of a user.
    void removeUser(QString const &userID); ///< Slot for the removal of a user.
    void configureUserAppleMail(QString const &userID, QString const &address); ///< Slot for the configuration of Apple mail.

private slots:
    void onAddUserButton(); ///< Add a user to the user list.
    void onEditUserButton(); ///< Edit the currently selected user.
    void onRemoveUserButton(); ///< Remove the currently selected user.
    void onSelectionChanged(QItemSelection, QItemSelection); ///< Slot for the change of the selection.

private: // member functions.
    void updateGUIState(); ///< Update the GUI state.
    qint32 selectedIndex() const; ///< Get the index of the selected row.
    bridgepp::SPUser selectedUser(); ///< Get the selected user.

private: // data members.
    Ui::UsersTab ui_ {}; ///< The UI for the tab.
    UserTable users_; ///< The User list.
};


#endif //BRIDGE_GUI_TESTER_USERS_TAB_H
