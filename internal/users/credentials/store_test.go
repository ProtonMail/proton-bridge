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

package credentials

import (
	"bytes"
	"encoding/base64"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	r "github.com/stretchr/testify/require"
)

const (
	testSep      = "\n"
	secretFormat = "%v" + testSep + // UserID,
		"%v" + testSep + // Name,
		"%v" + testSep + // Emails,
		"%v" + testSep + // APIToken,
		"%v" + testSep + // Mailbox,
		"%v" + testSep + // BridgePassword,
		"%v" + testSep + // Version string
		"%v" + testSep + // Timestamp,
		"%v" + testSep + // IsHidden,
		"%v" // IsCombinedAddressMode
)

// the best would be to run this test on mac, win, and linux separately

type testCredentials struct {
	UserID,
	Name,
	Emails,
	APIToken,
	Mailbox,
	BridgePassword,
	Version string
	Timestamp int64
	IsHidden,
	IsCombinedAddressMode bool
}

func init() { //nolint:gochecknoinits
	gob.Register(testCredentials{})
}

func (s *testCredentials) MarshalGob() string {
	buf := bytes.Buffer{}
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(s); err != nil {
		return ""
	}
	log.Infof("MarshalGob: %#v\n", buf.String())
	return base64.StdEncoding.EncodeToString(buf.Bytes())
}

func (s *testCredentials) Clear() {
	s.UserID = ""
	s.Name = ""
	s.Emails = ""
	s.APIToken = ""
	s.Mailbox = ""
	s.BridgePassword = ""
	s.Version = ""
	s.Timestamp = 0
	s.IsHidden = false
	s.IsCombinedAddressMode = false
}

func (s *testCredentials) UnmarshalGob(secret string) error {
	s.Clear()
	b, err := base64.StdEncoding.DecodeString(secret)
	if err != nil {
		log.Infoln("decode base64", b)
		return err
	}
	buf := bytes.NewBuffer(b)
	dec := gob.NewDecoder(buf)
	if err = dec.Decode(s); err != nil {
		log.Info("decode gob", b, buf.Bytes())
		return err
	}
	return nil
}

func (s *testCredentials) ToJSON() string {
	if b, err := json.Marshal(s); err == nil {
		log.Infof("MarshalJSON: %#v\n", string(b))
		return base64.StdEncoding.EncodeToString(b)
	}
	return ""
}

func (s *testCredentials) FromJSON(secret string) error {
	b, err := base64.StdEncoding.DecodeString(secret)
	if err != nil {
		return err
	}
	if err = json.Unmarshal(b, s); err == nil {
		return nil
	}
	return err
}

func (s *testCredentials) MarshalFmt() string {
	buf := bytes.Buffer{}
	fmt.Fprintf(
		&buf, secretFormat,
		s.UserID,
		s.Name,
		s.Emails,
		s.APIToken,
		s.Mailbox,
		s.BridgePassword,
		s.Version,
		s.Timestamp,
		s.IsHidden,
		s.IsCombinedAddressMode,
	)
	log.Infof("MarshalFmt: %#v\n", buf.String())
	return base64.StdEncoding.EncodeToString(buf.Bytes())
}

func (s *testCredentials) UnmarshalFmt(secret string) error {
	b, err := base64.StdEncoding.DecodeString(secret)
	if err != nil {
		return err
	}
	buf := bytes.NewBuffer(b)
	log.Infoln("decode fmt", b, buf.Bytes())
	_, err = fmt.Fscanf(
		buf, secretFormat,
		&s.UserID,
		&s.Name,
		&s.Emails,
		&s.APIToken,
		&s.Mailbox,
		&s.BridgePassword,
		&s.Version,
		&s.Timestamp,
		&s.IsHidden,
		&s.IsCombinedAddressMode,
	)
	if err != nil {
		return err
	}
	return nil
}

func (s *testCredentials) MarshalStrings() string { // this is the most space efficient
	items := []string{
		s.UserID,         // 0
		s.Name,           // 1
		s.Emails,         // 2
		s.APIToken,       // 3
		s.Mailbox,        // 4
		s.BridgePassword, // 5
		s.Version,        // 6
	}
	items = append(items, fmt.Sprint(s.Timestamp)) // 7

	if s.IsHidden { // 8
		items = append(items, "1")
	} else {
		items = append(items, "")
	}

	if s.IsCombinedAddressMode { // 9
		items = append(items, "1")
	} else {
		items = append(items, "")
	}

	str := strings.Join(items, sep)

	log.Infof("MarshalJoin: %#v\n", str)
	return base64.StdEncoding.EncodeToString([]byte(str))
}

