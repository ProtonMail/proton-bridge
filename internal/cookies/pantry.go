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

	"github.com/ProtonMail/proton-bridge/internal/preferences"
)

// pantry persists and loads cookies to some persistent storage location.
type pantry struct {
	gs GetterSetter
}

func (p *pantry) persistCookies(url string, cookies []*http.Cookie) error {
	b, err := json.Marshal(cookies)
	if err != nil {
		return err
	}

	val, err := p.loadFromJSON()
	if err != nil {
		return err
	}

	val[url] = string(b)

	return p.saveToJSON(val)
}

func (p *pantry) loadCookies() (map[string][]*http.Cookie, error) {
	res := make(map[string][]*http.Cookie)

	val, err := p.loadFromJSON()
	if err != nil {
		return nil, err
	}

	for url, rawCookies := range val {
		var cookies []*http.Cookie

		if err := json.Unmarshal([]byte(rawCookies), &cookies); err != nil {
			return nil, err
		}

		res[url] = cookies
	}

	return res, nil
}

type dataStructure map[string]string

func (p *pantry) loadFromJSON() (dataStructure, error) {
	b := p.gs.Get(preferences.CookiesKey)

	if b == "" {
		return make(dataStructure), nil
	}

	var val dataStructure

	if err := json.Unmarshal([]byte(b), &val); err != nil {
		return nil, err
	}

	return val, nil
}

func (p *pantry) saveToJSON(val dataStructure) error {
	b, err := json.Marshal(val)
	if err != nil {
		return err
	}

	p.gs.Set(preferences.CookiesKey, string(b))

	return nil
}
