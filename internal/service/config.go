// Copyright (c) 2024 Proton AG
//
// This file is part of Proton Mail Bridge.Bridge.
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

package service

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Config is a structure containing the service configuration data that are exchanged by the gRPC server and client.
type Config struct {
	Port           int    `json:"port"`
	Cert           string `json:"cert"`
	Token          string `json:"token"`
	FileSocketPath string `json:"fileSocketPath"`
}

// save saves a gRPC service configuration to file.
func (s *Config) save(path string) error {
	// Another process may be waiting for this file to be available. In order to prevent this process to open
	// the file while we are writing in it, we write it with a temp file name, then rename it.
	tempPath := path + "_"
	if err := s._save(tempPath); err != nil {
		return err
	}

	return os.Rename(tempPath, path)
}

func (s *Config) _save(path string) error {
	f, err := os.Create(path) //nolint:errcheck,gosec
	if err != nil {
		return err
	}

	defer func() { _ = f.Close() }()

	return json.NewEncoder(f).Encode(s)
}

// Load loads a gRPC service configuration from file.
func (s *Config) Load(path string) error {
	f, err := os.Open(path) //nolint:errcheck,gosec
	if err != nil {
		return err
	}

	defer func() { _ = f.Close() }()

	return json.NewDecoder(f).Decode(s)
}

// SaveGRPCServerConfigFile save GRPC configuration file.
func SaveGRPCServerConfigFile(locations Locator, config *Config, filename string) (string, error) {
	settingsPath, err := locations.ProvideSettingsPath()
	if err != nil {
		return "", err
	}

	configPath := filepath.Join(settingsPath, filename)

	return configPath, config.save(configPath)
}
