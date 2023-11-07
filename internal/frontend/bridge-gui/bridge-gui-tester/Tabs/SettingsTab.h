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


#ifndef BRIDGE_GUI_TESTER_GENERAL_TAB_H
#define BRIDGE_GUI_TESTER_GENERAL_TAB_H


#include "Tab/ui_SettingsTab.h"


//****************************************************************************************************************************************************
/// \brief The 'General' tab of the main window.
//****************************************************************************************************************************************************
class SettingsTab : public QWidget {
Q_OBJECT
public: // data types.
    enum class TLSCertInstallResult {
        Success = 0,
        Canceled = 1,
        Failure = 2
    }; ///< Enumeration for the result of a TLS certificate installation.

    enum class BugReportResult {
        Success  = 0,
        Error = 1,
        DataSharingError = 2,
    }; ///< Enumeration for the result of bug report sending

public: // member functions.
    explicit SettingsTab(QWidget *parent = nullptr); ///< Default constructor.
    SettingsTab(SettingsTab const &) = delete; ///< Disabled copy-constructor.
    SettingsTab(SettingsTab &&) = delete; ///< Disabled assignment copy-constructor.
    ~SettingsTab() override = default; ///< Destructor.
    SettingsTab &operator=(SettingsTab const &) = delete; ///< Disabled assignment operator.
    SettingsTab &operator=(SettingsTab &&) = delete; ///< Disabled move assignment operator.

    QString bridgeVersion() const; ///< Get the Bridge version.
    QString os() const; ///< Return the OS string.
    QString currentEmailClient() const; ///< Return the content of the current email client
    void setGUIReady(bool ready); ///< Set the GUI as ready.
    bool showOnStartup() const; ///< Get the value for the 'Show On Startup' check.
    bool showSplashScreen() const; ///< Get the value for the 'Show Splash Screen' check.
    bool isAutostartOn() const; ///< Get the value for the 'Autostart' check.
    bool isBetaEnabled() const; ///< Get the value for the 'Beta Enabled' check.
    bool isAllMailVisible() const; ///< Get the value for the 'All Mail Visible' check.
    bool isTelemetryDisabled() const; ///< Get the value for the 'Disable Telemetry' check box.
    QString colorSchemeName() const; ///< Get the value of the 'Use Dark Theme' checkbox.
    qint32 eventDelayMs() const; ///< Get the delay for sending automatically generated events.
    QString logsPath() const; ///< Get the content of the 'Logs Path' edit.
    QString licensePath() const; ///< Get the content of the 'License Path' edit.
    QString releaseNotesPageLink() const; ///< Get the content of the 'Release Notes Page Link' edit.
    QString dependencyLicenseLink() const; ///< Get the content of the 'Dependency License Link' edit.
    QString landingPageLink() const; ///< Get the content of the 'Landing Page Link' edit.
    BugReportResult nextBugReportResult() const; ///< Get the value of the 'Next bug report result' combo box.
    bool isTLSCertificateInstalled() const; ///< Get the status of the 'TLS Certificate is installed' check box.
    TLSCertInstallResult nextTLSCertInstallResult() const; ///< Get the value of the 'Next TLS Certificate install result' combo box.
    bool nextTLSCertExportWillSucceed() const;  ///< Get the status of the 'Next TLS Cert export will succeed' check box.
    bool nextTLSKeyExportWillSucceed() const;  ///< Get the status of the 'Next TLS Key export will succeed' check box.
    QString hostname() const; ///< Get the value of the 'Hostname' edit.
    qint32 imapPort(); ///< Get the value of the IMAP port spin.
    qint32 smtpPort(); ///< Get the value of the SMTP port spin.
    bool useSSLForSMTP() const; ///< Get the value for the 'Use SSL for SMTP' check box.
    bool useSSLForIMAP() const; ///< Get the value for the 'Use SSL for IMAP' check box.
    bool isDoHEnabled() const; ///< Get the value for the 'DoH Enabled' check box.
    bool isPortFree() const; ///< Get the value for the "Is Port Free" check box.
    QString diskCachePath() const; ///< Get the value for the 'Disk Cache Path' edit.
    bool nextCacheChangeWillSucceed() const; ///< Get the value for the 'Next Cache Change will succeed' edit.
    bool isAutomaticUpdateOn() const; ///<Get the value for the 'Automatic Update' check box.

public slots:
    void updateGUIState(); ///< Update the GUI state.
    void setIsStreaming(bool isStreaming); ///< Set the isStreamingEvents value.
    void setClientPlatform(QString const &clientPlatform); ///< Set the client platform.
    void setIsAutostartOn(bool on); ///< Set the value for the 'Autostart' check box.
    void setIsBetaEnabled(bool enabled); ///< Set the value for the 'Beta Enabled' check box.
    void setIsAllMailVisible(bool visible); ///< Set the value for the 'All Mail Visible' check box.
    void setIsTelemetryDisabled(bool isDisabled); ///< Set the value for the 'Disable Telemetry' check box.
    void setColorSchemeName(QString const &name); ///< Set the value for the 'Use Dark Theme' check box.
    void setBugReport(QString const &osType, QString const &osVersion, QString const &emailClient, QString const &address, QString const &description,
        bool includeLogs); ///< Set the content of the bug report box.
    void installTLSCertificate(); ///< Install the TLS certificate.
    void exportTLSCertificates(QString const &folderPath); ///< Export the TLS certificates.
    void setMailServerSettings(qint32 imapPort, qint32 smtpPort, bool useSSLForIMAP, bool useSSLForSMTP); ///< Change the mail server settings.
    void setIsDoHEnabled(bool enabled); ///< Set the value for the 'DoH Enabled' check box.
    void setDiskCachePath(QString const &path); ///< Set the value for the 'Cache On Disk Enabled' check box.
    void setIsAutomaticUpdateOn(bool on); ///< Set the value for the 'Automatic Update' check box.

private: // member functions.
    void resetUI(); ///< Reset the widget.

private: // data members.
    Ui::SettingsTab ui_ {}; ///< The GUI for the tab
};


#endif //BRIDGE_GUI_TESTER_GENERAL_TAB_H
