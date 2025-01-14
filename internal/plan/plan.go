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
// along with Proton Mail Bridge.  If not, see <https://www.gnu.org/licenses/>.

package plan

import "strings"

const (
	Unknown    = "unknown"
	Other      = "other"
	Business   = "business"
	Individual = "individual"
	Group      = "group"
)

var planHierarchy = map[string]int{ //nolint:gochecknoglobals
	Business:   4,
	Group:      3,
	Individual: 2,
	Other:      1,
	Unknown:    0,
}

func IsHigherPriority(currentPlan, newPlan string) bool {
	newRank, ok := planHierarchy[newPlan]
	if !ok {
		return false
	}

	currentRank, ok2 := planHierarchy[currentPlan]
	if !ok2 {
		return true // we don't have a valid plan, might as well replace it
	}

	return newRank > currentRank
}

func MapUserPlan(planName string) string {
	if planName == "" {
		return Unknown
	}
	switch strings.TrimSpace(strings.ToLower(planName)) {
	case Individual:
		return Individual
	case Unknown:
		return Unknown
	case Business:
		return Business
	case Group:
		return Group
	case "mail2022":
		return Individual
	case "bundle2022":
		return Individual
	case "family2022":
		return Group
	case "visionary2022":
		return Group
	case "mailpro2022":
		return Business
	case "planbiz2024":
		return Business
	case "bundlepro2022":
		return Business
	case "bundlepro2024":
		return Business
	case "duo2024":
		return Group

	default:
		return Other
	}
}
