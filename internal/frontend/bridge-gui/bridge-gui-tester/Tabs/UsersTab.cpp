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
    , users_(nullptr) {
    ui_.setupUi(this);

    ui_.tableUserList->setModel(&users_);

    QItemSelectionModel *model = ui_.tableUserList->selectionModel();
    if (!model) {
        throw Exception("Could not get user table selection model.");
    }
    connect(model, &QItemSelectionModel::selectionChanged, this, &UsersTab::onSelectionChanged);

    ui_.tableUserList->setColumnWidth(0, 150);
    ui_.tableUserList->setColumnWidth(1, 250);
    ui_.tableUserList->setColumnWidth(2, 150);

    connect(ui_.buttonNewUser, &QPushButton::clicked, this, &UsersTab::onAddUserButton);
    connect(ui_.buttonEditUser, &QPushButton::clicked, this, &UsersTab::onEditUserButton);
    connect(ui_.tableUserList, &QTableView::doubleClicked, this, &UsersTab::onEditUserButton);
    connect(ui_.buttonRemoveUser, &QPushButton::clicked, this, &UsersTab::onRemoveUserButton);
    connect(ui_.buttonUserBadEvent, &QPushButton::clicked, this, &UsersTab::onSendUserBadEvent);
    connect(ui_.buttonImapLoginFailed, &QPushButton::clicked, this, &UsersTab::onSendIMAPLoginFailedEvent);
    connect(ui_.buttonUsedBytesChanged, &QPushButton::clicked, this, &UsersTab::onSendUsedBytesChangedEvent);
    connect(ui_.checkUsernamePasswordError, &QCheckBox::toggled, this, &UsersTab::updateGUIState);
    connect(ui_.checkSync, &QCheckBox::toggled, this, &UsersTab::onCheckSyncToggled);
    connect(ui_.sliderSync, &QSlider::valueChanged, this, &UsersTab::onSliderSyncValueChanged);

    users_.append(defaultUser());

    this->updateGUIState();
}


//****************************************************************************************************************************************************
//
//****************************************************************************************************************************************************
void UsersTab::onAddUserButton() {
    SPUser user = randomUser();
    UserDialog dialog(user, this);
    if (QDialog::Accepted != dialog.exec()) {
        return;
    }
    users_.append(user);
    GRPCService &grpc = app().grpc();
    if (grpc.isStreaming()) {
        grpc.sendEvent(newLoginFinishedEvent(user->id(), false));
    }
}


//****************************************************************************************************************************************************
//
//****************************************************************************************************************************************************
void UsersTab::onEditUserButton() {
    int index = selectedIndex();
    if ((index < 0) || (index >= users_.userCount())) {
        return;
    }

    SPUser user = this->selectedUser();
    UserDialog dialog(user, this);
    if (QDialog::Accepted != dialog.exec()) {
        return;
    }

    users_.touch(index);
    GRPCService &grpc = app().grpc();
    if (grpc.isStreaming()) {
        grpc.sendEvent(newUserChangedEvent(user->id()));
    }

    this->updateGUIState();
}


//****************************************************************************************************************************************************
//
//****************************************************************************************************************************************************
void UsersTab::onRemoveUserButton() {
    int index = selectedIndex();
    if ((index < 0) || (index >= users_.userCount())) {
        return;
    }

    SPUser const user = users_.userAtIndex(index);
    users_.remove(index);
    GRPCService &grpc = app().grpc();
    if (grpc.isStreaming()) {
        grpc.sendEvent(newUserChangedEvent(user->id()));
    }
}


//****************************************************************************************************************************************************
//
//****************************************************************************************************************************************************
void UsersTab::onSelectionChanged(QItemSelection, QItemSelection) {
    this->updateGUIState();
}


//****************************************************************************************************************************************************
//
//****************************************************************************************************************************************************
void UsersTab::onSendUserBadEvent() {
    SPUser const user = selectedUser();
    int const index = this->selectedIndex();

    if (!user) {
        app().log().error(QString("%1 failed. Unkown user.").arg(__FUNCTION__));
        return;
    }

    if (UserState::SignedOut == user->state()) {
        app().log().error(QString("%1 failed. User is already signed out").arg(__FUNCTION__));
    }

    GRPCService &grpc = app().grpc();
    if (grpc.isStreaming()) {
        QString const userID = user->id();
        grpc.sendEvent(newUserChangedEvent(userID));
        grpc.sendEvent(newUserBadEvent(userID, ui_.editUserBadEvent->text()));
    }

    this->updateGUIState();
}


