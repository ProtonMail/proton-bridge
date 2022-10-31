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

// Package cookies implements a persistent cookie jar which satisfies the http.CookieJar interface.
package cookies

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"sync"
	"time"

	"github.com/ProtonMail/proton-bridge/v2/internal/config/settings"
)

type cookiesByHost map[string][]*http.Cookie

// Jar implements http.CookieJar by wrapping the standard library's cookiejar.Jar.
// The jar uses a pantry to load cookies at startup and save cookies when set.
type Jar struct {
	jar      *cookiejar.Jar
	settings *settings.Settings
	cookies  cookiesByHost
	locker   sync.Locker
}

func NewCookieJar(s *settings.Settings) (*Jar, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}

	cookiesByHost, err := loadCookies(s)
	if err != nil {
		return nil, err
	}

	for host, cookies := range cookiesByHost {
		url, err := url.Parse(host)
		if err != nil {
			continue
		}

		jar.SetCookies(url, cookies)
	}

	return &Jar{
		jar:      jar,
		settings: s,
		cookies:  cookiesByHost,
		locker:   &sync.Mutex{},
	}, nil
}

func (j *Jar) SetCookies(u *url.URL, cookies []*http.Cookie) {
	j.locker.Lock()
	defer j.locker.Unlock()

	j.jar.SetCookies(u, cookies)

	for _, cookie := range cookies {
		if cookie.MaxAge > 0 {
			cookie.Expires = time.Now().Add(time.Duration(cookie.MaxAge) * time.Second)
		}
	}

	j.cookies[fmt.Sprintf("%v://%v", u.Scheme, u.Host)] = cookies
}

func (j *Jar) Cookies(u *url.URL) []*http.Cookie {
	j.locker.Lock()
	defer j.locker.Unlock()

	return j.jar.Cookies(u)
}

// PersistCookies persists the cookies to disk.
func (j *Jar) PersistCookies() error {
	j.locker.Lock()
	defer j.locker.Unlock()

	rawCookies, err := json.Marshal(j.cookies)
	if err != nil {
		return err
	}

	j.settings.Set(settings.CookiesKey, string(rawCookies))

	return nil
}

// loadCookies loads all non-expired cookies from disk.
func loadCookies(s *settings.Settings) (cookiesByHost, error) {
	rawCookies := s.Get(settings.CookiesKey)

	if rawCookies == "" {
		return make(cookiesByHost), nil
	}

	var cookiesByHost cookiesByHost

	if err := json.Unmarshal([]byte(rawCookies), &cookiesByHost); err != nil {
		return nil, err
	}

	for host, cookies := range cookiesByHost {
		if validCookies := discardExpiredCookies(cookies); len(validCookies) > 0 {
			cookiesByHost[host] = validCookies
		}
	}

	return cookiesByHost, nil
}

// discardExpiredCookies returns all the given cookies which aren't expired.
func discardExpiredCookies(cookies []*http.Cookie) []*http.Cookie {
	var validCookies []*http.Cookie

	for _, cookie := range cookies {
		if cookie.Expires.After(time.Now()) {
			validCookies = append(validCookies, cookie)
		}
	}

	return validCookies
}
