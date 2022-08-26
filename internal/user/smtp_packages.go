package user

import (
	"github.com/ProtonMail/gluon/rfc822"
	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"github.com/ProtonMail/proton-bridge/v2/pkg/message"
	"github.com/bradenaw/juniper/xslices"
	"gitlab.protontech.ch/go/liteapi"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
)

func createSendReq(
	kr *crypto.KeyRing,
	mimeBody message.MIMEBody,
	richBody, plainBody message.Body,
	recipients recipients,
	attKeys map[string]*crypto.SessionKey,
) (liteapi.SendDraftReq, error) {
	var req liteapi.SendDraftReq

	if recs := recipients.scheme(liteapi.PGPMIMEScheme, liteapi.ClearMIMEScheme); len(recs) > 0 {
		if err := req.AddMIMEPackage(kr, string(mimeBody), recs); err != nil {
			return liteapi.SendDraftReq{}, err
		}
	}

	if recs := recipients.scheme(liteapi.InternalScheme, liteapi.ClearScheme, liteapi.PGPInlineScheme); len(recs) > 0 {
		if recs := recs.content(rfc822.TextHTML); len(recs) > 0 {
			if err := req.AddPackage(kr, string(richBody), rfc822.TextHTML, recs, attKeys); err != nil {
				return liteapi.SendDraftReq{}, err
			}
		}

		if recs := recs.content(rfc822.TextPlain); len(recs) > 0 {
			if err := req.AddPackage(kr, string(plainBody), rfc822.TextPlain, recs, attKeys); err != nil {
				return liteapi.SendDraftReq{}, err
			}
		}
	}

	return req, nil
}

type recipients map[string]liteapi.SendPreferences

func (r recipients) scheme(scheme ...liteapi.EncryptionScheme) recipients {
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
