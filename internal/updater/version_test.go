// Copyright (c) 2025 Proton AG
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

package updater

import (
	"encoding/json"
	"testing"

	"github.com/Masterminds/semver/v3"
	"github.com/ProtonMail/proton-bridge/v3/internal/updater/versioncompare"
)

var mockJSONData = `
{
  "Releases": [
    {
      "CategoryName": "Stable",
      "Version": "2.1.0",
      "ReleaseDate": "2025-01-15T08:00:00Z",
      "File": [
        {
          "Url": "https://downloads.example.com/v2.1.0/MyApp-2.1.0.pkg",
          "Sha512CheckSum": "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
          "Identifier": "package"
        },
        {
          "Url": "https://downloads.example.com/v2.1.0/MyApp-2.1.0.dmg",
          "Sha512CheckSum": "cf83e1357eefb8bdf1542850d66d8007d620e4050b5715dc83f4a921d36ce9ce",
          "Identifier": "installer"
        }
      ],
      "RolloutProportion": 0.5,
      "MinAuto": "2.0.0",
      "Commit": "8f52d45c9f8c31aa391315ea24e40c4a7e0b2c1d",
      "ReleaseNotesPage": "https://example.com/releases/2.1.0/notes",
      "LandingPage": "https://example.com/releases/2.1.0"
    },
    {
      "CategoryName": "EarlyAccess",
      "Version": "2.2.0-beta.1",
      "ReleaseDate": "2025-01-20T10:00:00Z",
      "File": [
        {
          "Url": "https://downloads.example.com/beta/v2.2.0-beta.1/MyApp-2.2.0-beta.1.pkg",
          "Sha512CheckSum": "a9f0e44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
          "Identifier": "package"
        }
      ],
      "SystemVersion": {
        "Minimum": "13"
      },
      "RolloutProportion": 0.25,
      "MinAuto": "2.1.0",
      "Commit": "3e72d45c9f8c31aa391315ea24e40c4a7e0b2c1d",
      "ReleaseNotesPage": "https://example.com/releases/2.2.0-beta.1/notes",
      "LandingPage": "https://example.com/releases/2.2.0-beta.1"
    },
    {
      "CategoryName": "Stable",
      "Version": "2.0.0",
      "ReleaseDate": "2024-12-01T09:00:00Z",
      "File": [
        {
          "Url": "https://downloads.example.com/v2.0.0/MyApp-2.0.0.pkg",
          "Sha512CheckSum": "b5f0e44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
          "Identifier": "package"
        },
        {
          "Url": "https://downloads.example.com/v2.0.0/MyApp-2.0.0.dmg",
          "Sha512CheckSum": "d583e1357eefb8bdf1542850d66d8007d620e4050b5715dc83f4a921d36ce9ce",
          "Identifier": "installer"
        }
      ],
      "SystemVersion": {
			"Maximum": "12.0.0",
			"Minimum": "1.0.0"
		},
      "RolloutProportion": 1.0,
      "MinAuto": "1.9.0",
      "Commit": "2a42d45c9f8c31aa391315ea24e40c4a7e0b2c1d",
      "ReleaseNotesPage": "https://example.com/releases/2.0.0/notes",
      "LandingPage": "https://example.com/releases/2.0.0"
    }
  ]
}
`

var expectedVersionInfo = VersionInfo{
	Releases: []Release{
		{
			ReleaseCategory:   StableReleaseCategory,
			Version:           semver.MustParse("2.1.0"),
			RolloutProportion: 0.5,
			MinAuto:           semver.MustParse("2.0.0"),
			File: []File{
				{
					URL:        "https://downloads.example.com/v2.1.0/MyApp-2.1.0.pkg",
					Identifier: PackageIdentifier,
				},
				{
					URL:        "https://downloads.example.com/v2.1.0/MyApp-2.1.0.dmg",
					Identifier: InstallerIdentifier,
				},
			},
		},
		{
			ReleaseCategory:   EarlyAccessReleaseCategory,
			Version:           semver.MustParse("2.2.0-beta.1"),
			RolloutProportion: 0.25,
			MinAuto:           semver.MustParse("2.1.0"),
			File: []File{
				{
					URL:        "https://downloads.example.com/beta/v2.2.0-beta.1/MyApp-2.2.0-beta.1.pkg",
					Identifier: PackageIdentifier,
				},
			},
			SystemVersion: versioncompare.SystemVersion{Minimum: "13"},
		},
		{
			ReleaseCategory:   StableReleaseCategory,
			Version:           semver.MustParse("2.0.0"),
			RolloutProportion: 1.0,
			MinAuto:           semver.MustParse("1.9.0"),
			SystemVersion:     versioncompare.SystemVersion{Maximum: "12.0.0", Minimum: "1.0.0"},
			File: []File{
				{
					URL:        "https://downloads.example.com/v2.0.0/MyApp-2.0.0.pkg",
					Identifier: PackageIdentifier,
				},
				{
					URL:        "https://downloads.example.com/v2.0.0/MyApp-2.0.0.dmg",
					Identifier: InstallerIdentifier,
				},
			},
		},
	},
}

func Test_Releases_JsonParse(t *testing.T) {
	var versionInfo VersionInfo
	if err := json.Unmarshal([]byte(mockJSONData), &versionInfo); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	if len(expectedVersionInfo.Releases) != len(versionInfo.Releases) {
		t.Fatalf("expected %d releases, parsed %d releases", len(expectedVersionInfo.Releases), len(versionInfo.Releases))
	}

	for i, expectedRelease := range expectedVersionInfo.Releases {
		release := versionInfo.Releases[i]

		if release.ReleaseCategory != expectedRelease.ReleaseCategory {
			t.Errorf("Release %d: expected category %v, got %v", i, expectedRelease.ReleaseCategory, release.ReleaseCategory)
		}

		if release.Version.String() != expectedRelease.Version.String() {
			t.Errorf("Release %d: expected version %s, got %s", i, expectedRelease.Version, release.Version)
		}

		if release.RolloutProportion != expectedRelease.RolloutProportion {
			t.Errorf("Release %d: expected rollout proportion %f, got %f", i, expectedRelease.RolloutProportion, release.RolloutProportion)
		}

		if expectedRelease.MinAuto != nil && release.MinAuto.String() != expectedRelease.MinAuto.String() {
			t.Errorf("Release %d: expected min auto %s, got %s", i, expectedRelease.MinAuto, release.MinAuto)
		}

		if expectedRelease.SystemVersion.Minimum != release.SystemVersion.Minimum {
			t.Errorf("Release %d: expected system version minimum %s, got %s", i, expectedRelease.SystemVersion.Minimum, release.SystemVersion.Minimum)
		}

		if expectedRelease.SystemVersion.Maximum != release.SystemVersion.Maximum {
			t.Errorf("Release %d: expected system version minimum %s, got %s", i, expectedRelease.SystemVersion.Maximum, release.SystemVersion.Maximum)
		}

		if len(release.File) != len(expectedRelease.File) {
			t.Errorf("Release %d: expected %d files, got %d", i, len(expectedRelease.File), len(release.File))
		}

		for j, expectedFile := range expectedRelease.File {
			file := release.File[j]
			if file.URL != expectedFile.URL {
				t.Errorf("Release %d, File %d: expected URL %s, got %s", i, j, expectedFile.URL, file.URL)
			}
			if file.Identifier != expectedFile.Identifier {
				t.Errorf("Release %d, File %d: expected Identifier %v, got %v", i, j, expectedFile.Identifier, file.Identifier)
			}
		}
	}
}
