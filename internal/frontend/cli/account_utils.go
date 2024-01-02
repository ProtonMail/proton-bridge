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

package cli

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/ProtonMail/proton-bridge/v3/internal/bridge"
	"github.com/abiosoft/ishell"
)

// completeUsernames is a helper to complete usernames as the user types.
func (f *frontendCLI) completeUsernames(args []string) (usernames []string) {
	if len(args) > 1 {
		return
	}
	arg := ""
	if len(args) == 1 {
		arg = args[0]
	}

	for _, userID := range f.bridge.GetUserIDs() {
		user, err := f.bridge.GetUserInfo(userID)
		if err != nil {
			panic(err)
		}
		if strings.HasPrefix(strings.ToLower(user.Username), strings.ToLower(arg)) {
			usernames = append(usernames, user.Username)
		}
	}
	return
}

// noAccountWrapper is a decorator for functions which need any account to be properly functional.
func (f *frontendCLI) noAccountWrapper(callback func(*ishell.Context)) func(*ishell.Context) {
	return func(c *ishell.Context) {
		if len(f.bridge.GetUserIDs()) == 0 {
			f.Println("No active accounts. Please add account to continue.")
		} else {
			callback(c)
		}
	}
}

func (f *frontendCLI) askUserByIndexOrName(c *ishell.Context) bridge.UserInfo {
	user := f.getUserByIndexOrName("")
	if user.UserID != "" {
		return user
	}

	numberOfAccounts := len(f.bridge.GetUserIDs())
	indexRange := fmt.Sprintf("number between 0 and %d", numberOfAccounts-1)
	if len(c.Args) == 0 {
		f.Printf("Please choose %s or username.\n", indexRange)
		return bridge.UserInfo{}
	}
	arg := c.Args[0]
	user = f.getUserByIndexOrName(arg)
	if user.UserID == "" {
		f.Printf("Wrong input '%s'. Choose %s or username.\n", bold(arg), indexRange)
		return bridge.UserInfo{}
	}
	return user
}

func (f *frontendCLI) getUserByIndexOrName(arg string) bridge.UserInfo {
	userIDs := f.bridge.GetUserIDs()
	numberOfAccounts := len(userIDs)
	if numberOfAccounts == 0 {
		return bridge.UserInfo{}
	}
	res := make([]bridge.UserInfo, len(userIDs))
	for idx, userID := range userIDs {
		user, err := f.bridge.GetUserInfo(userID)
		if err != nil {
			panic(err)
		}
		res[idx] = user
	}
	if numberOfAccounts == 1 {
		return res[0]
	}
	if index, err := strconv.Atoi(arg); err == nil {
		if index < 0 || index >= numberOfAccounts {
			return bridge.UserInfo{}
		}
		return res[index]
	}
	for _, user := range res {
		if user.Username == arg {
			return user
		}
	}
	return bridge.UserInfo{}
}
