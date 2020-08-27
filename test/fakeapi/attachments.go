// Copyright (c) 2020 Proton Technologies AG
//
// This file is part of ProtonMail Bridge.Bridge.
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

package fakeapi

import (
	"encoding/base64"
	"io"
	"io/ioutil"
	"strings"

	"github.com/ProtonMail/proton-bridge/pkg/pmapi"
)

func (api *FakePMAPI) GetAttachment(attachmentID string) (io.ReadCloser, error) {
	if err := api.checkAndRecordCall(GET, "/mail/v4/attachments/"+attachmentID, nil); err != nil {
		return nil, err
	}
	data := strings.NewReader("data")
	return ioutil.NopCloser(data), nil
}

func (api *FakePMAPI) CreateAttachment(attachment *pmapi.Attachment, data io.Reader, signature io.Reader) (*pmapi.Attachment, error) {
	if err := api.checkAndRecordCall(POST, "/mail/v4/attachments", nil); err != nil {
		return nil, err
	}
	bytes, err := ioutil.ReadAll(data)
	if err != nil {
		return nil, err
	}
	attachment.KeyPackets = base64.StdEncoding.EncodeToString(bytes)
	return attachment, nil
}

func (api *FakePMAPI) DeleteAttachment(attID string) error {
	if err := api.checkAndRecordCall(DELETE, "/mail/v4/attachments/"+attID, nil); err != nil {
		return err
	}
	return nil
}
