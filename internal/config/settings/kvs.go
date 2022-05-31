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

package settings

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"sync"

	"github.com/sirupsen/logrus"
)

type keyValueStore struct {
	cache map[string]string
	path  string
	lock  *sync.RWMutex
}

// newKeyValueStore returns loaded preferences.
func newKeyValueStore(path string) *keyValueStore {
	p := &keyValueStore{
		path: path,
		lock: &sync.RWMutex{},
	}
	if err := p.load(); err != nil {
		logrus.WithError(err).Warn("Cannot load preferences file, creating new one")
	}
	return p
}

func (p *keyValueStore) load() error {
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
	defer f.Close() //nolint:errcheck,gosec

	return json.NewDecoder(f).Decode(&p.cache)
}

func (p *keyValueStore) save() error {
	if p.cache == nil {
		return errors.New("cannot save preferences: cache is nil")
	}

	p.lock.Lock()
	defer p.lock.Unlock()

	b, err := json.MarshalIndent(p.cache, "", "\t")
	if err != nil {
		return err
	}

	return ioutil.WriteFile(p.path, b, 0o600)
}

func (p *keyValueStore) setDefault(key, value string) {
	if p.Get(key) == "" {
		p.Set(key, value)
	}
}

func (p *keyValueStore) Get(key string) string {
	p.lock.RLock()
	defer p.lock.RUnlock()

	return p.cache[key]
}

func (p *keyValueStore) GetBool(key string) bool {
	return p.Get(key) == "true"
}

func (p *keyValueStore) GetInt(key string) int {
	if p.Get(key) == "" {
		return 0
	}

	value, err := strconv.Atoi(p.Get(key))
	if err != nil {
		logrus.WithError(err).Error("Cannot parse int")
	}

	return value
}

func (p *keyValueStore) GetFloat64(key string) float64 {
	if p.Get(key) == "" {
		return 0
	}

	value, err := strconv.ParseFloat(p.Get(key), 64)
	if err != nil {
		logrus.WithError(err).Error("Cannot parse float64")
	}

	return value
}

func (p *keyValueStore) Set(key, value string) {
	p.lock.Lock()
	p.cache[key] = value
	p.lock.Unlock()

	if err := p.save(); err != nil {
		logrus.WithError(err).Warn("Cannot save preferences")
	}
}

func (p *keyValueStore) SetBool(key string, value bool) {
	if value {
		p.Set(key, "true")
	} else {
		p.Set(key, "false")
	}
}

func (p *keyValueStore) SetInt(key string, value int) {
	p.Set(key, strconv.Itoa(value))
}

func (p *keyValueStore) SetFloat64(key string, value float64) {
	p.Set(key, fmt.Sprintf("%v", value))
}
