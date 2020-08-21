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

package cookies

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/ProtonMail/proton-bridge/internal/preferences"
)

// pantry persists and loads cookies to some persistent storage location.
type pantry struct {
	gs GetterSetter
}

func (p *pantry) persistCookies(host string, cookies []*http.Cookie) error {
	for _, cookie := range cookies {
		if cookie.MaxAge > 0 {
			cookie.Expires = time.Now().Add(time.Duration(cookie.MaxAge) * time.Second)
		}
	}

	cookiesByHost, err := p.loadFromJSON()
	if err != nil {
		return err
	}

	cookiesByHost[host] = cookies

	return p.saveToJSON(cookiesByHost)
}

func (p *pantry) discardExpiredCookies() error {
	cookiesByHost, err := p.loadFromJSON()
	if err != nil {
		return err
	}

	for host, cookies := range cookiesByHost {
		cookiesByHost[host] = discardExpiredCookies(cookies)
	}

	return p.saveToJSON(cookiesByHost)
}

type cookiesByHost map[string][]*http.Cookie

func (p *pantry) loadFromJSON() (cookiesByHost, error) {
	b := p.gs.Get(preferences.CookiesKey)

	if b == "" {
		return make(cookiesByHost), nil
	}

	var cookies cookiesByHost

	if err := json.Unmarshal([]byte(b), &cookies); err != nil {
		return nil, err
	}

	return cookies, nil
}

func (p *pantry) saveToJSON(cookies cookiesByHost) error {
	b, err := json.Marshal(cookies)
	if err != nil {
		return err
	}

	p.gs.Set(preferences.CookiesKey, string(b))

	return nil
}

func discardExpiredCookies(cookies []*http.Cookie) (validCookies []*http.Cookie) {
	for _, cookie := range cookies {
		if cookie.Expires.After(time.Now()) {
			validCookies = append(validCookies, cookie)
		}
	}

	return
}
