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
// along with Proton Mail Bridge.  If not, see <https://www.gnu.org/licenses/>.

package observability

import (
	"context"
	"strings"

	"github.com/ProtonMail/gluon/async"
	"github.com/ProtonMail/go-proton-api"
)

const (
	planUnknown    = "unknown"
	planOther      = "other"
	planBusiness   = "business"
	planIndividual = "individual"
	planGroup      = "group"
)

var planHierarchy = map[string]int{ //nolint:gochecknoglobals
	planBusiness:   4,
	planGroup:      3,
	planIndividual: 2,
	planOther:      1,
	planUnknown:    0,
}

type planGetter interface {
	GetOrganizationData(ctx context.Context) (proton.OrganizationResponse, error)
}

func isHigherPriority(currentPlan, newPlan string) bool {
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

func mapUserPlan(planName string) string {
	if planName == "" {
		return planUnknown
	}
	switch strings.TrimSpace(strings.ToLower(planName)) {
	case "mail2022":
		return planIndividual
	case "bundle2022":
		return planIndividual
	case "family2022":
		return planGroup
	case "visionary2022":
		return planGroup
	case "mailpro2022":
		return planBusiness
	case "planbiz2024":
		return planBusiness
	case "bundlepro2022":
		return planBusiness
	case "bundlepro2024":
		return planBusiness
	case "duo2024":
		return planGroup

	default:
		return planOther
	}
}

func (d *distinctionUtility) setUserPlan(planName string) {
	if planName == "" {
		return
	}

	d.userPlanLock.Lock()
	defer d.userPlanLock.Unlock()

	userPlanMapped := mapUserPlan(planName)
	if isHigherPriority(d.userPlanUnsafe, userPlanMapped) {
		d.userPlanUnsafe = userPlanMapped
	}
}

func (d *distinctionUtility) registerUserPlan(ctx context.Context, getter planGetter, panicHandler async.PanicHandler) {
	go func() {
		defer async.HandlePanic(panicHandler)

		orgRes, err := getter.GetOrganizationData(ctx)
		if err != nil {
			return
		}
		d.setUserPlan(orgRes.Organization.PlanName)
	}()
}

func (d *distinctionUtility) getUserPlanSafe() string {
	d.userPlanLock.Lock()
	defer d.userPlanLock.Unlock()
	return d.userPlanUnsafe
}
