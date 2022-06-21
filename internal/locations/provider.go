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

package locations

import (
	"os"
	"path/filepath"
)

// Provider provides standard locations.
type Provider interface {
	UserConfig() string
	UserCache() string
}

// DefaultProvider is a locations provider using the system-default storage locations.
type DefaultProvider struct {
	config, cache string
}

func NewDefaultProvider(name string) (*DefaultProvider, error) {
	config, err := os.UserConfigDir()
	if err != nil {
		return nil, err
	}

	cache, err := os.UserCacheDir()
	if err != nil {
		return nil, err
	}

	return &DefaultProvider{
		config: filepath.Join(config, name),
		cache:  filepath.Join(cache, name),
	}, nil
}

func (p *DefaultProvider) UserConfig() string {
	return p.config
}

func (p *DefaultProvider) UserCache() string {
	return p.cache
}
