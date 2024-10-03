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


#ifndef BRIDGE_GUI_NATIVE_TRAY_ICON_H
#define BRIDGE_GUI_NATIVE_TRAY_ICON_H


//**********************************************************************************************************************
/// \brief A native tray icon.
//**********************************************************************************************************************
class TrayIcon: public QSystemTrayIcon {
Q_OBJECT
public: // typedef enum
    enum class State {
        Normal,
        Error,
        Warn,
        Update,
    }; ///< Enumeration for the state.

public: // data members
    TrayIcon(); ///< Default constructor.
    ~TrayIcon() override = default; ///< Destructor.
    TrayIcon(TrayIcon const&) = delete; ///< Disabled copy-constructor.
    TrayIcon(TrayIcon&&) = delete; ///< Disabled assignment copy-constructor.
    TrayIcon& operator=(TrayIcon const&) = delete; ///< Disabled assignment operator.
    TrayIcon& operator=(TrayIcon&&) = delete; ///< Disabled move assignment operator.
    void setState(State state, QString const& stateString, QString const &statusIconPath); ///< Set the state of the icon
    void showErrorPopupNotification(QString const& title, QString const &message); ///< Display a pop up notification.
    void showUserNotification(QString const& title, QString const &subtitle); ///< Display an OS pop up notification (without icon).
signals:
    void selectUser(QString const& userID, bool forceShowWindow); ///< Signal for selecting a user with a given userID

private slots:
    void onMenuAboutToShow(); ///< Slot called before the context menu is shown.
    void onUserClicked(); ///< Slot triggered when clicking on a user in the context menu.
    static void onActivated(QSystemTrayIcon::ActivationReason reason); ///< Slot for the activation of the system tray icon.
    void handleDPIChange(); ///< Handles DPI change.
    void setIcon(); ///< set the tray icon.
    void onIconRefreshTimer(); ///< Timer for icon refresh.

private: // member functions.
    void generateDotIcons(); ///< generate the colored dot icons used for user status.
    void generateStatusIcon(QString const &svgPath, QColor const& color); ///< Generate the status icon.
    void refreshContextMenu(); ///< Refresh the context menu.

private: // data members
    State state_ { State::Normal }; ///< The state of the tray icon.
    QString stateString_; ///< The current state string.
    std::unique_ptr<QMenu> menu_; ///< The context menu for the tray icon. Not owned by the tray icon.
    QIcon statusIcon_; ///< The path of the status icon displayed in the context menu.
    QIcon greenDot_; ///< The green dot icon.
    QIcon greyDot_; ///< The grey dot icon.
    QIcon orangeDot_; ///< The orange dot icon.
    QIcon const notificationErrorIcon_; ///< The error icon used for notifications.

    QTimer iconRefreshTimer_; ///< The timer used to periodically refresh the icon when DPI changes.
    QDateTime iconRefreshDeadline_; ///< The deadline for refreshing the icon
};



#endif //BRIDGE_GUI_NATIVE_TRAY_ICON_H
