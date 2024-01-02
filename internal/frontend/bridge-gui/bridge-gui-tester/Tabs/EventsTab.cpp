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


#include "EventsTab.h"
#include "GRPCService.h"
#include <bridgepp/GRPC/EventFactory.h>


using namespace bridgepp;


//****************************************************************************************************************************************************
/// \brief Connect an address error button to the generation of an address error event.
///
/// \param[in] button The error button.
/// \param[in] edit The edit containing the address.
/// \param[in] eventGenerator The factory function creating the event.
//****************************************************************************************************************************************************
void connectAddressError(QPushButton const* button, QLineEdit* edit, SPStreamEvent (*eventGenerator)(QString const&)) {
    QObject::connect(button, &QPushButton::clicked, [edit, eventGenerator]() { app().grpc().sendEvent(eventGenerator(edit->text())); });
}


//****************************************************************************************************************************************************
/// \param[in] parent The parent widget.
//****************************************************************************************************************************************************
EventsTab::EventsTab(QWidget* parent)
    : QWidget(parent) {
    ui_.setupUi(this);
    this->resetUI();

    connect(ui_.buttonInternetOn, &QPushButton::clicked, []() { app().grpc().sendEvent(newInternetStatusEvent(true)); });
    connect(ui_.buttonInternetOff, &QPushButton::clicked, []() { app().grpc().sendEvent(newInternetStatusEvent(false)); });
    connect(ui_.buttonShowMainWindow, &QPushButton::clicked, []() { app().grpc().sendEvent(newShowMainWindowEvent()); });
    connect(ui_.buttonNoKeychain, &QPushButton::clicked, []() { app().grpc().sendEvent(newHasNoKeychainEvent()); });
    connect(ui_.buttonAPICertIssue, &QPushButton::clicked, []() { app().grpc().sendEvent(newApiCertIssueEvent()); });
    connectAddressError(ui_.buttonAddressChanged, ui_.editAddressErrors, newAddressChangedEvent);
    connectAddressError(ui_.buttonAddressChangedLogout, ui_.editAddressErrors, newAddressChangedLogoutEvent);
    //connect(ui_.checkNextCacheChangeWillSucceed, &QCheckBox::toggled, this, &SettingsTab::updateGUIState);
    connect(ui_.buttonUpdateError, &QPushButton::clicked, [&]() {
        app().grpc().sendEvent(newUpdateErrorEvent(static_cast<grpc::UpdateErrorType>(ui_.comboUpdateError->currentIndex())));
    });
    connect(ui_.buttonUpdateManualReady, &QPushButton::clicked, [&] {
        app().grpc().sendEvent(newUpdateManualReadyEvent(ui_.editUpdateVersion->text()));
    });
    connect(ui_.buttonUpdateForce, &QPushButton::clicked, [&] {
        app().grpc().sendEvent(newUpdateForceEvent(ui_.editUpdateVersion->text()));
    });
    connect(ui_.buttonUpdateManualRestart, &QPushButton::clicked, []() { app().grpc().sendEvent(newUpdateManualRestartNeededEvent()); });
    connect(ui_.buttonUpdateSilentRestart, &QPushButton::clicked, []() { app().grpc().sendEvent(newUpdateSilentRestartNeededEvent()); });
    connect(ui_.buttonUpdateIsLatest, &QPushButton::clicked, []() { app().grpc().sendEvent(newUpdateIsLatestVersionEvent()); });
    connect(ui_.buttonUpdateCheckFinished, &QPushButton::clicked, []() { app().grpc().sendEvent(newUpdateCheckFinishedEvent()); });
    connect(ui_.buttonUpdateVersionChanged, &QPushButton::clicked, []() { app().grpc().sendEvent(newUpdateVersionChangedEvent()); });
}


//****************************************************************************************************************************************************
/// \return The delay to apply before sending automatically generated events.
//****************************************************************************************************************************************************
qint32 EventsTab::eventDelayMs() const {
    return ui_.spinEventDelay->value();
}


//****************************************************************************************************************************************************
///  \return The bug report results
//****************************************************************************************************************************************************
EventsTab::BugReportResult EventsTab::nextBugReportResult() const {
    return static_cast<BugReportResult>(ui_.comboBugReportResult->currentIndex());
}


//****************************************************************************************************************************************************
/// \return The reply for the next IsPortFree gRPC call.
//****************************************************************************************************************************************************
bool EventsTab::isPortFree() const {
    return ui_.checkIsPortFree->isChecked();
}

//****************************************************************************************************************************************************
/// \return The value for the 'Next Cache Change Will Succeed' check box.
//****************************************************************************************************************************************************
bool EventsTab::nextCacheChangeWillSucceed() const {
    return ui_.checkNextCacheChangeWillSucceed->isChecked();
}

//****************************************************************************************************************************************************
//
//****************************************************************************************************************************************************
void EventsTab::resetUI() const {
    ui_.comboBugReportResult->setCurrentIndex(0);
    ui_.checkIsPortFree->setChecked(true);
    ui_.checkNextCacheChangeWillSucceed->setChecked(true);
}
