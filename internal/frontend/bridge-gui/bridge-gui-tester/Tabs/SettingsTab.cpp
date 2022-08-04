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


#include "SettingsTab.h"
#include "GRPCService.h"
#include <bridgepp/GRPC/EventFactory.h>
#include <bridgepp/BridgeUtils.h>


using namespace bridgepp;


namespace
{
QString const colorSchemeDark = "dark"; ///< The dark color scheme name.
QString const colorSchemeLight = "light"; ///< THe light color scheme name.
}


//****************************************************************************************************************************************************
/// \param[in] parent The parent widget of the tab.
//****************************************************************************************************************************************************
SettingsTab::SettingsTab(QWidget *parent)
    : QWidget(parent)
{
    ui_.setupUi(this);

    connect(ui_.buttonInternetOn, &QPushButton::clicked, []() { app().grpc().sendEvent(newInternetStatusEvent(true)); });
    connect(ui_.buttonInternetOff, &QPushButton::clicked, []() { app().grpc().sendEvent(newInternetStatusEvent(false)); });
    connect(ui_.buttonShowMainWindow, &QPushButton::clicked, []() { app().grpc().sendEvent(newShowMainWindowEvent()); });
    connect(ui_.checkNextCacheChangeWillSucceed, &QCheckBox::toggled, this, &SettingsTab::updateGUIState);
    this->resetUI();
    this->updateGUIState();
}



//****************************************************************************************************************************************************
//
//****************************************************************************************************************************************************
void SettingsTab::updateGUIState()
{
    bool connected = app().grpc().isStreaming();
    for (QWidget *widget: { ui_.groupVersion, ui_.groupGeneral, ui_.groupMail, ui_.groupPaths, ui_.groupCache })
        widget->setEnabled(!connected);
    ui_.comboCacheError -> setEnabled(!ui_.checkNextCacheChangeWillSucceed->isChecked());
}


//****************************************************************************************************************************************************
/// \param[in] isStreaming Is the event stream on?
//****************************************************************************************************************************************************
void SettingsTab::setIsStreaming(bool isStreaming)
{
    ui_.labelStreamingValue->setText(isStreaming ? "Yes" : "No");
    this->updateGUIState();
}


//****************************************************************************************************************************************************
/// \param[in] clientPlatform The client platform.
//****************************************************************************************************************************************************
void SettingsTab::setClientPlatform(QString const &clientPlatform)
{
    ui_.labelClientPlatformValue->setText(clientPlatform);
}


//****************************************************************************************************************************************************
/// \return The version of Bridge
//****************************************************************************************************************************************************
QString SettingsTab::bridgeVersion() const
{
    return ui_.editVersion->text();
}


//****************************************************************************************************************************************************
/// \return The OS as a Go GOOS compatible value ("darwin", "linux" or "windows").
//****************************************************************************************************************************************************
QString SettingsTab::os() const
{
    return ui_.comboOS->currentText();
}


//****************************************************************************************************************************************************
/// \return The value for the 'Current Email Client' edit.
//****************************************************************************************************************************************************
QString SettingsTab::currentEmailClient() const
{
    return ui_.editCurrentEmailClient->text();
}


//****************************************************************************************************************************************************
/// \param[in] ready Is the GUI ready?
//****************************************************************************************************************************************************
void SettingsTab::setGUIReady(bool ready)
{
    this->updateGUIState();
    ui_.labelGUIReadyValue->setText(ready ? "Yes" : "No");
}


//****************************************************************************************************************************************************
/// \return true iff the 'Show On Startup' check box is checked.
//****************************************************************************************************************************************************
bool SettingsTab::showOnStartup() const
{
    return ui_.checkShowOnStartup->isChecked();
}


//****************************************************************************************************************************************************
/// \return true iff the 'Show Splash Screen' check box is checked.
//****************************************************************************************************************************************************
bool SettingsTab::showSplashScreen() const
{
    return ui_.checkShowSplashScreen->isChecked();
}


//****************************************************************************************************************************************************
/// \return true iff the 'Show Splash Screen' check box is checked.
//****************************************************************************************************************************************************
bool SettingsTab::isFirstGUIStart() const
{
    return ui_.checkIsFirstGUIStart->isChecked();
}


//****************************************************************************************************************************************************
/// \return true iff the 'Show Splash Screen' check box is checked.
//****************************************************************************************************************************************************
bool SettingsTab::isAutostartOn() const
{
    return ui_.checkAutostart->isChecked();
}


//****************************************************************************************************************************************************
/// \return true iff the 'Show Splash Screen' check box is checked.
//****************************************************************************************************************************************************
void SettingsTab::setIsAutostartOn(bool on)
{
    ui_.checkAutostart->setChecked(on);
}


//****************************************************************************************************************************************************
/// \return true if the 'Use Dark Theme' check box is checked.
//****************************************************************************************************************************************************
QString SettingsTab::colorSchemeName() const
{
    return ui_.checkDarkTheme->isChecked() ? colorSchemeDark : colorSchemeLight;
}


