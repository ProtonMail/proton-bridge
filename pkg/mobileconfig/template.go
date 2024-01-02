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
// along with Proton Mail Bridge. If not, see <https://www.gnu.org/licenses/>.

package mobileconfig

const mailTemplate = `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
  <dict>
    <key>PayloadContent</key>
    <array>
      <dict>
        {{- if .AccountName}}
        <key>EmailAccountName</key>
        <string>{{.AccountName}}</string>
        {{- end}}
        {{- if .AccountDescription}}
        <key>EmailAccountDescription</key>
        <string>{{.AccountDescription}}</string>
        {{- end}}

        {{- if .IMAP}}
        <key>EmailAccountType</key>
        <string>EmailTypeIMAP</string>

        <key>EmailAddress</key>
        <string>{{.EmailAddress}}</string>

        <key>IncomingMailServerAuthentication</key>
        <string>EmailAuthPassword</string>

        <key>IncomingMailServerHostName</key>
        <string>{{.IMAP.Hostname}}</string>

        <key>IncomingMailServerPortNumber</key>
        <integer>{{.IMAP.Port}}</integer>

        <key>IncomingMailServerUseSSL</key>
        {{- if .IMAP.TLS}}
        <true/>
        {{- else}}
        <false/>
        {{- end}}

        <key>IncomingMailServerUsername</key>
        <string>{{.IMAP.Username}}</string>

        <key>IncomingPassword</key>
        <string>{{.IMAP.Password}}</string>
        {{- end}}

        {{ if .SMTP}}
        <key>OutgoingMailServerAuthentication</key>
        <string>{{if .SMTP.Username}}EmailAuthPassword{{else}}EmailAuthNone{{end}}</string>

        <key>OutgoingMailServerHostName</key>
        <string>{{.SMTP.Hostname}}</string>

        <key>OutgoingMailServerPortNumber</key>
        <integer>{{.SMTP.Port}}</integer>

        <key>OutgoingMailServerUseSSL</key>
        {{- if .SMTP.TLS}}
        <true/>
        {{- else}}
        <false/>
        {{- end}}

        {{- if .SMTP.Username}}
        <key>OutgoingMailServerUsername</key>
        <string>{{.SMTP.Username}}</string>
        {{- end}}

        {{- if .SMTP.Password}}
        <key>OutgoingPassword</key>
        <string>{{.SMTP.Password}}</string>
        {{- else}}
        <key>OutgoingPasswordSameAsIncomingPassword</key>
        <true/>
        {{- end}}
        {{end}}

        <key>PayloadDescription</key>
        <string>Configures email account.</string>

        <key>PayloadDisplayName</key>
        <string>{{.DisplayName}}</string>

        <key>PayloadIdentifier</key>
        <string>{{.Identifier}}</string>

        {{- if .Organization}}
        <key>PayloadOrganization</key>
        <string>{{.Organization}}</string>
        {{- end}}

        <key>PayloadType</key>
        <string>com.apple.mail.managed</string>

        <key>PayloadUUID</key>
        <string>{{.ContentUUID}}</string>

        <key>PayloadVersion</key>
        <integer>1</integer>

        <key>PreventAppSheet</key>
        <false/>

        <key>PreventMove</key>
        <false/>

        <key>SMIMEEnabled</key>
        <false/>
      </dict>
    </array>

    <key>PayloadDescription</key>
    <string>{{if .Description}}{{.Description}}{{else}}Install this profile to auto configure email account for {{.EmailAddress}}.{{- end}}</string>

    <key>PayloadDisplayName</key>
    <string>{{.DisplayName}}</string>

    <key>PayloadIdentifier</key>
    <string>{{.Identifier}}</string>

    {{- if .Organization}}
    <key>PayloadOrganization</key>
    <string>{{.Organization}}</string>
    {{- end}}

    <key>PayloadRemovalDisallowed</key>
    <false/>

    <key>PayloadType</key>
    <string>Configuration</string>

    <key>PayloadUUID</key>
    <string>{{.UUID}}</string>

    <key>PayloadVersion</key>
    <integer>1</integer>
  </dict>
</plist>`
