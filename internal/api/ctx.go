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
	"net/http"

	"github.com/ProtonMail/proton-bridge/v2/pkg/listener"
)

// httpHandler with Go's Response and Request.
type httpHandler func(http.ResponseWriter, *http.Request)

// handler with our context.
type handler func(handlerContext) error

type handlerContext struct {
	req           *http.Request
	resp          http.ResponseWriter
	eventListener listener.Listener
}

func wrapper(api *apiServer, callback handler) httpHandler {
	return func(w http.ResponseWriter, req *http.Request) {
		ctx := handlerContext{
			req:           req,
			resp:          w,
			eventListener: api.eventListener,
		}
		err := callback(ctx)
		if err != nil {
			log.Error("API callback of ", req.URL, " failed: ", err)
			http.Error(w, err.Error(), 500)
		}
	}
}
