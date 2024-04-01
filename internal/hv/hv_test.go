// Copyright (c) 2024 Proton AG
//
// This file is part of Proton Mail Bridge.Bridge.
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

package hv

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/ProtonMail/go-proton-api"
	"github.com/stretchr/testify/require"
)

func TestVerifyAndExtractHvRequest(t *testing.T) {
	det1, _ := json.Marshal("test")
	det2, _ := json.Marshal(proton.APIHVDetails{Methods: []string{"ownership-email"}, Token: "test"})
	det3, _ := json.Marshal(proton.APIHVDetails{Methods: []string{"captcha"}, Token: "test"})
	det4, _ := json.Marshal(proton.APIHVDetails{Methods: []string{"ownership-email", "test"}, Token: "test"})
	det5, _ := json.Marshal(proton.APIHVDetails{Methods: []string{"captcha", "ownership-email"}, Token: "test"})

	tests := []struct {
		err          error
		hasHvDetails bool
		hasErr       bool
	}{
		{err: nil,
			hasHvDetails: false,
			hasErr:       false},
		{err: fmt.Errorf("test"),
			hasHvDetails: false,
			hasErr:       false},
		{err: new(proton.APIError),
			hasHvDetails: false,
			hasErr:       false},
		{err: &proton.APIError{Status: 429},
			hasHvDetails: false,
			hasErr:       false},
		{err: &proton.APIError{Status: 9001},
			hasHvDetails: false,
			hasErr:       false},
		{err: &proton.APIError{Code: 9001},
			hasHvDetails: false,
			hasErr:       true},
		{err: &proton.APIError{Code: 9001, Details: det1},
			hasHvDetails: false,
			hasErr:       true},
		{err: &proton.APIError{Code: 9001, Details: det2},
			hasHvDetails: true,
			hasErr:       false},
		{err: &proton.APIError{Code: 9001, Details: det3},
			hasHvDetails: true,
			hasErr:       false},
		{err: &proton.APIError{Code: 9001, Details: det4},
			hasHvDetails: true,
			hasErr:       false},
		{err: &proton.APIError{Code: 9001, Details: det5},
			hasHvDetails: true,
			hasErr:       false},
	}

	for _, test := range tests {
		hvDetails, err := VerifyAndExtractHvRequest(test.err)
		hasHv := hvDetails != nil
		hasErr := err != nil
		require.True(t, hasHv == test.hasHvDetails && hasErr == test.hasErr)
	}
}

func TestIsHvRequest(t *testing.T) {
	tests := []struct {
		err    error
		result bool
	}{
		{
			err:    nil,
			result: false,
		},
		{
			err:    fmt.Errorf("test"),
			result: false,
		},
		{
			err:    new(proton.APIError),
			result: false,
		},
		{
			err:    &proton.APIError{Status: 429},
			result: false,
		},
		{
			err:    &proton.APIError{Status: 9001},
			result: false,
		},
		{
			err:    &proton.APIError{Code: 9001},
			result: true,
		},
	}

	for _, test := range tests {
		isHvRequest := IsHvRequest(test.err)
		require.Equal(t, test.result, isHvRequest)
	}
}

func TestFormatHvURL(t *testing.T) {
	tests := []struct {
		details *proton.APIHVDetails
		result  string
	}{
		{
			details: &proton.APIHVDetails{Methods: []string{"test"}, Token: "test"},
			result:  "https://verify.proton.me/?methods=test&token=test",
		},
		{
			details: &proton.APIHVDetails{Methods: []string{""}, Token: "test"},
			result:  "https://verify.proton.me/?methods=&token=test",
		},
		{
			details: &proton.APIHVDetails{Methods: []string{"test"}, Token: ""},
			result:  "https://verify.proton.me/?methods=test&token=",
		},
	}

	for _, el := range tests {
		result := FormatHvURL(el.details)
		require.Equal(t, el.result, result)
	}
}
