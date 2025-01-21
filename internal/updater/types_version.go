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
	"fmt"
	"strings"
)

type ReleaseCategory uint8
type FileIdentifier uint8

const (
	EarlyAccessReleaseCategory ReleaseCategory = iota
	StableReleaseCategory
)

const (
	PackageIdentifier FileIdentifier = iota
	InstallerIdentifier
)

var (
	releaseCategoryName = map[uint8]string{ //nolint:gochecknoglobals
		0: "EarlyAccess",
		1: "Stable",
	}
	releaseCategoryValue = map[string]uint8{ //nolint:gochecknoglobals
		"earlyaccess": 0,
		"stable":      1,
	}
	fileIdentifierName = map[uint8]string{ //nolint:gochecknoglobals
		0: "package",
		1: "installer",
	}
	fileIdentifierValue = map[string]uint8{ //nolint:gochecknoglobals
		"package":   0,
		"installer": 1,
	}
)

func ParseFileIdentifier(s string) (FileIdentifier, error) {
	s = strings.TrimSpace(strings.ToLower(s))
	val, ok := fileIdentifierValue[s]
	if !ok {
		return FileIdentifier(0), fmt.Errorf("%s is not a valid file identifier", s)
	}

	return FileIdentifier(val), nil
}

func (fi FileIdentifier) String() string {
	return fileIdentifierName[uint8(fi)]
}

func (fi FileIdentifier) MarshalJSON() ([]byte, error) {
	return json.Marshal(fi.String())
}

func (fi *FileIdentifier) UnmarshalJSON(data []byte) (err error) {
	var fileIdentifier string
	if err := json.Unmarshal(data, &fileIdentifier); err != nil {
		return err
	}

	parsedFileIdentifier, err := ParseFileIdentifier(fileIdentifier)
	if err != nil {
		return err
	}

	*fi = parsedFileIdentifier
	return nil
}

func ParseReleaseCategory(s string) (ReleaseCategory, error) {
	s = strings.TrimSpace(strings.ToLower(s))
	val, ok := releaseCategoryValue[s]
	if !ok {
		return ReleaseCategory(0), fmt.Errorf("%s is not a valid release category", s)
	}

	return ReleaseCategory(val), nil
}

func (rc ReleaseCategory) String() string {
	return releaseCategoryName[uint8(rc)]
}

func (rc ReleaseCategory) MarshalJSON() ([]byte, error) {
	return json.Marshal(rc.String())
}

func (rc *ReleaseCategory) UnmarshalJSON(data []byte) (err error) {
	var releaseCat string
	if err := json.Unmarshal(data, &releaseCat); err != nil {
		return err
	}

	parsedCat, err := ParseReleaseCategory(releaseCat)
	if err != nil {
		return err
	}

	*rc = parsedCat

	return nil
}

func (rc ReleaseCategory) UpdateEligible(channel Channel) bool {
	if channel == StableChannel && rc == StableReleaseCategory {
		return true
	}

	if channel == EarlyChannel && rc == EarlyAccessReleaseCategory || rc == StableReleaseCategory {
		return true
	}

	return false
}
