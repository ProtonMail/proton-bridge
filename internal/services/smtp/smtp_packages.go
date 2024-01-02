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
// along with Proton Mail Bridge.  If not, see <https://www.gnu.org/licenses/>.

package smtp

import (
	"fmt"

	"github.com/ProtonMail/gluon/rfc822"
	"github.com/ProtonMail/go-proton-api"
	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"github.com/ProtonMail/proton-bridge/v3/pkg/message"
	"github.com/bradenaw/juniper/xslices"
	"github.com/sirupsen/logrus"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
)

func createSendReq(
	kr *crypto.KeyRing,
	mimeBody message.MIMEBody,
	richBody, plainBody message.Body,
	recipients recipients,
	attKeys map[string]*crypto.SessionKey,
) (proton.SendDraftReq, error) {
	var req proton.SendDraftReq

	if recs := recipients.scheme(proton.PGPMIMEScheme, proton.ClearMIMEScheme); len(recs) > 0 {
		if err := req.AddMIMEPackage(kr, string(mimeBody), recs); err != nil {
			return proton.SendDraftReq{}, err
		}
	}

	if recs := recipients.scheme(proton.InternalScheme, proton.ClearScheme, proton.PGPInlineScheme); len(recs) > 0 {
		if recs := recipients.scheme(proton.PGPInlineScheme); len(recs) > 0 {
			logrus.WithFields(logrus.Fields{"service": "smtp", "settings": "recipient"}).Warn("PGPInline scheme used. Planed to be deprecated.")
		}
		if recs := recs.content(rfc822.TextHTML); len(recs) > 0 {
			if err := req.AddTextPackage(kr, string(richBody), rfc822.TextHTML, recs, attKeys); err != nil {
				return proton.SendDraftReq{}, err
			}
		}

		if recs := recs.content(rfc822.TextPlain); len(recs) > 0 {
			if err := req.AddTextPackage(kr, string(plainBody), rfc822.TextPlain, recs, attKeys); err != nil {
				return proton.SendDraftReq{}, err
			}
		}

		if recs := recs.content(rfc822.MultipartMixed); len(recs) > 0 {
			return proton.SendDraftReq{}, fmt.Errorf("invalid MIME type for MIME package: %s", rfc822.MultipartMixed)
		}
	}

	return req, nil
}

type recipients map[string]proton.SendPreferences

func (r recipients) scheme(scheme ...proton.EncryptionScheme) recipients {
	res := make(recipients)

	for _, addr := range xslices.Filter(maps.Keys(r), func(addr string) bool {
		return slices.Contains(scheme, r[addr].EncryptionScheme)
	}) {
		res[addr] = r[addr]
	}

	return res
}

func (r recipients) content(mimeType ...rfc822.MIMEType) recipients {
	res := make(recipients)

	for _, addr := range xslices.Filter(maps.Keys(r), func(addr string) bool {
		return slices.Contains(mimeType, r[addr].MIMEType)
	}) {
		res[addr] = r[addr]
	}

	return res
}
