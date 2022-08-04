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


#include "UsersTab.h"
#include "MainWindow.h"
#include "UserDialog.h"
#include <bridgepp/BridgeUtils.h>
#include <bridgepp/Exception/Exception.h>
#include <bridgepp/GRPC/EventFactory.h>


using namespace bridgepp;


//****************************************************************************************************************************************************
/// \param[in] parent The parent widget of the tab.
//****************************************************************************************************************************************************
UsersTab::UsersTab(QWidget *parent)
    : QWidget(parent)
    , users_(nullptr)
{
    ui_.setupUi(this);

    ui_.tableUserList->setModel(&users_);

    QItemSelectionModel *model = ui_.tableUserList->selectionModel();
    if (!model)
        throw Exception("Could not get user table selection model.");
    connect(model, &QItemSelectionModel::selectionChanged, this, &UsersTab::onSelectionChanged);

    ui_.tableUserList->setColumnWidth(0, 150);
    ui_.tableUserList->setColumnWidth(1, 250);
    ui_.tableUserList->setColumnWidth(2, 350);

    connect(ui_.buttonNewUser, &QPushButton::clicked, this, &UsersTab::onAddUserButton);
    connect(ui_.buttonEditUser, &QPushButton::clicked, this, &UsersTab::onEditUserButton);
    connect(ui_.tableUserList, &QTableView::doubleClicked, this, &UsersTab::onEditUserButton);
    connect(ui_.buttonRemoveUser, &QPushButton::clicked, this, &UsersTab::onRemoveUserButton);

    users_.append(randomUser());

    this->updateGUIState();
}


//****************************************************************************************************************************************************
//
//****************************************************************************************************************************************************
void UsersTab::onAddUserButton()
{
    SPUser user = randomUser();
    UserDialog dialog(user, this);
    if (QDialog::Accepted != dialog.exec())
        return;
    users_.append(user);
    GRPCService &grpc = app().grpc();
    if (grpc.isStreaming())
        grpc.sendEvent(newLoginFinishedEvent(user->id()));
}


//****************************************************************************************************************************************************
//
//****************************************************************************************************************************************************
void UsersTab::onEditUserButton()
{
    int index = selectedIndex();
    if ((index < 0) || (index >= users_.userCount()))
        return;

    SPUser user = this->selectedUser();
    UserDialog dialog(user, this);
    if (QDialog::Accepted != dialog.exec())
        return;

    users_.touch(index);
    GRPCService &grpc = app().grpc();
    if (grpc.isStreaming())
        grpc.sendEvent(newUserChangedEvent(user->id()));
}


//****************************************************************************************************************************************************
//
//****************************************************************************************************************************************************
void UsersTab::onRemoveUserButton()
{
    int index = selectedIndex();
    if ((index < 0) || (index >= users_.userCount()))
        return;

    SPUser const user = users_.userAtIndex(index);
    users_.remove(index);
    GRPCService &grpc = app().grpc();
    if (grpc.isStreaming())
        grpc.sendEvent(newUserChangedEvent(user->id()));
}


//****************************************************************************************************************************************************
//
//****************************************************************************************************************************************************
void UsersTab::onSelectionChanged(QItemSelection, QItemSelection)
{
    this->updateGUIState();
}


//****************************************************************************************************************************************************
//
//****************************************************************************************************************************************************
void UsersTab::updateGUIState()
{
    bool const hasSelectedUser = ui_.tableUserList->selectionModel()->hasSelection();
    ui_.buttonEditUser->setEnabled(hasSelectedUser);
    ui_.buttonRemoveUser->setEnabled(hasSelectedUser);
}


//****************************************************************************************************************************************************
//
//****************************************************************************************************************************************************
qint32 UsersTab::selectedIndex() const
{
    return ui_.tableUserList->selectionModel()->hasSelection() ? ui_.tableUserList->currentIndex().row() : -1;
}


//****************************************************************************************************************************************************
/// \return The selected user.
/// \return A null pointer if no user is selected.
//****************************************************************************************************************************************************
bridgepp::SPUser UsersTab::selectedUser()
{
    return users_.userAtIndex(this->selectedIndex());
}


//****************************************************************************************************************************************************
/// \return The list of users.
//****************************************************************************************************************************************************
UserTable &UsersTab::userTable()
{
    return users_;
}


//****************************************************************************************************************************************************
/// \param[in] userID The userID.
/// \return The user with the given userID.
/// \return A null pointer if the user is not in the list.
//****************************************************************************************************************************************************
bridgepp::SPUser UsersTab::userWithID(QString const &userID)
{
    return users_.userWithID(userID);
}


