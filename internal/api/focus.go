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

package api

import (
	"fmt"
	"net/http"

	"github.com/ProtonMail/proton-bridge/v2/internal/bridge"
	"github.com/ProtonMail/proton-bridge/v2/internal/events"
)

// focusHandler should be called from other instances (attempt to start bridge
// for the second time) to get focus in the currently running instance.
func focusHandler(ctx handlerContext) error {
	log.Info("Focus from other instance")
	ctx.eventListener.Emit(events.SecondInstanceEvent, "")
	fmt.Fprintf(ctx.resp, "OK")
	return nil
}

// CheckOtherInstanceAndFocus is helper for new instances to check if there is
// already a running instance and get it's focus.
func CheckOtherInstanceAndFocus(port int) error {
	addr := getAPIAddress(bridge.Host, port)
	resp, err := (&http.Client{}).Get("http://" + addr + "/focus")
	if err != nil {
		return err
	}
	defer resp.Body.Close() //nolint:errcheck

	if resp.StatusCode != 200 {
		log.Error("Focus error: ", resp.StatusCode)
	}
	return nil
}
