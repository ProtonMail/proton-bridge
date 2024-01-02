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

package locations

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
)

// Provider provides standard locations.
type Provider interface {
	UserConfig() string
	UserData() string
	UserCache() string
}

// DefaultProvider is a locations provider using the system-default storage locations.
type DefaultProvider struct {
	config, data, cache string
}

func NewDefaultProvider(name string) (*DefaultProvider, error) {
	config, err := os.UserConfigDir()
	if err != nil {
		return nil, err
	}

	data, err := userDataDir()
	if err != nil {
		return nil, err
	}

	cache, err := os.UserCacheDir()
	if err != nil {
		return nil, err
	}

	provider := &DefaultProvider{
		config: filepath.Join(config, name),
		data:   filepath.Join(data, name),
		cache:  filepath.Join(cache, name),
	}

	if err := os.MkdirAll(provider.config, 0o700); err != nil {
		return nil, err
	}

	if err := os.MkdirAll(provider.data, 0o700); err != nil {
		return nil, err
	}

	if err := os.MkdirAll(provider.cache, 0o700); err != nil {
		return nil, err
	}

	return provider, nil
}

// UserConfig returns a directory that can be used to store user-specific configuration.
// $XDG_CONFIG_HOME/protonmail is used on Unix systems; similar on others.
func (p *DefaultProvider) UserConfig() string {
	return p.config
}

// UserData returns a directory that can be used to store user-specific data.
// $XDG_DATA_HOME/protonmail is used on Unix systems; similar on others.
func (p *DefaultProvider) UserData() string {
	return p.data
}

// UserCache returns a directory that can be used to store user-specific non-essential data.
// $XDG_CACHE_HOME/protonmail is used on Unix systems; similar on others.
func (p *DefaultProvider) UserCache() string {
	return p.cache
}

// userDataDir returns a directory that can be used to store user-specific data.
// This is necessary because os.UserDataDir() is not implemented by the Go standard library, sadly.
// On non-linux systems, it is the same as os.UserConfigDir().
func userDataDir() (string, error) {
	if runtime.GOOS != "linux" {
		return os.UserConfigDir()
	}

	if dir := os.Getenv("XDG_DATA_HOME"); dir != "" {
		return dir, nil
	}

	if dir := os.Getenv("HOME"); dir != "" {
		return filepath.Join(dir, ".local", "share"), nil
	}

	return "", errors.New("neither $XDG_DATA_HOME nor $HOME are defined")
}