//****************************************************************************************************************************************************
//
//****************************************************************************************************************************************************
void UsersTab::onSendUsedBytesChangedEvent() {
    SPUser const user = selectedUser();
    int const index = this->selectedIndex();

    if (!user) {
        app().log().error(QString("%1 failed. Unkown user.").arg(__FUNCTION__));
        return;
    }

    if (UserState::Connected != user->state()) {
        app().log().error(QString("%1 failed. User is not connected").arg(__FUNCTION__));
    }

    qint64 const usedBytes = qint64(ui_.spinUsedBytes->value());
    user->setUsedBytes(usedBytes);
    users_.touch(index);

    GRPCService &grpc = app().grpc();
    if (grpc.isStreaming()) {
        QString const userID = user->id();
        grpc.sendEvent(newUsedBytesChangedEvent(userID, usedBytes));
    }

    this->updateGUIState();
}


//****************************************************************************************************************************************************
//
//****************************************************************************************************************************************************
void UsersTab::onSendIMAPLoginFailedEvent() {
    GRPCService &grpc = app().grpc();
    if (grpc.isStreaming()) {
        grpc.sendEvent(newIMAPLoginFailedEvent(ui_.editIMAPLoginFailedUsername->text()));
    }

    this->updateGUIState();
}


//****************************************************************************************************************************************************
//
//****************************************************************************************************************************************************
void UsersTab::updateGUIState() {
    SPUser const user = selectedUser();
    bool const hasSelectedUser = user.get();
    UserState const state = user ? user->state() : UserState::SignedOut;

    ui_.buttonEditUser->setEnabled(hasSelectedUser);
    ui_.buttonRemoveUser->setEnabled(hasSelectedUser);
    ui_.groupBoxBadEvent->setEnabled(hasSelectedUser && (UserState::SignedOut != state));
    ui_.groupBoxUsedSpace->setEnabled(hasSelectedUser && (UserState::Connected == state));
    ui_.editUsernamePasswordError->setEnabled(ui_.checkUsernamePasswordError->isChecked());
    ui_.spinUsedBytes->setValue(user ? user->usedBytes() : 0.0);
    ui_.groupboxSync->setEnabled(user.get());

    if (user)
        ui_.editIMAPLoginFailedUsername->setText(user->primaryEmailOrUsername());

    QSignalBlocker b(ui_.checkSync);
    bool const syncing = user && user->isSyncing();
    ui_.checkSync->setChecked(syncing);
    b = QSignalBlocker(ui_.sliderSync);
    ui_.sliderSync->setEnabled(syncing);
    qint32 const progressPercent = syncing ? qint32(user->syncProgress() * 100.0f) : 0;
    ui_.sliderSync->setValue(progressPercent);
    ui_.labelSync->setText(syncing ? QString("%1%").arg(progressPercent) : "" );
}


//****************************************************************************************************************************************************
//
//****************************************************************************************************************************************************
qint32 UsersTab::selectedIndex() const {
    return ui_.tableUserList->selectionModel()->hasSelection() ? ui_.tableUserList->currentIndex().row() : -1;
}


//****************************************************************************************************************************************************
/// \return The selected user.
/// \return A null pointer if no user is selected.
//****************************************************************************************************************************************************
bridgepp::SPUser UsersTab::selectedUser() {
    return users_.userAtIndex(this->selectedIndex());
}


//****************************************************************************************************************************************************
/// \return The list of users.
//****************************************************************************************************************************************************
UserTable &UsersTab::userTable() {
    return users_;
}


//****************************************************************************************************************************************************
/// \param[in] userID The userID.
/// \return The user with the given userID.
/// \return A null pointer if the user is not in the list.
//****************************************************************************************************************************************************
bridgepp::SPUser UsersTab::userWithID(QString const &userID) {
    return users_.userWithID(userID);
}


//****************************************************************************************************************************************************
/// \param[in] username The username.
/// \return The user with the given username.
/// \return A null pointer if the user is not in the list.
//****************************************************************************************************************************************************
bridgepp::SPUser UsersTab::userWithUsernameOrEmail(QString const &username) {
    return users_.userWithUsernameOrEmail(username);
}


//****************************************************************************************************************************************************
/// \return true iff the next login attempt should trigger a username/password error.
//****************************************************************************************************************************************************
bool UsersTab::nextUserUsernamePasswordError() const {
    return ui_.checkUsernamePasswordError->isChecked();
}


//****************************************************************************************************************************************************
/// \return true iff the next login attempt should trigger a free user error.
//****************************************************************************************************************************************************
bool UsersTab::nextUserFreeUserError() const {
    return ui_.checkFreeUserError->isChecked();
}


//****************************************************************************************************************************************************
/// \return true iff the next login attempt will require 2FA.
//****************************************************************************************************************************************************
bool UsersTab::nextUserTFARequired() const {
    return ui_.checkTFARequired->isChecked();
}


//****************************************************************************************************************************************************
/// \return true iff the next login attempt should trigger a 2FA error.
//****************************************************************************************************************************************************
bool UsersTab::nextUserTFAError() const {
    return ui_.checkTFAError->isChecked();
}


