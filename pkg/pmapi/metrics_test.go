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
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

const testSendSimpleMetricsBody = `{
    "Code": 1000
}
`

// FIXME(conman): Implement metrics then enable this test.
func _TestClient_SendSimpleMetric(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		Ok(t, checkMethodAndPath(r, "GET", "/metrics?Action=some_action&Category=some_category&Label=some_label"))

		w.Header().Set("Content-Type", "application/json")

		fmt.Fprint(w, testSendSimpleMetricsBody)
	}))
	defer s.Close()

	m := newManager(Config{HostURL: s.URL})

	err := m.SendSimpleMetric(context.TODO(), "some_category", "some_action", "some_label")
	if err != nil {
		t.Fatal("Expected no error while sending simple metric, got:", err)
	}
}
