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
// along with Proton Mail Bridge.  If not, see <https://www.gnu.org/licenses/>.

package cli

import (
	"context"

	"github.com/abiosoft/ishell"
)

func (f *frontendCLI) debugMailboxState(c *ishell.Context) {
	f.ShowPrompt(false)
	defer f.ShowPrompt(true)

	checkFlags := f.yesNoQuestion("Also check message flags")

	c.Println("Starting state check. Note that depending on your message count this may take a while.")

	if err := f.bridge.CheckClientState(context.Background(), checkFlags, func(s string) {
		c.Println(s)
	}); err != nil {
		c.Printf("State check failed : %v", err)
		return
	}

	c.Println("State check finished, see log for more details.")
}
