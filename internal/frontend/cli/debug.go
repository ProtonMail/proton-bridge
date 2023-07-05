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
	"os"

	"github.com/abiosoft/ishell"
)

func (f *frontendCLI) debugMailboxState(c *ishell.Context) {
	f.ShowPrompt(false)
	defer f.ShowPrompt(true)

	checkFlags := f.yesNoQuestion("Also check message flags")

	c.Println("Starting state check. Note that depending on your message count this may take a while.")

	result, err := f.bridge.CheckClientState(context.Background(), checkFlags, func(s string) {
		c.Println(s)
	})
	if err != nil {
		c.Printf("State check failed : %v", err)
		return
	}

	c.Println("State check finished, see log for more details.")

	if len(result.MissingMessages) == 0 {
		return
	}

	f.Println("\n\nSome missing messages were detected. Bridge can download these messages for you")
	f.Println("in a directory which you can later send to the developers for analysis.\n")
	f.Println(bold("Note that the Messages will be stored unencrypted on disk.") + " If you do not wish")
	f.Println("to continue, input no in the prompt below.\n")

	if !f.yesNoQuestion("Would you like to proceed") {
		return
	}

	location, err := os.MkdirTemp("", "debug-state-check-*")
	if err != nil {
		f.Printf("Failed to create temporary directory: %v\n", err)
		return
	}

	c.Printf("Messages will be downloaded to: %v\n\n", bold(location))

	if err := f.bridge.DebugDownloadFailedMessages(context.Background(), result, location, func(s string, i int, i2 int) {
		f.Printf("[%v] Retrieving message %v of %v\n", s, i, i2)
	}); err != nil {
		f.Println(err)
		return
	}

	c.Printf("\nMessage download finished. Data is available at %v\n", bold(location))
}
