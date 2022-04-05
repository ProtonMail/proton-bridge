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
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	r "github.com/stretchr/testify/require"
)

const testSendSimpleMetricsBody = `{
    "Code": 1000
}
`

func TestClient_SendSimpleMetric(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		r.NoError(t, checkMethodAndPath(req, "GET", "/metrics?Action=some_action&Category=some_category&Label=some_label"))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, testSendSimpleMetricsBody)
	}))
	defer s.Close()

	m := newManager(newTestConfig(s.URL))

	err := m.SendSimpleMetric(context.Background(), "some_category", "some_action", "some_label")
	r.NoError(t, err)
}
