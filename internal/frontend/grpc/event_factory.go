// Copyright (c) 2024 Proton AG
//
// This file is part of Proton Mail Bridge.Bridge.
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

package grpc

import (
	"github.com/ProtonMail/proton-bridge/v3/internal/kb"
	"github.com/bradenaw/juniper/xslices"
)

func NewInternetStatusEvent(connected bool) *StreamEvent {
	return appEvent(&AppEvent{Event: &AppEvent_InternetStatus{InternetStatus: &InternetStatusEvent{Connected: connected}}})
}

func NewToggleAutostartFinishedEvent() *StreamEvent {
	return appEvent(&AppEvent{Event: &AppEvent_ToggleAutostartFinished{ToggleAutostartFinished: &ToggleAutostartFinishedEvent{}}})
}

func NewResetFinishedEvent() *StreamEvent {
	return appEvent(&AppEvent{Event: &AppEvent_ResetFinished{ResetFinished: &ResetFinishedEvent{}}})
}

func NewReportBugFinishedEvent() *StreamEvent {
	return appEvent(&AppEvent{Event: &AppEvent_ReportBugFinished{ReportBugFinished: &ReportBugFinishedEvent{}}})
}

func NewReportBugSuccessEvent() *StreamEvent {
	return appEvent(&AppEvent{Event: &AppEvent_ReportBugSuccess{ReportBugSuccess: &ReportBugSuccessEvent{}}})
}

func NewReportBugErrorEvent() *StreamEvent {
	return appEvent(&AppEvent{Event: &AppEvent_ReportBugError{ReportBugError: &ReportBugErrorEvent{}}})
}

func NewReportBugFallbackEvent() *StreamEvent {
	return appEvent(&AppEvent{Event: &AppEvent_ReportBugFallback{ReportBugFallback: &ReportBugFallbackEvent{}}})
}

func NewCertInstallSuccessEvent() *StreamEvent {
	return appEvent(&AppEvent{Event: &AppEvent_CertificateInstallSuccess{CertificateInstallSuccess: &CertificateInstallSuccessEvent{}}})
}

func NewCertInstallCanceledEvent() *StreamEvent {
	return appEvent(&AppEvent{Event: &AppEvent_CertificateInstallCanceled{CertificateInstallCanceled: &CertificateInstallCanceledEvent{}}})
}

func NewCertInstallFailedEvent() *StreamEvent {
	return appEvent(&AppEvent{Event: &AppEvent_CertificateInstallFailed{CertificateInstallFailed: &CertificateInstallFailedEvent{}}})
}

func NewShowMainWindowEvent() *StreamEvent {
	return appEvent(&AppEvent{Event: &AppEvent_ShowMainWindow{ShowMainWindow: &ShowMainWindowEvent{}}})
}

func NewRequestKnowledgeBaseSuggestionsEvent(suggestions kb.ArticleList) *StreamEvent {
	s := xslices.Map(
		suggestions,
		func(article *kb.Article) *KnowledgeBaseSuggestion {
			return &KnowledgeBaseSuggestion{Url: article.URL, Title: article.Title}
		},
	)
	return appEvent(&AppEvent{Event: &AppEvent_KnowledgeBaseSuggestions{
		KnowledgeBaseSuggestions: &KnowledgeBaseSuggestionsEvent{
			Suggestions: s,
		},
	}})
}

func NewLoginError(err LoginErrorType, message string) *StreamEvent {
	return loginEvent(&LoginEvent{Event: &LoginEvent_Error{Error: &LoginErrorEvent{Type: err, Message: message}}})
}

func NewLoginTfaRequestedEvent(username string) *StreamEvent {
	return loginEvent(&LoginEvent{Event: &LoginEvent_TfaRequested{TfaRequested: &LoginTfaRequestedEvent{Username: username}}})
}

func NewLoginTwoPasswordsRequestedEvent(username string) *StreamEvent {
	return loginEvent(&LoginEvent{Event: &LoginEvent_TwoPasswordRequested{TwoPasswordRequested: &LoginTwoPasswordsRequestedEvent{Username: username}}})
}

func NewLoginFinishedEvent(userID string, wasSignedOut bool) *StreamEvent {
	return loginEvent(&LoginEvent{Event: &LoginEvent_Finished{Finished: &LoginFinishedEvent{UserID: userID, WasSignedOut: wasSignedOut}}})
}

func NewLoginAlreadyLoggedInEvent(userID string) *StreamEvent {
	return loginEvent(&LoginEvent{Event: &LoginEvent_AlreadyLoggedIn{AlreadyLoggedIn: &LoginFinishedEvent{UserID: userID}}})
}

func NewLoginHvRequestedEvent(hvChallengeURL string) *StreamEvent {
	return loginEvent(&LoginEvent{Event: &LoginEvent_HvRequested{HvRequested: &LoginHvRequestedEvent{HvUrl: hvChallengeURL}}})
}

