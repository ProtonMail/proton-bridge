// Copyright (c) 2020 Proton Technologies AG
//
// This file is part of ProtonMail Bridge.Bridge.
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

package context

import (
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"

	"github.com/ProtonMail/proton-bridge/pkg/pmapi"
)

type fakeConfig struct {
	dir string
}

// newFakeConfig creates a temporary folder for files.
// It's expected the test calls `ClearData` before finish to remove it from the file system.
func newFakeConfig() *fakeConfig {
	dir, err := ioutil.TempDir("", "example")
	if err != nil {
		panic(err)
	}

	return &fakeConfig{
		dir: dir,
	}
}

func (c *fakeConfig) ClearData() error {
	return os.RemoveAll(c.dir)
}
func (c *fakeConfig) GetAPIConfig() *pmapi.ClientConfig {
	return &pmapi.ClientConfig{
		AppVersion: "Bridge_" + os.Getenv("VERSION"),
		ClientID:   "bridge",
	}
}
func (c *fakeConfig) GetDBDir() string {
	return c.dir
}
func (c *fakeConfig) GetLogDir() string {
	return c.dir
}
func (c *fakeConfig) GetLogPrefix() string {
	return "test"
}
func (c *fakeConfig) GetPreferencesPath() string {
	return filepath.Join(c.dir, "prefs.json")
}
func (c *fakeConfig) GetTLSCertPath() string {
	return filepath.Join(c.dir, "cert.pem")
}
func (c *fakeConfig) GetTLSKeyPath() string {
	return filepath.Join(c.dir, "key.pem")
}
func (c *fakeConfig) GetEventsPath() string {
	return filepath.Join(c.dir, "events.json")
}
func (c *fakeConfig) GetIMAPCachePath() string {
	return filepath.Join(c.dir, "user_info.json")
}
func (c *fakeConfig) GetDefaultAPIPort() int {
	return 21042
}
func (c *fakeConfig) GetDefaultIMAPPort() int {
	return 21100 + rand.Intn(100)
}
func (c *fakeConfig) GetDefaultSMTPPort() int {
	return 21200 + rand.Intn(100)
}
