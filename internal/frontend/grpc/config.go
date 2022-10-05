// Copyright (c) 2022 Proton AG
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

package grpc

import (
	"encoding/json"
	"os"
)

// config is a structure containing the service configuration data that are exchanged by the gRPC server and client.
type config struct {
	Port  int    `json:"port"`
	Cert  string `json:"cert"`
	Token string `json:"token"`
}

// save saves a gRPC service configuration to file.
func (s *config) save(path string) error {
	// Another process may be waiting for this file to be available. In order to prevent this process to open
	// the file while we are writing in it, we write it with a temp file name, then rename it.
	tempPath := path + "_"
	if err := s._save(tempPath); err != nil {
		return err
	}

	return os.Rename(tempPath, path)
}

func (s *config) _save(path string) error {
	f, err := os.Create(path) //nolint:errcheck,gosec
	if err != nil {
		return err
	}

	defer func() { _ = f.Close() }()

	return json.NewEncoder(f).Encode(s)
}

// load loads a gRPC service configuration from file.
func (s *config) load(path string) error {
	f, err := os.Open(path) //nolint:errcheck,gosec
	if err != nil {
		return err
	}

	defer func() { _ = f.Close() }()

	return json.NewDecoder(f).Decode(s)
}