//****************************************************************************************************************************************************
/// \param[in] name True if the 'Use Dark Theme' check box should be checked.
//****************************************************************************************************************************************************
void SettingsTab::setColorSchemeName(QString const &name)
{
    ui_.checkDarkTheme->setChecked(name == colorSchemeDark);
}


//****************************************************************************************************************************************************
/// \return true if the 'Beta Enabled' check box is checked.
//****************************************************************************************************************************************************
bool SettingsTab::isBetaEnabled() const
{
    return ui_.checkBetaEnabled->isChecked();
}


//****************************************************************************************************************************************************
/// \param[in] enabled The new state for the 'Beta Enabled' check box.
//****************************************************************************************************************************************************
void SettingsTab::setIsBetaEnabled(bool enabled)
{
    ui_.checkBetaEnabled->setChecked(enabled);
}


//****************************************************************************************************************************************************
/// \return The delay to apply before sending automatically generated events.
//****************************************************************************************************************************************************
qint32 SettingsTab::eventDelayMs() const
{
    return ui_.spinEventDelay->value();
}


//****************************************************************************************************************************************************
/// \return The path
//****************************************************************************************************************************************************
QString SettingsTab::logsPath() const
{
    return ui_.editLogsPath->text();
}


//****************************************************************************************************************************************************
/// \return The path
//****************************************************************************************************************************************************
QString SettingsTab::licensePath() const
{
    return ui_.editLicensePath->text();
}


//****************************************************************************************************************************************************
/// \return The link.
//****************************************************************************************************************************************************
QString SettingsTab::releaseNotesPageLink() const
{
    return ui_.editReleaseNotesLink->text();
}


//****************************************************************************************************************************************************
/// \return The link.
//****************************************************************************************************************************************************
QString SettingsTab::dependencyLicenseLink() const
{
    return ui_.editDependencyLicenseLink->text();
}


//****************************************************************************************************************************************************
/// \return The link.
//****************************************************************************************************************************************************
QString SettingsTab::landingPageLink() const
{
    return ui_.editLandingPageLink->text();
}


//****************************************************************************************************************************************************
/// \param[in] osType The OS type.
/// \param[in] osVersion The OS version.
/// \param[in] emailClient The email client.
/// \param[in] address The email address.
/// \param[in] description The description.
/// \param[in] includeLogs Are the log included.
//****************************************************************************************************************************************************
void SettingsTab::setBugReport(QString const &osType, QString const &osVersion, QString const &emailClient, QString const &address,
    QString const &description, bool includeLogs)
{
    ui_.editOSType->setText(osType);
    ui_.editOSVersion->setText(osVersion);
    ui_.editEmailClient->setText(emailClient);
    ui_.editAddress->setText(address);
    ui_.editDescription->setPlainText(description);
    ui_.labelIncludeLogsValue->setText(includeLogs ? "Yes" : "No");
}


//****************************************************************************************************************************************************
/// \return The state of the check box.
//****************************************************************************************************************************************************
bool SettingsTab::nextBugReportWillSucceed() const
{
    return ui_.checkNextBugReportWillSucceed->isChecked();
}


//****************************************************************************************************************************************************
/// \return The value of the 'Hostname' edit.
//****************************************************************************************************************************************************
QString SettingsTab::hostname() const
{
    return ui_.editHostname->text();
}


//****************************************************************************************************************************************************
/// \return The value of the IMAP port spin box.
//****************************************************************************************************************************************************
qint32 SettingsTab::imapPort()
{
    return ui_.spinPortIMAP->value();
}


//****************************************************************************************************************************************************
/// \return The value of the SMTP port spin box.
//****************************************************************************************************************************************************
qint32 SettingsTab::smtpPort()
{
    return ui_.spinPortSMTP->value();
}


//****************************************************************************************************************************************************
/// \param[in] imapPort The IMAP port.
/// \param[in] smtpPort The SMTP port.
//****************************************************************************************************************************************************
void SettingsTab::changePorts(qint32 imapPort, qint32 smtpPort)
{
    ui_.spinPortIMAP->setValue(imapPort);
    ui_.spinPortSMTP->setValue(smtpPort);
}


//****************************************************************************************************************************************************
/// \return The state of the 'Use SSL for SMTP' check box.
//****************************************************************************************************************************************************
bool SettingsTab::useSSLForSMTP() const
{
    return ui_.checkUseSSLForSMTP->isChecked();
}


//****************************************************************************************************************************************************
/// \param[in] use The state of the 'Use SSL for SMTP' check box.
//****************************************************************************************************************************************************
void SettingsTab::setUseSSLForSMTP(bool use)
{
    ui_.checkUseSSLForSMTP->setChecked(use);
}


//****************************************************************************************************************************************************
/// \return The state of the the 'DoH enabled' check box.
//****************************************************************************************************************************************************
bool SettingsTab::isDoHEnabled() const
{
    return ui_.checkDoHEnabled->isChecked();
}


//****************************************************************************************************************************************************
/// \param[in] enabled The state of the 'DoH enabled' check box.
//****************************************************************************************************************************************************
void SettingsTab::setIsDoHEnabled(bool enabled)
{
    ui_.checkDoHEnabled->setChecked(enabled);
}


