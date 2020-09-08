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

package updates

import (
	"testing"

	"github.com/stretchr/testify/require"
)

type testDataValues struct {
	expectErr, expectedNewer bool
	first, second            string
}
type testDataList []testDataValues

func (tdl *testDataList) add(err, newer bool, first, second string) { //nolint[unparam]
	*tdl = append(*tdl, testDataValues{err, newer, first, second})
}

func (tdl *testDataList) addFirstIsNewer(first, second string) {
	tdl.add(false, true, first, second)
	tdl.add(false, false, second, first)
}

func TestCompareVersion(t *testing.T) {
	testData := testDataList{}
	// same is never newer
	testData.add(false, false, "1.1.1", "1.1.1")
	testData.add(false, false, "1.1.0", "1.1")
	testData.add(false, false, "1.0.0", "1")
	testData.add(false, false, ".1.1", "0.1.1")
	testData.add(false, false, "0.1.1", ".1.1")

	testData.addFirstIsNewer("1.1.10", "1.1.1")
	testData.addFirstIsNewer("1.10.1", "1.1.1")
	testData.addFirstIsNewer("10.1.1", "1.1.1")

	testData.addFirstIsNewer("1.1.1", "0.1.1")
	testData.addFirstIsNewer("1.1.1", "1.0.1")
	testData.addFirstIsNewer("1.1.1", "1.1.0")

	testData.addFirstIsNewer("1.1.1", "1")
	testData.addFirstIsNewer("1.1.1", "1.1")
	testData.addFirstIsNewer("1.1.1.1", "1.1.1")

	testData.addFirstIsNewer("1.1.1 beta", "1.1.0")
	testData.addFirstIsNewer("1z.1z.1z", "1.1.0")
	testData.addFirstIsNewer("1a.1b.1c", "1.1.0")

	for _, td := range testData {
		t.Log(td)
		isNewer, err := isFirstVersionNewer(td.first, td.second)
		if td.expectErr {
			require.True(t, err != nil, "expected error but got nil for %#v", td)
			require.True(t, true == isNewer, "error expected but first is not newer for %#v", td)
			continue
		}

		require.True(t, err == nil, "expected no error but have %v for %#v", err, td)
		require.True(t, isNewer == td.expectedNewer, "expected %v but have %v for %#v", td.expectedNewer, isNewer, err, td)
	}
}
