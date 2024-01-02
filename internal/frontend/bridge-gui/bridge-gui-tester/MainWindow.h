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


#ifndef BRIDGE_GUI_TESTER_MAIN_WINDOW_H
#define BRIDGE_GUI_TESTER_MAIN_WINDOW_H


#include "ui_MainWindow.h"
#include "GRPCService.h"
#include <bridgepp/Log/Log.h>


//**********************************************************************************************************************
/// \brief Main window class
//**********************************************************************************************************************
class MainWindow : public QMainWindow {
Q_OBJECT
public: // member functions.
    explicit MainWindow(QWidget *parent); ///< Default constructor.
    MainWindow(MainWindow const &) = delete; ///< Disabled copy-constructor.
    MainWindow(MainWindow &&) = delete; ///< Disabled assignment copy-constructor.
    ~MainWindow() override = default; ///< Destructor.
    MainWindow &operator=(MainWindow const &) = delete; ///< Disabled assignment operator.
    MainWindow &operator=(MainWindow &&) = delete; ///< Disabled move assignment operator.

    SettingsTab &settingsTab() const; ///< Returns a reference the 'Settings' tab.
    UsersTab &usersTab() const; ///< Returns a reference to the 'Users' tab.
    EventsTab &eventsTab() const; ///< Returns a reference to the 'Events' tab.
    KnowledgeBaseTab &knowledgeBaseTab() const; ///< Returns a reference to the 'Knowledge Base' tab.

public slots:
    void sendDelayedEvent(bridgepp::SPStreamEvent const &event) const; ///< Sends a gRPC event after the delay specified in the UI. The call is non blocking.

private slots:
    void addLogEntry(bridgepp::Log::Level level, QString const &message) const; ///< Add an entry to the log.
    void addBridgeGUILogEntry(bridgepp::Log::Level level, const QString &message) const; ///< Add an entry to the log.

private:
    Ui::MainWindow ui_ {}; ///< The GUI for the window.
};


#endif // BRIDGE_GUI_TESTER_MAIN_WINDOW_H