//****************************************************************************************************************************************************
/// \return true iff the next login attempt should trigger a 2FA error with abort.
//****************************************************************************************************************************************************
bool UsersTab::nextUserTFAAbort() const {
    return ui_.checkTFAAbort->isChecked();
}


//****************************************************************************************************************************************************
/// \return true iff the next login attempt will require a 2nd password.
//****************************************************************************************************************************************************
bool UsersTab::nextUserTwoPasswordsRequired() const {
    return ui_.checkTwoPasswordsRequired->isChecked();
}


//****************************************************************************************************************************************************
/// \return true iff the next login attempt should trigger a 2nd password error.
//****************************************************************************************************************************************************
bool UsersTab::nextUserTwoPasswordsError() const {
    return ui_.checkTwoPasswordsError->isChecked();
}


//****************************************************************************************************************************************************
/// \return true iff the next login attempt should trigger a 2nd password error with abort.
//****************************************************************************************************************************************************
bool UsersTab::nextUserTwoPasswordsAbort() const {
    return ui_.checkTwoPasswordsAbort->isChecked();
}


//****************************************************************************************************************************************************
/// \return the message for the username/password error.
//****************************************************************************************************************************************************
QString UsersTab::usernamePasswordErrorMessage() const {
    return ui_.editUsernamePasswordError->text();
}


//****************************************************************************************************************************************************
/// \param[in] userID The userID.
/// \param[in] makeItActive Should split mode be activated.
//****************************************************************************************************************************************************
void UsersTab::setUserSplitMode(QString const &userID, bool makeItActive) {
    qint32 const index = users_.indexOfUser(userID);
    SPUser const user = users_.userAtIndex(index);
    if (!user) {
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
void UsersTab::logoutUser(QString const &userID) {
    qint32 const index = users_.indexOfUser(userID);
    SPUser const user = users_.userAtIndex(index);
    if (!user) {
        app().log().error(QString("%1 failed. unknown user %1").arg(__FUNCTION__, userID));
        return;
    }
    user->setState(UserState::SignedOut);
    users_.touch(index);
    app().mainWindow().sendDelayedEvent(newUserChangedEvent(userID));
}


//****************************************************************************************************************************************************
/// \param[in] userID The userID.
//****************************************************************************************************************************************************
void UsersTab::removeUser(QString const &userID) {
    qint32 const index = users_.indexOfUser(userID);
    SPUser const user = users_.userAtIndex(index);
    if (!user) {
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
void UsersTab::configureUserAppleMail(QString const &userID, QString const &address) {
    app().log().info(QString("Apple mail configuration was requested for user %1, address %2").arg(userID, address));
}


//****************************************************************************************************************************************************
/// \param[in] userID The userID.
/// \param[in] doResync Did the user request a resync?
//****************************************************************************************************************************************************
void UsersTab::processBadEventUserFeedback(QString const &userID, bool doResync) {
    app().log().info(QString("Feedback received for bad event: doResync = %1, userID = %2").arg(doResync ? "true" : "false", userID));
    if (doResync) {
        return; // we do not do any form of emulation for resync.
    }

    SPUser user = users_.userWithID(userID);
    if (!user) {
        app().log().error(QString("%1(): could not find user with id %1.").arg(__func__, userID));
    }

    user->setState(UserState::SignedOut);
    users_.touch(userID);
    GRPCService &grpc = app().grpc();
    if (grpc.isStreaming()) {
        grpc.sendEvent(newUserChangedEvent(userID));
    }

    this->updateGUIState();
}


//****************************************************************************************************************************************************
/// \param[in] checked Is the sync checkbox checked?
//****************************************************************************************************************************************************
void UsersTab::onCheckSyncToggled(bool checked) {
    SPUser const user = this->selectedUser();
    if ((!user) || (user->isSyncing() == checked)) {
        return;
    }

    user->setIsSyncing(checked);
    user->setSyncProgress(0.0);
    GRPCService &grpc = app().grpc();

    // we do not apply delay for these event.
    if (checked) {
        grpc.sendEvent(newSyncStartedEvent(user->id()));
        grpc.sendEvent(newSyncProgressEvent(user->id(), 0.0, 1, 1));
    } else {
        grpc.sendEvent(newSyncFinishedEvent(user->id()));
    }

    this->updateGUIState();
}


//****************************************************************************************************************************************************
/// \param[in] value The value for the slider.
//****************************************************************************************************************************************************
void UsersTab::onSliderSyncValueChanged(int value) {
    SPUser const user = this->selectedUser();
    if ((!user) || (!user->isSyncing()) || user->syncProgress() == value) {
        return;
    }

    double const progress = value / 100.0;
    user->setSyncProgress(progress);
    app().grpc().sendEvent(newSyncProgressEvent(user->id(), progress, 1, 1));  // we do not simulate elapsed & remaining.
    this->updateGUIState();
}
