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


#include "MainWindow.h"
#include <bridgepp/Log/Log.h>


using namespace bridgepp;


namespace {


//****************************************************************************************************************************************************
/// \param[in] level The log level.
/// \param[in] message The log message.
/// \param[in] logEdit The plain text edit widget that displays the log.
//****************************************************************************************************************************************************
void addEntryToLogEdit(bridgepp::Log::Level level, const QString &message, QPlainTextEdit &logEdit) {
    /// \todo This may cause performance issue when log grows big. A better alternative should be implemented.
    QString log = logEdit.toPlainText().trimmed();
    if (!log.isEmpty()) {
        log += "\n";
    }
    logEdit.setPlainText(log + Log::logEntryToString(level, QDateTime::currentDateTime(), message));
}


} // Anonymous namespace


//****************************************************************************************************************************************************
/// \param[in] parent The parent widget of the window.
//****************************************************************************************************************************************************
MainWindow::MainWindow(QWidget *parent)
    : QMainWindow(parent) {
    ui_.setupUi(this);
    ui_.tabTop->setCurrentIndex(0);
    ui_.tabBottom->setCurrentIndex(0);
    ui_.splitter->setStretchFactor(0, 0);
    ui_.splitter->setStretchFactor(1, 1);
    ui_.splitter->setSizes({ 100, 10000 });
    connect(&app().log(), &Log::entryAdded, this, &MainWindow::addLogEntry);
    connect(&app().bridgeGUILog(), &Log::entryAdded, this, &MainWindow::addBridgeGUILogEntry);
}


//****************************************************************************************************************************************************
/// \return A reference to the 'General' tab.
//****************************************************************************************************************************************************
SettingsTab &MainWindow::settingsTab() const {
    return *ui_.settingsTab;
}


//****************************************************************************************************************************************************
/// \return A reference to the users tab.
//****************************************************************************************************************************************************
UsersTab &MainWindow::usersTab() const {
    return *ui_.usersTab;
}


//****************************************************************************************************************************************************
/// \return A reference to the events tab.
//****************************************************************************************************************************************************
EventsTab& MainWindow::eventsTab() const {
    return *ui_.eventsTab;
}


//****************************************************************************************************************************************************
/// \return A reference to the knowledge base tab.
//****************************************************************************************************************************************************
KnowledgeBaseTab& MainWindow::knowledgeBaseTab() const {
    return *ui_.knowledgeBaseTab;
}


//****************************************************************************************************************************************************
/// \param[in] level The log level.
/// \param[in] message The log message
//****************************************************************************************************************************************************
void MainWindow::addLogEntry(bridgepp::Log::Level level, const QString &message) const {
    addEntryToLogEdit(level, message, *ui_.editLog);
}


//****************************************************************************************************************************************************
/// \param[in] level The log level.
/// \param[in] message The log message
//****************************************************************************************************************************************************
void MainWindow::addBridgeGUILogEntry(bridgepp::Log::Level level, const QString &message) const {
    addEntryToLogEdit(level, message, *ui_.editBridgeGUILog);
}


//****************************************************************************************************************************************************
/// \param[in] event The event.
//****************************************************************************************************************************************************
void MainWindow::sendDelayedEvent(SPStreamEvent const &event) const {
    QTimer::singleShot(this->eventsTab().eventDelayMs(), [event] { app().grpc().sendEvent(event); });
}


