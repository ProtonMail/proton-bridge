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

package transfer

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	r "github.com/stretchr/testify/require"
)

func TestGetFolderNames(t *testing.T) {
	root, clean := createTestingFolderStructure(t)
	defer clean()

	tests := []struct {
		suffix    string
		wantNames []string
	}{
		{
			"",
			[]string{
				"bar",
				"baz",
				filepath.Base(root),
				"foo",
				"qwerty",
				"test",
			},
		},
		{
			".eml",
			[]string{
				"bar",
				"baz",
				filepath.Base(root),
				"foo",
			},
		},
		{
			".txt",
			[]string{
				filepath.Base(root),
			},
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.suffix, func(t *testing.T) {
			names, err := getFolderNamesWithFileSuffix(root, tc.suffix)
			r.NoError(t, err)
			r.Equal(t, tc.wantNames, names)
		})
	}
}

func TestGetFilePathsWithSuffix(t *testing.T) {
	root, clean := createTestingFolderStructure(t)
	defer clean()

	tests := []struct {
		suffix    string
		wantPaths []string
	}{
		{
			".eml",
			[]string{
				"foo/bar/baz/msg1.eml",
				"foo/bar/baz/msg2.eml",
				"foo/bar/baz/msg3.eml",
				"foo/bar/msg4.eml",
				"foo/bar/msg5.eml",
				"foo/baz/msg6.eml",
				"foo/msg7.eml",
				"msg10.eml",
				"test/foo/msg8.eml",
				"test/foo/msg9.eml",
			},
		},
		{
			".txt",
			[]string{
				"info.txt",
			},
		},
		{
			".hello",
			[]string{},
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.suffix, func(t *testing.T) {
			paths, err := getFilePathsWithSuffix(root, tc.suffix)
			r.NoError(t, err)
			r.Equal(t, tc.wantPaths, paths)
		})
	}
}

func createTestingFolderStructure(t *testing.T) (string, func()) {
	root, err := ioutil.TempDir("", "folderstructure")
	r.NoError(t, err)

	for _, path := range []string{
		"foo/bar/baz",
		"foo/baz",
		"test/foo",
		"qwerty",
	} {
		err = os.MkdirAll(filepath.Join(root, path), os.ModePerm)
		r.NoError(t, err)
	}

	for _, path := range []string{
		"foo/bar/baz/msg1.eml",
		"foo/bar/baz/msg2.eml",
		"foo/bar/baz/msg3.eml",
		"foo/bar/msg4.eml",
		"foo/bar/msg5.eml",
		"foo/baz/msg6.eml",
		"foo/msg7.eml",
		"test/foo/msg8.eml",
		"test/foo/msg9.eml",
		"msg10.eml",
		"info.txt",
	} {
		f, err := os.Create(filepath.Join(root, path))
		r.NoError(t, err)
		err = f.Close()
		r.NoError(t, err)
	}

	return root, func() {
		_ = os.RemoveAll(root)
	}
}

func TestGetMessageTime(t *testing.T) {
	tests := []struct {
		body     string
		wantTime int64
		wantErr  string
	}{
		{"", 0, "failed to read headers: EOF"},
		{"Subject: hello\n\n", 0, ""},
		{"Date: Thu, 23 Apr 2020 04:52:44 +0000\n\n", 1587617564, ""},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.body, func(t *testing.T) {
			time, err := getMessageTime([]byte(tc.body))
			if tc.wantErr == "" {
				r.NoError(t, err)
			} else {
				r.EqualError(t, err, tc.wantErr)
			}
			r.Equal(t, tc.wantTime, time)
		})
	}
}

func TestGetMessageHeader(t *testing.T) {
	body := `Subject: Hello
From: user@example.com

Body
`
	header, err := getMessageHeader([]byte(body))
	r.NoError(t, err)
	r.Equal(t, header.Get("subject"), "Hello")
	r.Equal(t, header.Get("from"), "user@example.com")
}
