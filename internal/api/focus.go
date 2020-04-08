// Copyright (c) 2020 Proton Technologies AG
//
// This file is part of ProtonMail Bridge.
//
// ProtonMail Bridge is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// ProtonMail Bridge is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with ProtonMail Bridge.  If not, see <https://www.gnu.org/licenses/>.

package api

import (
	"crypto/tls"
	"fmt"
	"net/http"

	"github.com/ProtonMail/proton-bridge/internal/bridge"
	"github.com/ProtonMail/proton-bridge/internal/events"
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
func CheckOtherInstanceAndFocus(port int, tls *tls.Config) error {
	transport := &http.Transport{TLSClientConfig: tls}
	client := &http.Client{Transport: transport}

	addr := getAPIAddress(bridge.Host, port)
	resp, err := client.Get("https://" + addr + "/focus")
	if err != nil {
		return err
	}
	defer resp.Body.Close() //nolint[errcheck]

	if resp.StatusCode != 200 {
		log.Error("Focus error: ", resp.StatusCode)
	}
	return nil
}
