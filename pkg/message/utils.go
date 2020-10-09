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

package message

import (
	"bytes"
	"html/template"
	"io"
	"net/http"
	"net/textproto"

	"github.com/ProtonMail/proton-bridge/pkg/pmapi"
	mail "github.com/ProtonMail/proton-bridge/pkg/net/mail"
)

func WriteHeader(w io.Writer, h textproto.MIMEHeader) (err error) {
	if err = http.Header(h).Write(w); err != nil {
		return
	}
	_, err = io.WriteString(w, "\r\n")
	return
}

const customMessageTemplate = `
<html>
	<head></head>
	<body style="font-family: Arial,'Helvetica Neue',Helvetica,sans-serif; font-size: 14px;">
		<div style="color:#555; background-color:#cf9696; padding:20px; border-radius: 4px;">
			<strong>Decryption error</strong><br/>
			Decryption of this message's encrypted content failed.
			<pre>{{.Error}}</pre>
		</div>

		{{if .AttachBody}}
		<div style="color:#333; background-color:#f4f4f4;  border: 1px solid #acb0bf; border-radius: 2px; padding:1rem; margin:1rem 0; font-family:monospace; font-size: 1em;">
			<pre>{{.Body}}</pre>
		</div>
		{{- end}}
	</body>
</html>
`

type customMessageData struct {
	Error      string
	AttachBody bool
	Body       string
}

func CustomMessage(m *pmapi.Message, decodeError error, attachBody bool) error {
	t := template.Must(template.New("customMessage").Parse(customMessageTemplate))

	b := new(bytes.Buffer)

	if err := t.Execute(b, customMessageData{
		Error:      decodeError.Error(),
		AttachBody: attachBody,
		Body:       m.Body,
	}); err != nil {
		return err
	}

	m.MIMEType = pmapi.ContentTypeHTML
	m.Body = b.String()

	// NOTE: we need to set header in custom message header, so we check that is non-nil.
	if m.Header == nil {
		m.Header = make(mail.Header)
	}
	return nil
}
