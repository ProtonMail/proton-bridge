// Copyright (c) 2022 Proton AG
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

package fakeapi

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"net/textproto"

	"github.com/ProtonMail/proton-bridge/v2/pkg/pmapi"
)

// dataPacketOutlineLightInstagram48png is data packet with encrypted data and
// session key
//
//  gpg: encrypted with 2048-bit RSA key, ID 70B8CA23079F2167, created 2019-09-23
//     "james-test@protonmail.blue <james-test@protonmail.blue>"
//
// If you need to rebuild you can dump KeyPacket string from `CreateAttachment`
// function when called during message sending test.
const dataPacketOutlineLightInstagram48png = `wcBMA3C4yiMHnyFnAQgAnH7Qs4lvnbpwdFh5fTgJVTwKZnoLHcbukczf1dld1h9+83wv1FUNosy08KAX3IbDPGnFf5hMTGyEvEcNI0/2HWgootPPCWVvHKfKjjBmstUxbBgRJWi35bBz0WzMarB4BWM8xO2ffqrylhgUhBdK5c26qZvU6veq4cfkKA4SFT9fld8KHY+Ph+v1Q3lP5p3lLX0CyH0CtX7HxdPXk30J+HkugyYOcQ/2upiGzmnIobmZE9kedEA5CWNQa1gxoQTzuxDOzYRZFQPidywMj8pOPCK+1825O/8IeIL/NbXpb1qEW0roOAjYVO/NQzWyWSuIOc6F/Y+ZxzpjhRhe8son+tLBuQEm1rzWfj5Yg4x56fkRPCtfNKaIfYCGEexQNXZWX8zMxjDWM93JDjLeb0C5FzlY2tME2+zYM/SegzFHiagWlZROgQvoiPxcl67FbcOd1YrNP63gw1dwBt4tg1kbwlcSQ76cHeJ6r/Sjg2Q2v52Ee7h3K5Q2h8UDCcgIfTrS4SAKbWjPRRypMJXBp+GH6LASD4m/A6cjjcuOg5Ssh2KaKjGsYrrHVnplmWfKZ18/OrFYSHKxytiIHrLd3GE7SxbyI6LvfKxa5QAKbPBxL39FjcryaJ9l7iWI63zYhOS6bi6fRWLCq6vK30kPvgn+mivGtAxLfSrgAlODmXPxBM4ZIVQNxYv35Fhxk/ENEvRALRv4Lmdv+lHwK1DdtAY/XozkghpLG9ySThqoktK16BKtPeCAEjktlRUt8x/F3Nl5IhOCxxt21xTnne4erz4g6HtCy1Q4fblJTi5iS4C27MdDAPZLZYxz70vxuoeXK4g61pPNVaYNLRE6vdlQUM0YYy58pHarW1+YQ54HXT5eFP/wz+xvNDW/2RCkvIBxioxiTiJNRzEtfivYGowImkxJkMlhv6g9i17Dz5ANca9y4hEKi+u55drRn1Nfikw8c8JmopFELF1eQ1gbDYdb4X40qV1AApTEH1hRYNZeaMwmLkGDvBUDy99p0i6+BTO/nr9wghRCyxv9urDvNMbGxLnbNq2wdNtYdR4UurjmrpyctqyeaERxUbHhGFqAlkRGbd8ZOFHWrCVpKKYihptSbITZnK4d1vJ7IKu/+nIRI9mwN5IRnJjVo31AvzWpSGnJDtidAjAHeIVwrKk/eONETXwBXuAaVUd8PuMJCv4xuw==`

func newTestAttachment(iAtt int, msgID string) *pmapi.Attachment {
	attID := fmt.Sprintf("attID%d", iAtt)
	attContentID := fmt.Sprintf("<%s@protonmail.com>", attID)
	return &pmapi.Attachment{
		ID:          attID,
		MessageID:   msgID,
		Name:        "outline-light-instagram-48.png",
		MIMEType:    "image/png",
		Disposition: "attachment",
		ContentID:   attContentID,
		Header:      textproto.MIMEHeader{},
		KeyPackets:  dataPacketOutlineLightInstagram48png,
	}
}

func (api *FakePMAPI) GetAttachment(_ context.Context, attachmentID string) (io.ReadCloser, error) {
	if err := api.checkAndRecordCall(GET, "/mail/v4/attachments/"+attachmentID, nil); err != nil {
		return nil, err
	}
	b, err := base64.StdEncoding.DecodeString(dataPacketOutlineLightInstagram48png)
	if err != nil {
		return nil, err
	}
	r := bytes.NewReader(b)
	return ioutil.NopCloser(r), nil
}

func (api *FakePMAPI) CreateAttachment(_ context.Context, attachment *pmapi.Attachment, data io.Reader, signature io.Reader) (*pmapi.Attachment, error) {
	if err := api.checkAndRecordCall(POST, "/mail/v4/attachments", nil); err != nil {
		return nil, err
	}
	bytes, err := ioutil.ReadAll(data)
	if err != nil {
		return nil, err
	}
	attachment.KeyPackets = base64.StdEncoding.EncodeToString(bytes)
	msg := api.getMessage(attachment.MessageID)
	if msg == nil {
		return nil, fmt.Errorf("no such message ID %q", attachment.MessageID)
	}
	msg.Attachments = append(msg.Attachments, attachment)
	return attachment, nil
}
