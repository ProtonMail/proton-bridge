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

package grpc

import (
	"regexp"
	"strings"

	"github.com/ProtonMail/proton-bridge/v2/internal/frontend/types"
)

var (
	reMultiSpaces     = regexp.MustCompile(`\s{2,}`)
	reStartWithSymbol = regexp.MustCompile(`^[.,/#!$@%^&*;:{}=\-_` + "`" + `~()]`)
)

// getInitials based on webapp implementation:
// https://github.com/ProtonMail/WebClients/blob/55d96a8b4afaaa4372fc5f1ef34953f2070fd7ec/packages/shared/lib/helpers/string.ts#L145
func getInitials(fullName string) string {
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

// grpcUserFromBridge converts a bridge user to a gRPC user.
func grpcUserFromBridge(user types.User) *User {
	return &User{
		Id:             user.ID(),
		Username:       user.Username(),
		AvatarText:     getInitials(user.Username()),
		LoggedIn:       user.IsConnected(),
		SplitMode:      user.IsCombinedAddressMode(),
		SetupGuideSeen: true, // users listed have already seen the setup guide.
		UsedBytes:      user.UsedBytes(),
		TotalBytes:     user.TotalBytes(),
		Password:       user.GetBridgePassword(),
		Addresses:      user.GetAddresses(),
	}
}
