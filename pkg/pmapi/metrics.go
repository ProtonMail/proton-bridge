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

package pmapi

import (
	"net/url"
)

// SendSimpleMetric makes a simple GET request to send a simple metrics report.
func (c *client) SendSimpleMetric(category, action, label string) (err error) {
	v := url.Values{}
	v.Set("Category", category)
	v.Set("Action", action)
	v.Set("Label", label)

	req, err := c.NewRequest("GET", "/metrics?"+v.Encode(), nil)
	if err != nil {
		return
	}

	var res Res
	if err = c.DoJSON(req, &res); err != nil {
		return
	}

	err = res.Err()
	return
}
