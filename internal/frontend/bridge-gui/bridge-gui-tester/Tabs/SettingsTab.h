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


#ifndef BRIDGE_GUI_TESTER_GENERAL_TAB_H
#define BRIDGE_GUI_TESTER_GENERAL_TAB_H


#include "Tab/ui_SettingsTab.h"


//****************************************************************************************************************************************************
/// \brief The 'General' tab of the main window.
//****************************************************************************************************************************************************
class SettingsTab : public QWidget
{
Q_OBJECT
public: // member functions.
    explicit SettingsTab(QWidget *parent = nullptr); ///< Default constructor.
    SettingsTab(SettingsTab const &) = delete; ///< Disabled copy-constructor.
    SettingsTab(SettingsTab &&) = delete; ///< Disabled assignment copy-constructor.
    ~SettingsTab() = default; ///< Destructor.
    SettingsTab &operator=(SettingsTab const &) = delete; ///< Disabled assignment operator.
    SettingsTab &operator=(SettingsTab &&) = delete; ///< Disabled move assignment operator.

    QString bridgeVersion() const; ///< Get the Bridge version.
    QString os() const; ///< Return the OS string.
    QString currentEmailClient() const; ///< Return the content of the current email client
    void setGUIReady(bool ready); ///< Set the GUI as ready.
    bool showOnStartup() const; ///< Get the value for the 'Show On Startup' check.
    bool showSplashScreen() const; ///< Get the value for the 'Show Splash Screen' check.
    bool isFirstGUIStart() const; ///< Get the value for the 'Is First GUI Start' check.
    bool isAutostartOn() const; ///< Get the value for the 'Autostart' check.
    bool isBetaEnabled() const; ///< Get the value for the 'Beta Enabled' check.
    QString colorSchemeName() const; ///< Get the value of the 'Use Dark Theme' checkbox.
    qint32 eventDelayMs() const; ///< Get the delay for sending automatically generated events.
    QString logsPath() const; ///< Get the content of the 'Logs Path' edit.
    QString licensePath() const; ///< Get the content of the 'License Path' edit.
    QString releaseNotesPageLink() const; ///< Get the content of the 'Release Notes Page Link' edit.
    QString dependencyLicenseLink() const; ///< Get the content of the 'Dependency License Link' edit.
    QString landingPageLink() const; ///< Get the content of the 'Landing Page Link' edit.
    bool nextBugReportWillSucceed() const; ///< Get the status of the 'Next Bug Report Will Fail' check box.
    QString hostname() const; ///< Get the value of the 'Hostname' edit.
    qint32 imapPort(); ///< Get the value of the IMAP port spin.
    qint32 smtpPort(); ///< Get the value of the SMTP port spin.
    bool useSSLForSMTP() const; ///< Get the value for the 'Use SSL for SMTP' check box.
    bool isDoHEnabled() const; ///< Get the value for the 'DoH Enabled' check box.
    bool isPortFree() const; ///< Get the value for the "Is Port Free" check box.
    bool isCacheOnDiskEnabled() const; ///< get the value for the 'Cache On Disk Enabled' check box.
    QString diskCachePath() const; ///< Get the value for the 'Disk Cache Path' edit.
    bool nextCacheChangeWillSucceed() const; ///< Get the value for the 'Next Cache Change will succeed' edit.
    qint32 cacheError() const; ///< Return the index of the selected cache error.
    bool isAutomaticUpdateOn() const; ///<Get the value for the 'Automatic Update' check box.

public: // slots
    void updateGUIState(); ///< Update the GUI state.
    void setIsStreaming(bool isStreaming); ///< Set the isStreamingEvents value.
    void setClientPlatform(QString const &clientPlatform); ///< Set the client platform.
    void setIsAutostartOn(bool on); ///< Set the value for the 'Autostart' check.
    void setIsBetaEnabled(bool enabled); ///< Get the value for the 'Beta Enabled' check.
    void setColorSchemeName(QString const &name); ///< Set the value for the 'Use Dark Theme' checkbox.
    void setBugReport(QString const &osType, QString const &osVersion, QString const &emailClient, QString const &address, QString const &description,
        bool includeLogs); ///< Set the content of the bug report box.
    void changePorts(qint32 imapPort, qint32 smtpPort); ///< Change the IMAP and SMTP ports.
    void setUseSSLForSMTP(bool use); ///< Set the value for the 'Use SSL for SMTP' check box.
    void setIsDoHEnabled(bool enabled); ///< Set the value for the 'DoH Enabled' check box.
    void changeLocalCache(bool enabled, QString const &path); ///< Set the value for the 'Cache On Disk Enabled' check box.
    void setIsAutomaticUpdateOn(bool on); ///< Set the value for the 'Automatic Update' check box.

private: // member functions.
    void resetUI(); ///< Reset the widget.

private: // data members.
    Ui::SettingsTab ui_; ///< The GUI for the tab
};


#endif //BRIDGE_GUI_TESTER_GENERAL_TAB_H
