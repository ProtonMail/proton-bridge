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

package observability

import (
	"github.com/ProtonMail/proton-bridge/v3/internal/plan"
)

func (d *distinctionUtility) setUserPlan(planName string) {
	if planName == "" {
		return
	}

	d.userPlanLock.Lock()
	defer d.userPlanLock.Unlock()

	userPlanMapped := plan.MapUserPlan(planName)
	if plan.IsHigherPriority(d.userPlanUnsafe, userPlanMapped) {
		d.userPlanUnsafe = userPlanMapped
	}
}

func (d *distinctionUtility) getUserPlanSafe() string {
	d.userPlanLock.Lock()
	defer d.userPlanLock.Unlock()
	return d.userPlanUnsafe
}
