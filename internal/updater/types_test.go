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
// along with Proton Mail Bridge. If not, see <https://www.gnu.org/licenses/>.

package updater

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_ReleaseCategory_UpdateEligible(t *testing.T) {
	// If release is beta only beta users can update
	require.True(t, EarlyAccessReleaseCategory.UpdateEligible(EarlyChannel))
	require.False(t, EarlyAccessReleaseCategory.UpdateEligible(StableChannel))

	// If the release is stable and is the newest then both beta and stable users can update
	require.True(t, StableReleaseCategory.UpdateEligible(EarlyChannel))
	require.True(t, StableReleaseCategory.UpdateEligible(StableChannel))
}

func Test_ReleaseCategory_JsonUnmarshal(t *testing.T) {
	tests := []struct {
		input    string
		expected ReleaseCategory
		wantErr  bool
	}{
		{
			input:    `{"ReleaseCategory": "EarlyAccess"}`,
			expected: EarlyAccessReleaseCategory,
		},
		{
			input:    `{"ReleaseCategory": "Earlyaccess"}`,
			expected: EarlyAccessReleaseCategory,
		},
		{
			input:    `{"ReleaseCategory": "earlyaccess"}`,
			expected: EarlyAccessReleaseCategory,
		},
		{
			input:    `{"ReleaseCategory": "   earlyaccess   "}`,
			expected: EarlyAccessReleaseCategory,
		},
		{
			input:    `{"ReleaseCategory": "Stable"}`,
			expected: StableReleaseCategory,
		},
		{
			input:    `{"ReleaseCategory": "Stable   "}`,
			expected: StableReleaseCategory,
		},
		{
			input:    `{"ReleaseCategory": "stable"}`,
			expected: StableReleaseCategory,
		},
		{
			input:   `{"ReleaseCategory": "invalid"}`,
			wantErr: true,
		},
	}

	var data struct {
		ReleaseCategory ReleaseCategory
	}

	for _, test := range tests {
		err := json.Unmarshal([]byte(test.input), &data)
		if err != nil && !test.wantErr {
			t.Errorf("json.Unmarshal() error = %v, wantErr %v", err, test.wantErr)
			return
		}

		if test.wantErr && err == nil {
			t.Errorf("expected err got nil")
		}

		if !test.wantErr && data.ReleaseCategory != test.expected {
			t.Errorf("got %v, want %v", data.ReleaseCategory, test.expected)
		}
	}
}

func Test_ReleaseCategory_JsonMarshal(t *testing.T) {
	tests := []struct {
		input struct {
			ReleaseCategory ReleaseCategory `json:"ReleaseCategory"`
		}
		expectedOutput string
		wantErr        bool
	}{
		{
			input: struct {
				ReleaseCategory ReleaseCategory `json:"ReleaseCategory"`
			}{ReleaseCategory: StableReleaseCategory},
			expectedOutput: `{"ReleaseCategory":"Stable"}`,
		},
		{
			input: struct {
				ReleaseCategory ReleaseCategory `json:"ReleaseCategory"`
			}{ReleaseCategory: EarlyAccessReleaseCategory},
			expectedOutput: `{"ReleaseCategory":"EarlyAccess"}`,
		},
		{
			input: struct {
				ReleaseCategory ReleaseCategory `json:"ReleaseCategory"`
			}{ReleaseCategory: 4},
			wantErr: true,
		},
	}

	for _, test := range tests {
		output, err := json.Marshal(test.input)

		if test.wantErr {
			if err == nil && len(output) == 0 {
				t.Errorf("expected error or non-empty output for invalid category")
				return
			}
		} else {
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if string(output) != test.expectedOutput {
				t.Errorf("json.Marshal() = %v, want %v", string(output), test.expectedOutput)
			}
		}
	}
}

func Test_FileIdentifier_JsonUnmarshal(t *testing.T) {
	tests := []struct {
		input    string
		expected FileIdentifier
		wantErr  bool
	}{
		{
			input:    `{"Identifier": "package"}`,
			expected: PackageIdentifier,
		},
		{
			input:    `{"Identifier": "Package"}`,
			expected: PackageIdentifier,
		},
		{
			input:    `{"Identifier": "pACKage"}`,
			expected: PackageIdentifier,
		},
		{
			input:    `{"Identifier": "pACKage    "}`,
			expected: PackageIdentifier,
		},
		{
			input:    `{"Identifier": "installer"}`,
			expected: InstallerIdentifier,
		},
		{
			input:    `{"Identifier": "Installer"}`,
			expected: InstallerIdentifier,
		},
		{
			input:    `{"Identifier": "iNSTaller  "}`,
			expected: InstallerIdentifier,
		},
		{
			input:   `{"Identifier": "error"}`,
			wantErr: true,
		},
	}

	var data struct {
		Identifier FileIdentifier
	}

	for _, test := range tests {
		err := json.Unmarshal([]byte(test.input), &data)
		if err != nil && !test.wantErr {
			t.Errorf("json.Unmarshal() error = %v, wantErr %v", err, test.wantErr)
			return
		}

		if test.wantErr && err == nil {
			t.Errorf("expected err got nil")
		}

		if !test.wantErr && data.Identifier != test.expected {
			t.Errorf("got %v, want %v", data.Identifier, test.expected)
		}
	}
}

func Test_FileIdentifier_JsonMarshal(t *testing.T) {
	tests := []struct {
		input struct {
			Identifier FileIdentifier
		}
		expectedOutput string
		wantErr        bool
	}{
		{
			input: struct {
				Identifier FileIdentifier
			}{Identifier: PackageIdentifier},
			expectedOutput: `{"Identifier":"package"}`,
		},
		{
			input: struct {
				Identifier FileIdentifier
			}{Identifier: InstallerIdentifier},
			expectedOutput: `{"Identifier":"installer"}`,
		},
		{
			input: struct {
				Identifier FileIdentifier
			}{Identifier: 4},
			wantErr: true,
		},
	}

	for _, test := range tests {
		output, err := json.Marshal(test.input)

		if test.wantErr {
			if err == nil && len(output) == 0 {
				t.Errorf("expected error or non-empty output for invalid identifier")
				return
			}
		} else {
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if string(output) != test.expectedOutput {
				t.Errorf("json.Marshal() = %v, want %v", string(output), test.expectedOutput)
			}
		}
	}
}
