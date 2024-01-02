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


#ifndef BRIDGE_GUI_TESTER_GRPC_QT_PROXY_H
#define BRIDGE_GUI_TESTER_GRPC_QT_PROXY_H


#include <bridgepp/GRPC/GRPCUtils.h>


//****************************************************************************************************************************************************
/// \brief Proxy object used by the gRPC service (which does not inherit QObject) to use the Qt Signal/Slot system.
//****************************************************************************************************************************************************
class GRPCQtProxy : public QObject {
Q_OBJECT
public: // member functions.
    GRPCQtProxy(); ///< Default constructor.
    GRPCQtProxy(GRPCQtProxy const &) = delete; ///< Disabled copy-constructor.
    GRPCQtProxy(GRPCQtProxy &&) = delete; ///< Disabled assignment copy-constructor.
    ~GRPCQtProxy() override = default; ///< Destructor.
    GRPCQtProxy &operator=(GRPCQtProxy const &) = delete; ///< Disabled assignment operator.
    GRPCQtProxy &operator=(GRPCQtProxy &&) = delete; ///< Disabled move assignment operator.

    void connectSignals() const; // connect the signals to the main window.
    void sendDelayedEvent(bridgepp::SPStreamEvent const &event); ///< Sends a delayed stream event.
    void setIsAutostartOn(bool on); ///< Forwards a SetIsAutostartOn call via a Qt signal.
    void setIsBetaEnabled(bool enabled); ///< Forwards a SetIsBetaEnabled call via a Qt signal.
    void setIsAllMailVisible(bool visible); ///< Forwards a SetIsAllMailVisible call via a Qt signal.
    void setIsTelemetryDisabled(bool isDisabled); ///< Forwards a SetIsTelemetryDisabled call via a Qt signal.
    void setColorSchemeName(QString const &name); ///< Forward a SetColorSchemeName call via a Qt Signal
    void reportBug(QString const &osType, QString const &osVersion, QString const &emailClient, QString const &address,
        QString const &description, bool includeLogs); ///< Forwards a ReportBug call via a Qt signal.
    void requestKnowledgeBaseSuggestions(QString const &userInput); ///< Forwards a RequestKnowledgeBaseSuggestions call via a Qt signal.
    void installTLSCertificate(); ///< Forwards a InstallTLScertificate call via a Qt signal.
    void exportTLSCertificates(QString const &folderPath); //< Forward an 'ExportTLSCertificates' call via a Qt signal.
    void setIsStreaming(bool isStreaming); ///< Forward a isStreaming internal messages via a Qt signal.
    void setClientPlatform(QString const &clientPlatform); ///< Forward a setClientPlatform call via a Qt signal.
    void setMailServerSettings(qint32 imapPort, qint32 smtpPort, bool useSSLForIMAP, bool useSSLForSMTP); ///< Forwards a setMailServerSettings' call via a Qt signal.
    void setIsDoHEnabled(bool enabled); ///< Forwards a setIsDoHEnabled call via a Qt signal.
    void setDiskCachePath(QString const &path); ///< Forwards a setDiskCachePath call via a Qt signal.
    void setIsAutomaticUpdateOn(bool on); ///< Forwards a SetIsAutomaticUpdateOn call via a Qt signal.
    void setUserSplitMode(QString const &userID, bool makeItActive); ///< Forwards a setUserSplitMode call via a Qt signal.
    void sendBadEventUserFeedback(QString const &userID, bool doResync); ///< Forwards a sendBadEventUserFeedback call via a Qt signal.
    void logoutUser(QString const &userID); ///< Forwards a logoutUser call via a Qt signal.
    void removeUser(QString const &userID); ///< Forwards a removeUser call via a Qt signal.
    void configureUserAppleMail(QString const &userID, QString const &address); ///< Forwards a configureUserAppleMail call via a Qt signal.

signals:
    void delayedEventRequested(bridgepp::SPStreamEvent const &event); ///< Signal for sending a delayed event. delayed is set in the UI.
    void setIsAutostartOnReceived(bool on); ///< Forwards a SetIsAutostartOn call via a Qt signal.
    void setIsBetaEnabledReceived(bool enabled); ///< Forwards a SetIsBetaEnabled call via a Qt signal.
    void setIsAllMailVisibleReceived(bool enabled); ///< Forwards a SetIsBetaEnabled call via a Qt signal.
    void setIsTelemetryDisabledReceived(bool isDisabled); ///< Forwards a SetIsTelemetryDisabled call via a Qt signal.
    void setColorSchemeNameReceived(QString const &name); ///< Forward a SetColorScheme call via a Qt Signal
    void reportBugReceived(QString const &osType, QString const &osVersion, QString const &emailClient, QString const &address,
        QString const &description, bool includeLogs); ///< Signal for the ReportBug gRPC call
    void requestKnowledgeBaseSuggestionsReceived(QString const &userInput); ///< Signal for the RequestKnowledgeBaseSuggestions gRPC call.
    void installTLSCertificateReceived(); ///< Signal for the InstallTLSCertificate gRPC call.
    void exportTLSCertificatesReceived(QString const &folderPath); ///< Signal for the ExportTLSCertificates gRPC call.
    void setIsStreamingReceived(bool isStreaming); ///< Signal for the IsStreaming internal message.
    void setClientPlatformReceived(QString const &clientPlatform); ///< Signal for the SetClientPlatform gRPC call.
    void setMailServerSettingsReceived(qint32 imapPort, qint32 smtpPort, bool useSSLForIMAP, bool userSSLForSMTP); ///< Signal for the SetMailServerSettings gRPC call.
    void setIsDoHEnabledReceived(bool enabled); ///< Signal for the SetIsDoHEnabled gRPC call.
    void setDiskCachePathReceived(QString const &path); ///< Signal for the setDiskCachePath gRPC call.
    void setIsAutomaticUpdateOnReceived(bool on); ///< Signal for the SetIsAutomaticUpdateOn gRPC call.
    void setUserSplitModeReceived(QString const &userID, bool makeItActive); ///< Signal for the SetUserSplitModeReceived gRPC call.
    void sendBadEventUserFeedbackReceived(QString const &userID, bool doResync); ///< Signal for the SendBadEventUserFeedback gRPC call.
    void logoutUserReceived(QString const &userID); ///< Signal for the LogoutUserReceived gRPC call.
    void removeUserReceived(QString const &userID); ///< Signal for the RemoveUserReceived gRPC call.
    void configureUserAppleMailReceived(QString const &userID, QString const &address); ///< Signal for the ConfigureAppleMail gRPC call.
};


#endif //BRIDGE_GUI_TESTER_GRPC_QT_PROXY_H
