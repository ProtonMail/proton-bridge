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

package rfc5322

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestParseDateTime(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{
			input: `Fri, 21 Nov 1997 09:55:06`,
			want:  `1997-11-21T09:55:06Z`,
		},
		{
			input: `Fri, 21 Nov 1997 09:55:06 -0600`,
			want:  `1997-11-21T09:55:06-06:00`,
		},
		{
			input: `Tue, 1 Jul 2003 10:52:37 +0200`,
			want:  `2003-07-01T10:52:37+02:00`,
		},
		{
			input: `Thu, 13 Feb 1969 23:32:54 -0330`,
			want:  `1969-02-13T23:32:54-03:30`,
		},
		{
			input: "Thu, 13 Feb 1969 23:32 -0330 (Newfoundland Time)",
			want:  `1969-02-13T23:32:00-03:30`,
		},
		{
			input: `2 Jan 2006 15:04:05 -0700`,
			want:  `2006-01-02T15:04:05-07:00`,
		},
		{
			input: `2 Jan 2006 15:04:05 MST`,
			want:  `2006-01-02T15:04:05-07:00`,
		},
		{
			input: `2 Jan 2006 15:04 -0700`,
			want:  `2006-01-02T15:04:00-07:00`,
		},
		{
			input: `2 Jan 2006 15:04 MST`,
			want:  `2006-01-02T15:04:00-07:00`,
		},
		{
			input: `2 Jan 06 15:04:05 -0700`,
			want:  `2006-01-02T15:04:05-07:00`,
		},
		{
			input: `2 Jan 06 15:04:05 MST`,
			want:  `2006-01-02T15:04:05-07:00`,
		},
		{
			input: `2 Jan 06 15:04 -0700`,
			want:  `2006-01-02T15:04:00-07:00`,
		},
		{
			input: `2 Jan 06 15:04 MST`,
			want:  `2006-01-02T15:04:00-07:00`,
		},
		{
			input: `02 Jan 2006 15:04:05 -0700`,
			want:  `2006-01-02T15:04:05-07:00`,
		},
		{
			input: `02 Jan 2006 15:04:05 MST`,
			want:  `2006-01-02T15:04:05-07:00`,
		},
		{
			input: `02 Jan 2006 15:04 -0700`,
			want:  `2006-01-02T15:04:00-07:00`,
		},
		{
			input: `02 Jan 2006 15:04 MST`,
			want:  `2006-01-02T15:04:00-07:00`,
		},
		{
			input: `02 Jan 06 15:04:05 -0700`,
			want:  `2006-01-02T15:04:05-07:00`,
		},
		{
			input: `02 Jan 06 15:04:05 MST`,
			want:  `2006-01-02T15:04:05-07:00`,
		},
		{
			input: `02 Jan 06 15:04 -0700`,
			want:  `2006-01-02T15:04:00-07:00`,
		},
		{
			input: `02 Jan 06 15:04 MST`,
			want:  `2006-01-02T15:04:00-07:00`,
		},
		{
			input: `Mon, 2 Jan 2006 15:04:05 -0700`,
			want:  `2006-01-02T15:04:05-07:00`,
		},
		{
			input: `Mon, 2 Jan 2006 15:04:05 MST`,
			want:  `2006-01-02T15:04:05-07:00`,
		},
		{
			input: `Mon, 2 Jan 2006 15:04 -0700`,
			want:  `2006-01-02T15:04:00-07:00`,
		},
		{
			input: `Mon, 2 Jan 2006 15:04 MST`,
			want:  `2006-01-02T15:04:00-07:00`,
		},
		{
			input: `Mon, 2 Jan 06 15:04:05 -0700`,
			want:  `2006-01-02T15:04:05-07:00`,
		},
		{
			input: `Mon, 2 Jan 06 15:04:05 MST`,
			want:  `2006-01-02T15:04:05-07:00`,
		},
		{
			input: `Mon, 2 Jan 06 15:04 -0700`,
			want:  `2006-01-02T15:04:00-07:00`,
		},
		{
			input: `Mon, 2 Jan 06 15:04 MST`,
			want:  `2006-01-02T15:04:00-07:00`,
		},
		{
			input: `Mon, 02 Jan 2006 15:04:05 -0700`,
			want:  `2006-01-02T15:04:05-07:00`,
		},
		{
			input: `Mon, 02 Jan 2006 15:04:05 MST`,
			want:  `2006-01-02T15:04:05-07:00`,
		},
		{
			input: `Mon, 02 Jan 2006 15:04 -0700`,
			want:  `2006-01-02T15:04:00-07:00`,
		},
		{
			input: `Mon, 02 Jan 2006 15:04 MST`,
			want:  `2006-01-02T15:04:00-07:00`,
		},
		{
			input: `Mon, 02 Jan 06 15:04:05 -0700`,
			want:  `2006-01-02T15:04:05-07:00`,
		},
		{
			input: `Mon, 02 Jan 06 15:04:05 MST`,
			want:  `2006-01-02T15:04:05-07:00`,
		},
		{
			input: `Mon, 02 Jan 06 15:04 -0700`,
			want:  `2006-01-02T15:04:00-07:00`,
		},
		{
			input: `Mon, 02 Jan 06 15:04 MST`,
			want:  `2006-01-02T15:04:00-07:00`,
		},
	}
	for _, test := range tests {
		test := test

		t.Run(test.input, func(t *testing.T) {
			got, err := ParseDateTime(test.input)
			assert.NoError(t, err)
			assert.Equal(t, test.want, got.Format(time.RFC3339))
		})
	}
}

func TestParseDateTimeObsolete(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{
			input: `21 Nov 97 09:55:06 GMT`,
			want:  `1997-11-21T09:55:06Z`,
		},
		{
			input: `Wed, 01 Jan 2020 12:00:00 UTC`,
			want:  `2020-01-01T12:00:00Z`,
		},
		{
			input: `Wed, 01 Jan 2020 13:00:00 UTC`,
			want:  `2020-01-01T13:00:00Z`,
		},
		{
			input: `Wed, 01 Jan 2020 12:30:00 UTC`,
			want:  `2020-01-01T12:30:00Z`,
		},
	}
	for _, test := range tests {
		test := test

		t.Run(test.input, func(t *testing.T) {
			got, err := ParseDateTime(test.input)
			assert.NoError(t, err)
			assert.Equal(t, test.want, got.Format(time.RFC3339))
		})
	}
}

func TestParseDateTimeRelaxed(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{
			input: `Mon, 28 Jan 2019 20:59:01 0000`,
			want:  `2019-01-28T20:59:01Z`,
		},
		{
			input: `Mon, 25 Sep 2017 5:25:40 +0200`,
			want:  `2017-09-25T05:25:40+02:00`,
		},
	}
	for _, test := range tests {
		test := test

		t.Run(test.input, func(t *testing.T) {
			got, err := ParseDateTime(test.input)
			assert.NoError(t, err)
			assert.Equal(t, test.want, got.Format(time.RFC3339))
		})
	}
}