func NewUpdateErrorEvent(errorType UpdateErrorType) *StreamEvent {
	return updateEvent(&UpdateEvent{Event: &UpdateEvent_Error{Error: &UpdateErrorEvent{Type: errorType}}})
}

func NewUpdateManualReadyEvent(version string) *StreamEvent {
	return updateEvent(&UpdateEvent{Event: &UpdateEvent_ManualReady{ManualReady: &UpdateManualReadyEvent{Version: version}}})
}

func NewUpdateManualRestartNeededEvent() *StreamEvent {
	return updateEvent(&UpdateEvent{Event: &UpdateEvent_ManualRestartNeeded{ManualRestartNeeded: &UpdateManualRestartNeededEvent{}}})
}

func NewUpdateForceEvent(version string) *StreamEvent {
	return updateEvent(&UpdateEvent{Event: &UpdateEvent_Force{Force: &UpdateForceEvent{Version: version}}})
}

func NewUpdateSilentRestartNeededEvent() *StreamEvent {
	return updateEvent(&UpdateEvent{Event: &UpdateEvent_SilentRestartNeeded{SilentRestartNeeded: &UpdateSilentRestartNeeded{}}})
}

func NewUpdateIsLatestVersionEvent() *StreamEvent {
	return updateEvent(&UpdateEvent{Event: &UpdateEvent_IsLatestVersion{IsLatestVersion: &UpdateIsLatestVersion{}}})
}

func NewUpdateCheckFinishedEvent() *StreamEvent {
	return updateEvent(&UpdateEvent{Event: &UpdateEvent_CheckFinished{CheckFinished: &UpdateCheckFinished{}}})
}

func NewUpdateVersionChangedEvent() *StreamEvent {
	return updateEvent(&UpdateEvent{Event: &UpdateEvent_VersionChanged{VersionChanged: &UpdateVersionChanged{}}})
}

func NewDiskCacheErrorEvent(err DiskCacheErrorType) *StreamEvent {
	return cacheEvent(&DiskCacheEvent{Event: &DiskCacheEvent_Error{Error: &DiskCacheErrorEvent{Type: err}}})
}

func NewDiskCachePathChangedEvent(path string) *StreamEvent {
	return cacheEvent(&DiskCacheEvent{Event: &DiskCacheEvent_PathChanged{PathChanged: &DiskCachePathChangedEvent{Path: path}}})
}

func NewDiskCachePathChangeFinishedEvent() *StreamEvent {
	return cacheEvent(&DiskCacheEvent{Event: &DiskCacheEvent_PathChangeFinished{PathChangeFinished: &DiskCachePathChangeFinishedEvent{}}})
}

func NewMailServerSettingsErrorEvent(err MailServerSettingsErrorType) *StreamEvent {
	return mailServerSettingsEvent(&MailServerSettingsEvent{
		Event: &MailServerSettingsEvent_Error{
			Error: &MailServerSettingsErrorEvent{Type: err},
		},
	})
}

func NewMailServerSettingsChangedEvent(settings *ImapSmtpSettings) *StreamEvent {
	return mailServerSettingsEvent(&MailServerSettingsEvent{
		Event: &MailServerSettingsEvent_MailServerSettingsChanged{
			MailServerSettingsChanged: &MailServerSettingsChangedEvent{Settings: settings},
		},
	})
}

func NewChangeMailServerSettingsFinishedEvent() *StreamEvent {
	return mailServerSettingsEvent(&MailServerSettingsEvent{
		Event: &MailServerSettingsEvent_ChangeMailServerSettingsFinished{
			ChangeMailServerSettingsFinished: &ChangeMailServerSettingsFinishedEvent{},
		},
	})
}

func NewKeychainChangeKeychainFinishedEvent() *StreamEvent {
	return keychainEvent(&KeychainEvent{Event: &KeychainEvent_ChangeKeychainFinished{ChangeKeychainFinished: &ChangeKeychainFinishedEvent{}}})
}

func NewKeychainHasNoKeychainEvent() *StreamEvent {
	return keychainEvent(&KeychainEvent{Event: &KeychainEvent_HasNoKeychain{HasNoKeychain: &HasNoKeychainEvent{}}})
}

func NewKeychainRebuildKeychainEvent() *StreamEvent {
	return keychainEvent(&KeychainEvent{Event: &KeychainEvent_RebuildKeychain{RebuildKeychain: &RebuildKeychainEvent{}}})
}

func NewMailAddressChangeEvent(email string) *StreamEvent {
	return mailEvent(&MailEvent{Event: &MailEvent_AddressChanged{AddressChanged: &AddressChangedEvent{Address: email}}})
}

func NewMailAddressChangeLogoutEvent(email string) *StreamEvent {
	return mailEvent(&MailEvent{Event: &MailEvent_AddressChangedLogout{AddressChangedLogout: &AddressChangedLogoutEvent{Address: email}}})
}

