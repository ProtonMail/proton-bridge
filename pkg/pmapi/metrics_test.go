// Copyright (c) 2020 Proton Technologies AG
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
	"fmt"
	"net/http"
	"testing"
)

const testSendSimpleMetricsBody = `{
    "Code": 1000
}
`

func TestClient_SendSimpleMetric(t *testing.T) {
	s, c := newTestServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		Ok(t, checkMethodAndPath(r, "GET", "/metrics?Action=some_action&Category=some_category&Label=some_label"))

		fmt.Fprint(w, testSendSimpleMetricsBody)
	}))
	defer s.Close()

	err := c.SendSimpleMetric("some_category", "some_action", "some_label")
	if err != nil {
		t.Fatal("Expected no error while sending simple metric, got:", err)
	}
}
