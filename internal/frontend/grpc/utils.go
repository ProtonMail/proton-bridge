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

package grpc

import (
	"regexp"
	"strings"

	"github.com/ProtonMail/proton-bridge/v3/internal/bridge"
	"github.com/ProtonMail/proton-bridge/v3/internal/vault"
	"github.com/sirupsen/logrus"
)

var (
	reMultiSpaces     = regexp.MustCompile(`\s{2,}`)
	reStartWithSymbol = regexp.MustCompile(`^[.,/#!$@%^&*;:{}=\-_` + "`" + `~()]`)
)

// getInitials based on webapp implementation:
// https://github.com/ProtonMail/WebClients/blob/55d96a8b4afaaa4372fc5f1ef34953f2070fd7ec/packages/shared/lib/helpers/string.ts#L145
func getInitials(fullName string) string {
	fullName = strings.TrimSpace(fullName)
	if fullName == "" {
		return "?"
	}

	words := strings.Split(
		reMultiSpaces.ReplaceAllString(fullName, " "),
		" ",
	)

	n := 0
	for _, word := range words {
		if !reStartWithSymbol.MatchString(word) {
			words[n] = word
			n++
		}
	}

	if n == 0 {
		return "?"
	}

	initials := words[0][0:1]
	if n != 1 {
		initials += words[n-1][0:1]
	}
	return strings.ToUpper(initials)
}

// grpcUserFromInfo converts a bridge user to a gRPC user.
func grpcUserFromInfo(user bridge.UserInfo) *User {
	return &User{
		Id:         user.UserID,
		Username:   user.Username,
		AvatarText: getInitials(user.Username),
		State:      userStateToGrpc(user.State),
		SplitMode:  user.AddressMode == vault.SplitMode,
		UsedBytes:  int64(user.UsedSpace),
		TotalBytes: int64(user.MaxSpace),
		Password:   user.BridgePass,
		Addresses:  user.Addresses,
	}
}

func userStateToGrpc(state bridge.UserState) UserState {
	switch state {
	case bridge.SignedOut:
		return UserState_SIGNED_OUT
	case bridge.Locked:
		return UserState_LOCKED
	case bridge.Connected:
		return UserState_CONNECTED
	default:
		panic("Unknown user state")
	}
}

// logrusLevelFromGrpcLevel converts a gRPC log level to a logrus log level.
func logrusLevelFromGrpcLevel(level LogLevel) logrus.Level {
	switch level {
	case LogLevel_LOG_PANIC:
		return logrus.PanicLevel
	case LogLevel_LOG_FATAL:
		return logrus.FatalLevel
	case LogLevel_LOG_ERROR:
		return logrus.ErrorLevel
	case LogLevel_LOG_WARN:
		return logrus.WarnLevel
	case LogLevel_LOG_INFO:
		return logrus.InfoLevel
	case LogLevel_LOG_DEBUG:
		return logrus.DebugLevel
	case LogLevel_LOG_TRACE:
		return logrus.TraceLevel
	default:
		return logrus.ErrorLevel
	}
}
