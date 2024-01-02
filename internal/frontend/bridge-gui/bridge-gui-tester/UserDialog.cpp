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


#include "UserDialog.h"


using namespace bridgepp;


//****************************************************************************************************************************************************
/// \param[in] user The user.
/// \param[in] parent The parent widget of the dialog.
//****************************************************************************************************************************************************
UserDialog::UserDialog(const bridgepp::SPUser &user, QWidget *parent)
    : QDialog(parent)
    , user_(user) {
    ui_.setupUi(this);

    connect(ui_.buttonOK, &QPushButton::clicked, this, &UserDialog::onOK);
    connect(ui_.buttonCancel, &QPushButton::clicked, this, &UserDialog::reject);

    ui_.editUserID->setText(user_->id());
    ui_.editUsername->setText(user_->username());
    ui_.editPassword->setText(user->password());
    ui_.editAddresses->setPlainText(user->addresses().join("\n"));
    ui_.editAvatarText->setText(user_->avatarText());
    this->setState(user->state());
    ui_.checkSplitMode->setChecked(user_->splitMode());
    ui_.spinUsedBytes->setValue(user->usedBytes());
    ui_.spinTotalBytes->setValue(user->totalBytes());
}


//****************************************************************************************************************************************************
//
//****************************************************************************************************************************************************
void UserDialog::onOK() {
    user_->setID(ui_.editUserID->text());
    user_->setUsername(ui_.editUsername->text());
    user_->setPassword(ui_.editPassword->text());
    user_->setAddresses(ui_.editAddresses->toPlainText().split(QRegularExpression(R"(\s+)"), Qt::SkipEmptyParts));
    user_->setAvatarText(ui_.editAvatarText->text());
    user_->setState(this->state());
    user_->setSplitMode(ui_.checkSplitMode->isChecked());
    user_->setUsedBytes(static_cast<float>(ui_.spinUsedBytes->value()));
    user_->setTotalBytes(static_cast<float>(ui_.spinTotalBytes->value()));

    this->accept();
}


//****************************************************************************************************************************************************
/// \return The user state  that is currently selected in the dialog.
//****************************************************************************************************************************************************
UserState UserDialog::state() const {
    return static_cast<UserState>(ui_.comboState->currentIndex());
}


//****************************************************************************************************************************************************
/// \param[in] state The user state to select in the dialog.
//****************************************************************************************************************************************************
void UserDialog::setState(UserState state) const {
    ui_.comboState->setCurrentIndex(static_cast<qint32>(state));
}
