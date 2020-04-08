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

package connection

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/ProtonMail/proton-bridge/pkg/logs"
	"github.com/ProtonMail/proton-bridge/pkg/pmapi"
)

// Errors for possible connection issues
var (
	ErrNoInternetConnection = errors.New("no internet connection")
	ErrCanNotReachAPI       = errors.New("can not reach PM API")
	log                     = logs.GetLogEntry("connection") //nolint[gochecknoglobals]
)

// CheckInternetConnection does a check of API connection. It checks two of our endpoints in parallel.
// One endpoint is part of the protonmail API, while the other is not.
// This allows us to determine whether there is a problem with the connection itself or only a problem with our API.
// Two errors can be returned, ErrNoInternetConnection or ErrCanNotReachAPI.
func CheckInternetConnection() error {
	client := &http.Client{
		// TODO: Set transport properly! (Need access to ClientManager somehow)
		// Transport: pmapi.NewDialerWithPinning(pmapi.CurrentUserAgent).TransportWithPinning(),
	}

	// Do not cumulate timeouts, use goroutines.
	retStatus := make(chan error)
	retAPI := make(chan error)

	// Check protonstatus.com without SSL for performance reasons. vpn_status endpoint is fast and
	// returns only OK; this endpoint is not known by the public. We check the connection only.
	go checkConnection(client, "http://protonstatus.com/vpn_status", retStatus)

	// Check of API reachability also uses a fast endpoint.
	// TODO: This should check active proxy, not the RootURL
	go checkConnection(client, pmapi.RootURL+"/tests/ping", retAPI)

	errStatus := <-retStatus
	errAPI := <-retAPI

	if errStatus != nil {
		if errAPI != nil {
			log.Error("Checking internet connection failed with ", errStatus, " and ", errAPI)
			return ErrNoInternetConnection
		}
		log.Warning("API OK, but status: ", errStatus)
		return nil
	}

	if errAPI != nil {
		log.Error("Status OK, but API: ", errAPI)
		return ErrCanNotReachAPI
	}

	return nil
}

func checkConnection(client *http.Client, url string, errorChannel chan error) {
	resp, err := client.Get(url)
	if err != nil {
		errorChannel <- err
		return
	}
	_ = resp.Body.Close()
	if resp.StatusCode != 200 {
		errorChannel <- fmt.Errorf("HTTP status code %d", resp.StatusCode)
		return
	}
	errorChannel <- nil
}
