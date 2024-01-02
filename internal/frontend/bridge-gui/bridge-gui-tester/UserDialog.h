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


#ifndef BRIDGE_GUI_TESTER_USER_DIALOG_H
#define BRIDGE_GUI_TESTER_USER_DIALOG_H


#include "ui_UserDialog.h"
#include <bridgepp/User/User.h>


//****************************************************************************************************************************************************
/// \brief User dialog class.
//****************************************************************************************************************************************************
class UserDialog : public QDialog {
Q_OBJECT
public: // member functions.
    UserDialog(const bridgepp::SPUser &user, QWidget *parent); ///< Default constructor.
    UserDialog(UserDialog const &) = delete; ///< Disabled copy-constructor.
    UserDialog(UserDialog &&) = delete; ///< Disabled assignment copy-constructor.
    ~UserDialog() override = default; ///< Destructor.
    UserDialog &operator=(UserDialog const &) = delete; ///< Disabled assignment operator.
    UserDialog &operator=(UserDialog &&) = delete; ///< Disabled move assignment operator.

private: // member functions
    bridgepp::UserState state() const; ///< Get the user state selected in the dialog.
    void setState(bridgepp::UserState state) const; ///< Set the user state selected in the dialog

private slots:
    void onOK(); ///< Slot for the OK button.

private:
    Ui::UserDialog ui_ {}; ///< The UI for the dialog.
    bridgepp::SPUser user_; ///< The user
};


#endif //BRIDGE_GUI_TESTER_USER_DIALOG_H