func (s *testCredentials) UnmarshalStrings(secret string) error {
	b, err := base64.StdEncoding.DecodeString(secret)
	if err != nil {
		return err
	}
	items := strings.Split(string(b), sep)
	if len(items) != 10 {
		return ErrWrongFormat
	}

	s.UserID = items[0]
	s.Name = items[1]
	s.Emails = items[2]
	s.APIToken = items[3]
	s.Mailbox = items[4]
	s.BridgePassword = items[5]
	s.Version = items[6]
	if _, err = fmt.Sscanf(items[7], "%d", &s.Timestamp); err != nil {
		s.Timestamp = 0
	}
	if s.IsHidden = false; items[8] == "1" {
		s.IsHidden = true
	}
	if s.IsCombinedAddressMode = false; items[9] == "1" {
		s.IsCombinedAddressMode = true
	}
	return nil
}

func (s *testCredentials) IsSame(rhs *testCredentials) bool {
	return s.Name == rhs.Name &&
		s.Emails == rhs.Emails &&
		s.APIToken == rhs.APIToken &&
		s.Mailbox == rhs.Mailbox &&
		s.BridgePassword == rhs.BridgePassword &&
		s.Version == rhs.Version &&
		s.Timestamp == rhs.Timestamp &&
		s.IsHidden == rhs.IsHidden &&
		s.IsCombinedAddressMode == rhs.IsCombinedAddressMode
}

func TestMarshalFormats(t *testing.T) {
	input := testCredentials{UserID: "007", Emails: "ja@pm.me;jakub@cu.th", Timestamp: 152469263742, IsHidden: true}
	log.Infof("input %#v\n", input)

	secretStrings := input.MarshalStrings()
	log.Infof("secretStrings %#v %d\n", secretStrings, len(secretStrings))
	secretGob := input.MarshalGob()
	log.Infof("secretGob %#v %d\n", secretGob, len(secretGob))
	secretJSON := input.ToJSON()
	log.Infof("secretJSON %#v %d\n", secretJSON, len(secretJSON))
	secretFmt := input.MarshalFmt()
	log.Infof("secretFmt %#v %d\n", secretFmt, len(secretFmt))

	output := testCredentials{APIToken: "refresh"}
	r.NoError(t, output.UnmarshalStrings(secretStrings))
	log.Infof("strings out %#v \n", output)
	r.True(t, input.IsSame(&output), "strings out not same")

	output = testCredentials{APIToken: "refresh"}
	r.NoError(t, output.UnmarshalGob(secretGob))
	log.Infof("gob out %#v\n \n", output)
	r.Equal(t, input, output)

	output = testCredentials{APIToken: "refresh"}
	r.NoError(t, output.FromJSON(secretJSON))
	log.Infof("json out %#v \n", output)
	r.True(t, input.IsSame(&output), "json out not same")

	/*
		// Simple Fscanf not working!
		output = testCredentials{APIToken: "refresh"}
			r.NoError(t, output.UnmarshalFmt(secretFmt))
			log.Infof("fmt out %#v \n", output)
			r.True(t, input.IsSame(&output), "fmt out not same")
	*/
}

func TestMarshal(t *testing.T) {
	input := Credentials{
		UserID:                "",
		Name:                  "007",
		Emails:                "ja@pm.me;aj@cus.tom",
		APIToken:              "sdfdsfsdfsdfsdf",
		MailboxPassword:       []byte("cdcdcdcd"),
		BridgePassword:        "wew123",
		Version:               "k11",
		Timestamp:             152469263742,
		IsHidden:              true,
		IsCombinedAddressMode: false,
	}
	log.Infof("input %#v\n", input)

	secret := input.Marshal()
	log.Infof("secret %#v %d\n", secret, len(secret))

	output := Credentials{APIToken: "refresh"}
	r.NoError(t, output.Unmarshal(secret))
	log.Infof("output %#v\n", output)
	r.Equal(t, input, output)
}
