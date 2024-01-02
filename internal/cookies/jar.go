// Copyright (c) 2024 Proton AG
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
	"net/url"
	"sync"
	"time"
)

type cookiesByHost map[string][]*http.Cookie

type Persister interface {
	GetCookies() ([]byte, error)
	SetCookies([]byte) error
}

// Jar implements http.CookieJar by wrapping the standard library's cookiejar.Jar.
// The jar uses a pantry to load cookies at startup and save cookies when set.
type Jar struct {
	jar http.CookieJar

	persister Persister
	cookies   cookiesByHost
	locker    sync.RWMutex
}

func NewCookieJar(jar http.CookieJar, persister Persister) (*Jar, error) {
	cookiesByHost, err := loadCookies(persister)
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
		jar: jar,

		persister: persister,
		cookies:   cookiesByHost,
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
	j.locker.RLock()
	defer j.locker.RUnlock()

	return j.jar.Cookies(u)
}

// PersistCookies persists the cookies to disk.
func (j *Jar) PersistCookies() error {
	j.locker.RLock()
	defer j.locker.RUnlock()

	rawCookies, err := json.Marshal(j.cookies)
	if err != nil {
		return err
	}

	return j.persister.SetCookies(rawCookies)
}

// loadCookies loads all non-expired cookies from disk.
func loadCookies(persister Persister) (cookiesByHost, error) {
	rawCookies, err := persister.GetCookies()
	if err != nil {
		return nil, err
	}

	if len(rawCookies) == 0 {
		return make(cookiesByHost), nil
	}

	var cookiesByHost cookiesByHost

	if err := json.Unmarshal(rawCookies, &cookiesByHost); err != nil {
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
