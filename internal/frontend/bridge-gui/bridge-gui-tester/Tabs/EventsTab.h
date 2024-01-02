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


#ifndef BRIDGE_GUI_TESTER_EVENTS_TAB_H
#define BRIDGE_GUI_TESTER_EVENTS_TAB_H


#include "ui_EventsTab.h"


//****************************************************************************************************************************************************
/// \brief Events tabs
//****************************************************************************************************************************************************
class EventsTab: public QWidget {
    Q_OBJECT

public: // data types
enum class BugReportResult {
    Success  = 0,
    Error = 1,
    DataSharingError = 2,
}; ///< Enumeration for the result of bug report sending

public: // member functions.
    explicit EventsTab(QWidget *parent = nullptr); ///< Default constructor.
    EventsTab(EventsTab const&) = delete; ///< Disabled copy-constructor.
    EventsTab(EventsTab&&) = delete; ///< Disabled assignment copy-constructor.
    ~EventsTab() override = default; ///< Destructor.
    EventsTab& operator=(EventsTab const&) = delete; ///< Disabled assignment operator.
    EventsTab& operator=(EventsTab&&) = delete; ///< Disabled move assignment operator.

    qint32 eventDelayMs() const; ///< Get the delay for sending automatically generated events.
    BugReportResult nextBugReportResult() const; ///< Get the value of the 'Next bug report result' combo box.
    bool isPortFree() const; ///< Get the value for the "Is Port Free" check box.
    bool nextCacheChangeWillSucceed() const; ///< Get the value for the 'Next Cache Change will succeed' edit.

    void resetUI() const; ///< Resets the UI.

private: // data members
    Ui::EventsTab ui_ {}; ///< The UI for the widget.
};


#endif