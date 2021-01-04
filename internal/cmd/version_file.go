// Copyright (c) 2021 Proton Technologies AG
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

package cmd

import "github.com/ProtonMail/proton-bridge/internal/updates"

// GenerateVersionFiles writes a JSON file with details about current build.
// Those files are used for upgrading the app.
func GenerateVersionFiles(updates *updates.Updates, dir string) {
	log.Info("Generating version files")
	for _, goos := range []string{"windows", "darwin", "linux"} {
		log.Debug("Generating JSON for ", goos)
		if err := updates.CreateJSONAndSign(dir, goos); err != nil {
			log.Error(err)
		}
	}
}
