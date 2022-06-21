// Copyright (c) 2022 Proton AG
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

package pmapi

import (
	"context"
)

// Report sends request as json or multipart (if has attachment).
func (m *manager) ReportBug(ctx context.Context, rep ReportBugReq) error {
	if rep.ClientType == 0 {
		rep.ClientType = EmailClientType
	}

	if rep.Client == "" {
		rep.Client = m.cfg.GetUserAgent()
	}

	if rep.ClientVersion == "" {
		rep.ClientVersion = m.cfg.AppVersion
	}

	r := m.r(ctx).SetMultipartFormData(rep.GetMultipartFormData())

	for _, att := range rep.Attachments {
		r = r.SetMultipartField(att.name, att.name, att.mime, att.body)
	}

	if _, err := wrapNoConnection(r.Post("/reports/bug")); err != nil {
		return err
	}

	return nil
}
