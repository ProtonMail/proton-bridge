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

package message

import (
	"bytes"
	"encoding/base64"
	"io"
	"io/ioutil"
	"mime"
	"mime/quotedprintable"
	"strings"

	"github.com/ProtonMail/gopenpgp/v2/crypto"
	pmmime "github.com/ProtonMail/proton-bridge/v2/pkg/mime"
	"github.com/emersion/go-message/textproto"
	"github.com/pkg/errors"
)

func EncryptRFC822(kr *crypto.KeyRing, r io.Reader) ([]byte, error) {
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	header, body, err := readHeaderBody(b)
	if err != nil {
		return nil, err
	}

	buf := new(bytes.Buffer)

	result, err := writeEncryptedPart(kr, header, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	if err := textproto.WriteHeader(buf, *header); err != nil {
		return nil, err
	}

	if _, err := result.WriteTo(buf); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func writeEncryptedPart(kr *crypto.KeyRing, header *textproto.Header, r io.Reader) (io.WriterTo, error) {
	decoder := getTransferDecoder(r, header.Get("Content-Transfer-Encoding"))
	encoded := new(bytes.Buffer)

	contentType, contentParams, err := parseContentType(header.Get("Content-Type"))
	// Ignoring invalid media parameter makes it work for invalid tutanota RFC2047-encoded attachment filenames since we often only really need the content type and not the optional media parameters.
	if err != nil && !errors.Is(err, mime.ErrInvalidMediaParameter) {
		return nil, err
	}

	switch {
	case contentType == "", strings.HasPrefix(contentType, "text/"), strings.HasPrefix(contentType, "message/"):
		header.Del("Content-Transfer-Encoding")

		if charset, ok := contentParams["charset"]; ok {
			if reader, err := pmmime.CharsetReader(charset, decoder); err == nil {
				decoder = reader

				// We can decode the charset to utf-8 so let's set that as the content type charset parameter.
				contentParams["charset"] = "utf-8"

				header.Set("Content-Type", mime.FormatMediaType(contentType, contentParams))
			}
		}

		if err := encode(&writeCloser{encoded}, func(w io.Writer) error {
			return writeEncryptedTextPart(w, decoder, kr)
		}); err != nil {
			return nil, err
		}

	case contentType == "multipart/encrypted":
		if _, err := encoded.ReadFrom(decoder); err != nil {
			return nil, err
		}

	case strings.HasPrefix(contentType, "multipart/"):
		if err := encode(&writeCloser{encoded}, func(w io.Writer) error {
			return writeEncryptedMultiPart(kr, w, header, decoder)
		}); err != nil {
			return nil, err
		}

	default:
		header.Set("Content-Transfer-Encoding", "base64")

		if err := encode(base64.NewEncoder(base64.StdEncoding, encoded), func(w io.Writer) error {
			return writeEncryptedAttachmentPart(w, decoder, kr)
		}); err != nil {
			return nil, err
		}
	}

	return encoded, nil
}

func writeEncryptedTextPart(w io.Writer, r io.Reader, kr *crypto.KeyRing) error {
	dec, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}

	var arm string

	if msg, err := crypto.NewPGPMessageFromArmored(string(dec)); err != nil {
		enc, err := kr.Encrypt(crypto.NewPlainMessage(dec), kr)
		if err != nil {
			return err
		}

		if arm, err = enc.GetArmored(); err != nil {
			return err
		}
	} else if arm, err = msg.GetArmored(); err != nil {
		return err
	}

	if _, err := io.WriteString(w, arm); err != nil {
		return err
	}

	return nil
}

func writeEncryptedAttachmentPart(w io.Writer, r io.Reader, kr *crypto.KeyRing) error {
	dec, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}

	enc, err := kr.Encrypt(crypto.NewPlainMessage(dec), kr)
	if err != nil {
		return err
	}

	if _, err := w.Write(enc.GetBinary()); err != nil {
		return err
	}

	return nil
}

func writeEncryptedMultiPart(kr *crypto.KeyRing, w io.Writer, header *textproto.Header, r io.Reader) error {
	_, contentParams, err := parseContentType(header.Get("Content-Type"))
	if err != nil {
		return err
	}

	scanner, err := newPartScanner(r, contentParams["boundary"])
	if err != nil {
		return err
	}

	parts, err := scanner.scanAll()
	if err != nil {
		return err
	}

	writer := newPartWriter(w, contentParams["boundary"])

	for _, part := range parts {
		header, body, err := readHeaderBody(part.b)
		if err != nil {
			return err
		}

		result, err := writeEncryptedPart(kr, header, bytes.NewReader(body))
		if err != nil {
			return err
		}

		if err := writer.createPart(func(w io.Writer) error {
			if err := textproto.WriteHeader(w, *header); err != nil {
				return err
			}

			if _, err := result.WriteTo(w); err != nil {
				return err
			}

			return nil
		}); err != nil {
			return err
		}
	}

	return writer.done()
}

func getTransferDecoder(r io.Reader, encoding string) io.Reader {
	switch strings.ToLower(encoding) {
	case "base64":
		return base64.NewDecoder(base64.StdEncoding, r)

	case "quoted-printable":
		return quotedprintable.NewReader(r)

	default:
		return r
	}
}

func encode(wc io.WriteCloser, fn func(io.Writer) error) error {
	if err := fn(wc); err != nil {
		return err
	}

	return wc.Close()
}

type writeCloser struct {
	io.Writer
}

func (writeCloser) Close() error { return nil }

func parseContentType(val string) (string, map[string]string, error) {
	if val == "" {
		val = "text/plain"
	}

	return pmmime.ParseMediaType(val)
}
