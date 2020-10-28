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

package pmapi

import (
	"errors"
	"fmt"
	"net/http"
	"time"
)

const protonStatusURL = "http://protonstatus.com/vpn_status"

// ErrNoInternetConnection indicates that both protonstatus and the API are unreachable.
var ErrNoInternetConnection = errors.New("no internet connection")

// CheckConnection returns an error if there is no internet connection.
// This should be moved to the ConnectionManager when it is implemented.
func (cm *ClientManager) CheckConnection() error {
	// We use a normal dialer here which doesn't check tls fingerprints.
	client := &http.Client{Timeout: time.Second * 10}

	// Do not cumulate timeouts, use goroutines.
	retStatus := make(chan error)
	retAPI := make(chan error)

	// vpn_status endpoint is fast and returns only OK. We check the connection only.
	go checkConnection(client, protonStatusURL, retStatus)

	// Check of API reachability also uses a fast endpoint.
	go checkConnection(client, cm.GetRootURL()+"/tests/ping", retAPI)

	errStatus := <-retStatus
	errAPI := <-retAPI

	switch {
	case errStatus == nil && errAPI == nil:
		return nil

	case errStatus == nil && errAPI != nil:
		cm.log.Error("ProtonStatus is reachable but API is not")
		return ErrAPINotReachable

	case errStatus != nil && errAPI == nil:
		cm.log.Warn("API is reachable but protonstatus is not")
		return nil

	case errStatus != nil && errAPI != nil:
		cm.log.Error("Both ProtonStatus and API are unreachable")
		return ErrNoInternetConnection
	}

	return nil
}

// CheckConnection returns an error if there is no internet connection.
func CheckConnection() error {
	client := &http.Client{Timeout: time.Second * 10}
	retStatus := make(chan error)
	go checkConnection(client, protonStatusURL, retStatus)
	return <-retStatus
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