//****************************************************************************************************************************************************
/// \return true iff the next login attempt should trigger a username/password error.
//****************************************************************************************************************************************************
bool UsersTab::nextUserUsernamePasswordError() const
{
    return ui_.checkUsernamePasswordError->isChecked();
}


//****************************************************************************************************************************************************
/// \return true iff the next login attempt should trigger a free user error.
//****************************************************************************************************************************************************
bool UsersTab::nextUserFreeUserError() const
{
    return ui_.checkFreeUserError->isChecked();
}


//****************************************************************************************************************************************************
/// \return true iff the next login attempt will require 2FA.
//****************************************************************************************************************************************************
bool UsersTab::nextUserTFARequired() const
{
    return ui_.checkTFARequired->isChecked();
}


//****************************************************************************************************************************************************
/// \return true iff the next login attempt should trigger a 2FA error.
//****************************************************************************************************************************************************
bool UsersTab::nextUserTFAError() const
{
    return ui_.checkTFAError->isChecked();
}


//****************************************************************************************************************************************************
/// \return true iff the next login attempt should trigger a 2FA error with abort.
//****************************************************************************************************************************************************
bool UsersTab::nextUserTFAAbort() const
{
    return ui_.checkTFAAbort->isChecked();
}


//****************************************************************************************************************************************************
/// \return true iff the next login attempt will require a 2nd password.
//****************************************************************************************************************************************************
bool UsersTab::nextUserTwoPasswordsRequired() const
{
    return ui_.checkTwoPasswordsRequired->isChecked();
}


//****************************************************************************************************************************************************
/// \return true iff the next login attempt should trigger a 2nd password error.
//****************************************************************************************************************************************************
bool UsersTab::nextUserTwoPasswordsError() const
{
    return ui_.checkTwoPasswordsError->isChecked();
}


//****************************************************************************************************************************************************
/// \return true iff the next login attempt should trigger a 2nd password error with abort.
//****************************************************************************************************************************************************
bool UsersTab::nextUserTwoPasswordsAbort() const
{
    return ui_.checkTwoPasswordsAbort->isChecked();
}


//****************************************************************************************************************************************************
/// \return true iff the next login attempt should trigger a 2nd password error with abort.
//****************************************************************************************************************************************************
bool UsersTab::nextUserAlreadyLoggedIn() const
{
    return ui_.checkAlreadyLoggedIn->isChecked();
}


//****************************************************************************************************************************************************
/// \param[in] userID The userID.
/// \param[in] makeItActive Should split mode be activated.
//****************************************************************************************************************************************************
void UsersTab::setUserSplitMode(QString const &userID, bool makeItActive)
{
    qint32 const index = users_.indexOfUser(userID);
    SPUser const user = users_.userAtIndex(index);
    if (!user)
    {
        app().log().error(QString("%1 failed. unknown user %1").arg(__FUNCTION__, userID));
        return;
    }
    user->setSplitMode(makeItActive);
    users_.touch(index);
    MainWindow &mainWindow = app().mainWindow();
    mainWindow.sendDelayedEvent(newUserChangedEvent(userID));
    mainWindow.sendDelayedEvent(newToggleSplitModeFinishedEvent(userID));
}

//****************************************************************************************************************************************************
/// \param[in] userID The userID.
//****************************************************************************************************************************************************
void UsersTab::logoutUser(QString const &userID)
{
    qint32 const index = users_.indexOfUser(userID);
    SPUser const user = users_.userAtIndex(index);
    if (!user)
    {
        app().log().error(QString("%1 failed. unknown user %1").arg(__FUNCTION__, userID));
        return;
    }
    user->setLoggedIn(false);
    users_.touch(index);
    app().mainWindow().sendDelayedEvent(newUserChangedEvent(userID));
}


//****************************************************************************************************************************************************
/// \param[in] userID The userID.
//****************************************************************************************************************************************************
void UsersTab::removeUser(QString const &userID)
{
    qint32 const index = users_.indexOfUser(userID);
    SPUser const user = users_.userAtIndex(index);
    if (!user)
    {
        app().log().error(QString("%1 failed. unknown user %1").arg(__FUNCTION__, userID));
        return;
    }
    users_.remove(index);
    app().mainWindow().sendDelayedEvent(newUserChangedEvent(userID));
}


//****************************************************************************************************************************************************
/// \param[in] userID The userID.
/// \param[in] address The address.
//****************************************************************************************************************************************************
void UsersTab::configureUserAppleMail(QString const &userID, QString const &address)
{
    app().log().info(QString("Apple mail configuration was requested for user %1, address %2").arg(userID, address));

}