func NewMailApiCertIssue() *StreamEvent { //nolint:revive,stylecheck
	return mailEvent(&MailEvent{Event: &MailEvent_ApiCertIssue{ApiCertIssue: &ApiCertIssueEvent{}}})
}

func NewUserToggleSplitModeFinishedEvent(userID string) *StreamEvent {
	return userEvent(&UserEvent{Event: &UserEvent_ToggleSplitModeFinished{ToggleSplitModeFinished: &ToggleSplitModeFinishedEvent{UserID: userID}}})
}

func NewUserDisconnectedEvent(email string) *StreamEvent {
	return userEvent(&UserEvent{Event: &UserEvent_UserDisconnected{UserDisconnected: &UserDisconnectedEvent{Username: email}}})
}

func NewUserChangedEvent(userID string) *StreamEvent {
	return userEvent(&UserEvent{Event: &UserEvent_UserChanged{UserChanged: &UserChangedEvent{UserID: userID}}})
}

func NewUserBadEvent(userID string, errorMessage string) *StreamEvent {
	return userEvent(&UserEvent{Event: &UserEvent_UserBadEvent{UserBadEvent: &UserBadEvent{UserID: userID, ErrorMessage: errorMessage}}})
}

func NewUsedBytesChangedEvent(userID string, usedBytes uint64) *StreamEvent {
	return userEvent(&UserEvent{Event: &UserEvent_UsedBytesChangedEvent{UsedBytesChangedEvent: &UsedBytesChangedEvent{UserID: userID, UsedBytes: int64(usedBytes)}}})
}

func newIMAPLoginFailedEvent(username string) *StreamEvent {
	return userEvent(&UserEvent{Event: &UserEvent_ImapLoginFailedEvent{ImapLoginFailedEvent: &ImapLoginFailedEvent{Username: username}}})
}

func NewSyncStartedEvent(userID string) *StreamEvent {
	return userEvent(&UserEvent{Event: &UserEvent_SyncStartedEvent{SyncStartedEvent: &SyncStartedEvent{UserID: userID}}})
}

func NewSyncFinishedEvent(userID string) *StreamEvent {
	return userEvent(&UserEvent{Event: &UserEvent_SyncFinishedEvent{SyncFinishedEvent: &SyncFinishedEvent{UserID: userID}}})
}

func NewSyncProgressEvent(userID string, progress float64, elapsedMs, remainingMs int64) *StreamEvent {
	return userEvent(&UserEvent{Event: &UserEvent_SyncProgressEvent{SyncProgressEvent: &SyncProgressEvent{
		UserID:      userID,
		Progress:    progress,
		ElapsedMs:   elapsedMs,
		RemainingMs: remainingMs,
	}}})
}

func NewGenericErrorEvent(errorCode ErrorCode) *StreamEvent {
	return genericErrorEvent(&GenericErrorEvent{Code: errorCode})
}

func NewRepairStartedEvent() *StreamEvent {
	return appEvent(&AppEvent{Event: &AppEvent_RepairStarted{RepairStarted: &RepairStartedEvent{}}})
}

func NewAllUsersLoadedEvent() *StreamEvent {
	return appEvent(&AppEvent{Event: &AppEvent_AllUsersLoaded{AllUsersLoaded: &AllUsersLoadedEvent{}}})
}

// Event category factory functions.

func appEvent(appEvent *AppEvent) *StreamEvent {
	return &StreamEvent{Event: &StreamEvent_App{App: appEvent}}
}

func loginEvent(event *LoginEvent) *StreamEvent {
	return &StreamEvent{Event: &StreamEvent_Login{Login: event}}
}

func updateEvent(event *UpdateEvent) *StreamEvent {
	return &StreamEvent{Event: &StreamEvent_Update{Update: event}}
}

func cacheEvent(event *DiskCacheEvent) *StreamEvent {
	return &StreamEvent{Event: &StreamEvent_Cache{Cache: event}}
}

func mailServerSettingsEvent(event *MailServerSettingsEvent) *StreamEvent {
	return &StreamEvent{Event: &StreamEvent_MailServerSettings{MailServerSettings: event}}
}

func keychainEvent(event *KeychainEvent) *StreamEvent {
	return &StreamEvent{Event: &StreamEvent_Keychain{Keychain: event}}
}

func mailEvent(event *MailEvent) *StreamEvent {
	return &StreamEvent{Event: &StreamEvent_Mail{Mail: event}}
}

func userEvent(event *UserEvent) *StreamEvent {
	return &StreamEvent{Event: &StreamEvent_User{User: event}}
}

func genericErrorEvent(event *GenericErrorEvent) *StreamEvent {
	return &StreamEvent{Event: &StreamEvent_GenericError{GenericError: event}}
}
