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

// Package api provides HTTP API of the Bridge.
//
// API endpoints:
//  * /focus, see focusHandler
package api

import (
	"crypto/tls"
	"fmt"
	"net/http"

	"github.com/ProtonMail/proton-bridge/internal/bridge"
	"github.com/ProtonMail/proton-bridge/internal/events"
	"github.com/ProtonMail/proton-bridge/internal/preferences"
	"github.com/ProtonMail/proton-bridge/pkg/config"
	"github.com/ProtonMail/proton-bridge/pkg/listener"
	"github.com/ProtonMail/proton-bridge/pkg/logs"
	"github.com/ProtonMail/proton-bridge/pkg/ports"
)

var (
	log = logs.GetLogEntry("api") //nolint[gochecknoglobals]
)

type apiServer struct {
	host          string
	pref          *config.Preferences
	tls           *tls.Config
	certPath      string
	keyPath       string
	eventListener listener.Listener
}

// NewAPIServer returns prepared API server struct.
func NewAPIServer(pref *config.Preferences, tls *tls.Config, certPath, keyPath string, eventListener listener.Listener) *apiServer { //nolint[golint]
	return &apiServer{
		host:          bridge.Host,
		pref:          pref,
		tls:           tls,
		certPath:      certPath,
		keyPath:       keyPath,
		eventListener: eventListener,
	}
}

// Starts the server.
func (api *apiServer) ListenAndServe() {
	mux := http.NewServeMux()
	mux.HandleFunc("/focus", wrapper(api, focusHandler))

	addr := api.getAddress()
	server := &http.Server{
		Addr:         addr,
		Handler:      mux,
		TLSConfig:    api.tls,
		TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler)),
	}

	log.Info("API listening at ", addr)
	if err := server.ListenAndServeTLS(api.certPath, api.keyPath); err != nil {
		api.eventListener.Emit(events.ErrorEvent, "API failed: "+err.Error())
		log.Error("API failed: ", err)
	}
	defer server.Close() //nolint[errcheck]
}

func (api *apiServer) getAddress() string {
	port := api.pref.GetInt(preferences.APIPortKey)
	newPort := ports.FindFreePortFrom(port)
	if newPort != port {
		api.pref.SetInt(preferences.APIPortKey, newPort)
	}
	return getAPIAddress(api.host, newPort)
}

func getAPIAddress(host string, port int) string {
	return fmt.Sprintf("%s:%d", host, port)
}
