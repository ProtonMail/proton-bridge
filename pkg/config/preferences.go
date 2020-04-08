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

package config

import (
	"encoding/json"
	"errors"
	"os"
	"strconv"
	"sync"
)

type Preferences struct {
	cache map[string]string
	path  string
	lock  *sync.RWMutex
}

// NewPreferences returns loaded preferences.
func NewPreferences(preferencesPath string) *Preferences {
	p := &Preferences{
		path: preferencesPath,
		lock: &sync.RWMutex{},
	}
	if err := p.load(); err != nil {
		log.Warn("Cannot load preferences: ", err)
	}
	return p
}

func (p *Preferences) load() error {
	if p.cache != nil {
		return nil
	}

	p.lock.Lock()
	defer p.lock.Unlock()

	p.cache = map[string]string{}

	f, err := os.Open(p.path)
	if err != nil {
		return err
	}
	defer f.Close() //nolint[errcheck]

	return json.NewDecoder(f).Decode(&p.cache)
}

func (p *Preferences) save() error {
	if p.cache == nil {
		return errors.New("cannot save preferences: cache is nil")
	}

	p.lock.Lock()
	defer p.lock.Unlock()

	f, err := os.Create(p.path)
	if err != nil {
		return err
	}
	defer f.Close() //nolint[errcheck]

	return json.NewEncoder(f).Encode(p.cache)
}

func (p *Preferences) SetDefault(key, value string) {
	if p.Get(key) == "" {
		p.Set(key, value)
	}
}

func (p *Preferences) Get(key string) string {
	p.lock.RLock()
	defer p.lock.RUnlock()

	return p.cache[key]
}

func (p *Preferences) GetBool(key string) bool {
	return p.Get(key) == "true"
}

func (p *Preferences) GetInt(key string) int {
	value, err := strconv.Atoi(p.Get(key))
	if err != nil {
		log.Error("Cannot parse int: ", err)
	}
	return value
}

func (p *Preferences) Set(key, value string) {
	p.lock.Lock()
	p.cache[key] = value
	p.lock.Unlock()

	if err := p.save(); err != nil {
		log.Warn("Cannot save preferences: ", err)
	}
}

func (p *Preferences) SetBool(key string, value bool) {
	if value {
		p.Set(key, "true")
	} else {
		p.Set(key, "false")
	}
}

func (p *Preferences) SetInt(key string, value int) {
	p.Set(key, strconv.Itoa(value))
}
