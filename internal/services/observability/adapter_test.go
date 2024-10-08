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
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func Test_AdapterCustomMetrics(t *testing.T) {
	customMetric := map[string]interface{}{
		"Name":      "name",
		"Version":   1,
		"Timestamp": time.Now().Unix(),
		"Data": map[string]interface{}{
			"Value": 1,
			"Labels": map[string]string{
				"error": "customError",
			},
		},
	}

	ok, metric := VerifyAndParseGenericMetrics(customMetric)
	require.True(t, ok)

	require.Equal(t, metric.Name, customMetric["Name"])
	require.Equal(t, metric.Timestamp, customMetric["Timestamp"])
	require.Equal(t, metric.Version, customMetric["Version"])
	require.Equal(t, metric.Data, customMetric["Data"])
}

func Test_AdapterGluonMetrics(t *testing.T) {
	metrics := GenerateAllGluonMetrics()

	for _, metric := range metrics {
		ok, m := VerifyAndParseGenericMetrics(metric)
		fmt.Println(m)
		require.True(t, ok)
	}
}