//****************************************************************************************************************************************************
/// \return The reply for the next IsPortFree gRPC call.
//****************************************************************************************************************************************************
bool SettingsTab::isPortFree() const
{
    return ui_.checkIsPortFree->isChecked();
}


//****************************************************************************************************************************************************
/// \return true iff cache on disk is enabled.
//****************************************************************************************************************************************************
bool SettingsTab::isCacheOnDiskEnabled() const
{
    return ui_.checkCacheOnDiskEnabled->isChecked();
}


//****************************************************************************************************************************************************
/// \param[in] enabled Is the cache on disk enabled?
//****************************************************************************************************************************************************
void SettingsTab::changeLocalCache(bool enabled, QString const &path)
{
    ui_.checkCacheOnDiskEnabled->setChecked(enabled);
    ui_.editDiskCachePath->setText(path);
}


//****************************************************************************************************************************************************
/// \return The disk cache path.
//****************************************************************************************************************************************************
QString SettingsTab::diskCachePath() const
{
    return ui_.editDiskCachePath->text();
}


//****************************************************************************************************************************************************
/// \return The value for the 'Next Cache Change Will Succeed' check box.
//****************************************************************************************************************************************************
bool SettingsTab::nextCacheChangeWillSucceed() const
{
    return ui_.checkNextCacheChangeWillSucceed->isChecked();
}


//****************************************************************************************************************************************************
/// \return The index of the selected cache error.
//****************************************************************************************************************************************************
qint32 SettingsTab::cacheError() const
{
    return ui_.comboCacheError->currentIndex();
}


//****************************************************************************************************************************************************
/// \return the value for the 'Automatic Update' check.
//****************************************************************************************************************************************************
bool SettingsTab::isAutomaticUpdateOn() const
{
    return ui_.checkAutomaticUpdate->isChecked();
}


//****************************************************************************************************************************************************
/// \param[in] on The value for the 'Automatic Update' check.
//****************************************************************************************************************************************************
void SettingsTab::setIsAutomaticUpdateOn(bool on)
{
    ui_.checkAutomaticUpdate->setChecked(on);
}


//****************************************************************************************************************************************************
//
//****************************************************************************************************************************************************
void SettingsTab::resetUI()
{
    this->setGUIReady(false);
    this->setIsStreaming(false);
    this->setClientPlatform("Unknown");

    ui_.editVersion->setText(BRIDGE_APP_VERSION);
    ui_.comboOS->setCurrentText(bridgepp::goos());
    ui_.editCurrentEmailClient->setText("Thunderbird/102.0.3");
    ui_.checkShowOnStartup->setChecked(true);
    ui_.checkShowSplashScreen->setChecked(false);
    ui_.checkIsFirstGUIStart->setChecked(false);
    ui_.checkAutostart->setChecked(true);
    ui_.checkBetaEnabled->setChecked(true);
    ui_.checkDarkTheme->setChecked(false);

    QString const tmpDir = QStandardPaths::writableLocation(QStandardPaths::TempLocation);

    QString const logsDir = QDir(tmpDir).absoluteFilePath("logs");
    QDir().mkpath(logsDir);
    ui_.editLogsPath->setText(QDir::toNativeSeparators(logsDir));

    QString const filePath = QDir(tmpDir).absoluteFilePath("LICENSE.txt");
    QFile file(filePath);
    if (!file.exists())
    {
        // we don't really care if it fails.
        file.open(QIODevice::WriteOnly | QIODevice::Text);
        file.write(QString("This is were the license should be.").toLocal8Bit());
        file.close();
    }
    ui_.editLicensePath->setText(filePath);

    ui_.editReleaseNotesLink->setText("https://en.wikipedia.org/wiki/Release_notes");
    ui_.editDependencyLicenseLink->setText("https://en.wikipedia.org/wiki/Dependency_relation");
    ui_.editLandingPageLink->setText("https://proton.me");

    ui_.editOSType->setText(QString());
    ui_.editOSVersion->setText(QString());
    ui_.editEmailClient->setText(QString());
    ui_.editAddress->setText(QString());
    ui_.editDescription->setPlainText(QString());
    ui_.labelIncludeLogsValue->setText(QString());
    ui_.checkNextBugReportWillSucceed->setChecked(true);

    ui_.editHostname->setText("localhost");
    ui_.spinPortIMAP->setValue(1143);
    ui_.spinPortSMTP->setValue(1025);
    ui_.checkUseSSLForSMTP->setChecked(false);
    ui_.checkDoHEnabled->setChecked(true);
    ui_.checkIsPortFree->setChecked(true);

    ui_.checkCacheOnDiskEnabled->setChecked(true);
    QString const cacheDir = QDir(tmpDir).absoluteFilePath("cache");
    QDir().mkpath(cacheDir);
    ui_.editDiskCachePath->setText(QDir::toNativeSeparators(cacheDir));
    ui_.checkNextCacheChangeWillSucceed->setChecked(true);
    ui_.comboCacheError->setCurrentIndex(0);

    ui_.checkAutomaticUpdate->setChecked(true);
}

