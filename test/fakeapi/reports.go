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

package fakeapi

import (
	"net/url"

	"github.com/ProtonMail/proton-bridge/pkg/pmapi"
)

func (api *FakePMAPI) Report(report pmapi.ReportReq) error {
	return api.checkInternetAndRecordCall(POST, "/reports/bug", report)
}

func (api *FakePMAPI) SendSimpleMetric(category, action, label string) error {
	v := url.Values{}
	v.Set("Category", category)
	v.Set("Action", action)
	v.Set("Label", label)
	return api.checkInternetAndRecordCall(GET, "/metrics?"+v.Encode(), nil)
}

func (api *FakePMAPI) ReportSentryCrash(err error) error {
	return nil
}
