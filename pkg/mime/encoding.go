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

package pmmime

import (
	"fmt"
	"io"
	"mime"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/html/charset"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/htmlindex"
)

var WordDec = &mime.WordDecoder{
	CharsetReader: func(charset string, input io.Reader) (io.Reader, error) {
		dec, err := SelectDecoder(charset)
		if err != nil {
			return nil, err
		}
		if dec == nil { // utf-8
			return input, nil
		}
		return dec.Reader(input), nil
	},
}

// Expects trimmed lowercase.
func getEncoding(charset string) (enc encoding.Encoding, err error) {
	preparsed := strings.Trim(strings.ToLower(charset), " \t\r\n")

	// koi
	re := regexp.MustCompile("(cs)?koi[-_ ]?8?[-_ ]?(r|ru|u|uk)?$")
	matches := re.FindAllStringSubmatch(preparsed, -1)
	if len(matches) == 1 && len(matches[0]) == 3 {
		preparsed = "koi8-"
		switch matches[0][2] {
		case "u", "uk":
			preparsed += "u"
		default:
			preparsed += "r"
		}
	}

	// windows-XXXX
	re = regexp.MustCompile("(cp|(cs)?win(dows)?)[-_ ]?([0-9]{3,4})$")
	matches = re.FindAllStringSubmatch(preparsed, -1)
	if len(matches) == 1 && len(matches[0]) == 5 {
		switch matches[0][4] {
		case "874", "1250", "1251", "1252", "1253", "1254", "1255", "1256", "1257", "1258":
			preparsed = "windows-" + matches[0][4]
		}
	}

	// iso
	re = regexp.MustCompile("iso[-_ ]?([0-9]{4})[-_ ]?([0-9]+|jp)?[-_ ]?(i|e)?")
	matches = re.FindAllStringSubmatch(preparsed, -1)
	if len(matches) == 1 && len(matches[0]) == 4 {
		if matches[0][1] == "2022" && matches[0][2] == "jp" {
			preparsed = "iso-2022-jp"
		}
		if matches[0][1] == "8859" {
			switch matches[0][2] {
			case "1", "2", "3", "4", "5", "7", "8", "9", "10", "11", "13", "14", "15", "16":
				preparsed = "iso-8859-" + matches[0][2]
				if matches[0][3] == "i" {
					preparsed += "-" + matches[0][3]
				}
			case "":
				preparsed = "iso-8859-1"
			}
		}
	}

	// Latin is tricky.
	re = regexp.MustCompile("^(cs|csiso)?l(atin)?[-_ ]?([0-9]{1,2})$")
	matches = re.FindAllStringSubmatch(preparsed, -1)
	if len(matches) == 1 && len(matches[0]) == 4 {
		switch matches[0][3] {
		case "1":
			preparsed = "windows-1252"
		case "2", "3", "4", "5":
			preparsed = "iso-8859-" + matches[0][3]
		case "6":
			preparsed = "iso-8859-10"
		case "8":
			preparsed = "iso-8859-14"
		case "9":
			preparsed = "iso-8859-15"
		case "10":
			preparsed = "iso-8859-16"
		}
	}

	// Missing substitutions.
	switch preparsed {
	case "csutf8", "iso-utf-8", "utf8mb4":
		preparsed = "utf-8"

	case "cp932", "windows-932", "windows-31J", "ibm-943", "cp943":
		preparsed = "shift_jis"
	case "eucjp", "ibm-eucjp":
		preparsed = "euc-jp"
	case "euckr", "ibm-euckr", "cp949":
		preparsed = "euc-kr"
	case "euccn", "ibm-euccn":
		preparsed = "gbk"
	case "zht16mswin950", "cp950":
		preparsed = "big5"

	case "csascii",
		"ansi_x3.4-1968",
		"ansi_x3.4-1986",
		"ansi_x3.110-1983",
		"cp850",
		"cp858",
		"us",
		"iso646",
		"iso-646",
		"iso646-us",
		"iso_646.irv:1991",
		"cp367",
		"ibm367",
		"ibm-367",
		"iso-ir-6":
		preparsed = "ascii"

	case "ibm852":
		preparsed = "iso-8859-2"
	case "iso-ir-199", "iso-celtic":
		preparsed = "iso-8859-14"
	case "iso-ir-226":
		preparsed = "iso-8859-16"

	case "macroman":
		preparsed = "macintosh"
	}

	enc, _ = htmlindex.Get(preparsed)
	if enc == nil {
		err = fmt.Errorf("can not get encoding for '%s' (or '%s')", charset, preparsed)
	}
	return
}

func SelectDecoder(charset string) (decoder *encoding.Decoder, err error) {
	var enc encoding.Encoding
	lcharset := strings.Trim(strings.ToLower(charset), " \t\r\n")
	switch lcharset {
	case "utf7", "utf-7", "unicode-1-1-utf-7":
		return NewUtf7Decoder(), nil
	default:
		enc, err = getEncoding(lcharset)
	}
	if err == nil {
		decoder = enc.NewDecoder()
	}
	return
}

// DecodeHeader if needed. Returns error if raw contains non-utf8 characters.
func DecodeHeader(raw string) (decoded string, err error) {
	if decoded, err = WordDec.DecodeHeader(raw); err != nil {
		decoded = raw
	}
	if !utf8.ValidString(decoded) {
		err = fmt.Errorf("header contains non utf8 chars: %v", err)
	}
	return
}

// EncodeHeader using quoted printable and utf8
func EncodeHeader(s string) string {
	return mime.QEncoding.Encode("utf-8", s)
}

// DecodeCharset decodes the orginal using content type parameters.
// If the charset parameter is missing it checks that the content is valid utf8.
// If it isn't, it checks if it's embedded in the html/xml.
// If it isn't, it falls back to windows-1252.
// It then reencodes it as utf-8.
func DecodeCharset(original []byte, contentType string) ([]byte, error) {
	// If the contentType itself is specified, use that.
	if contentType != "" {
		_, params, err := ParseMediaType(contentType)
		if err != nil {
			return nil, err
		}

		if charset, ok := params["charset"]; ok {
			decoder, err := SelectDecoder(charset)
			if err != nil {
				return original, errors.Wrap(err, "unknown charset was specified")
			}

			return decoder.Bytes(original)
		}
	}

	// The charset was not specified. First try utf8.
	if utf8.Valid(original) {
		return original, nil
	}

	// encoding will be windows-1252 if it can't be determined properly.
	encoding, name, certain := charset.DetermineEncoding(original, contentType)

	if !certain {
		logrus.WithField("encoding", name).Warn("Determined encoding but was not certain")
	}

	// Reencode as UTF-8.
	decoded, err := encoding.NewDecoder().Bytes(original)
	if err != nil {
		return original, errors.Wrap(err, "failed to decode as windows-1252")
	}

	// If the decoded string is not valid utf8, it wasn't windows-1252, so give up.
	if !utf8.Valid(decoded) {
		return original, errors.Wrap(err, "failed to decode as windows-1252")
	}

	return decoded, nil
}

// ParseMediaType from MIME doesn't support RFC2231 for non asci / utf8 encodings so we have to pre-parse it.
func ParseMediaType(v string) (mediatype string, params map[string]string, err error) {
	v, _ = changeEncodingAndKeepLastParamDefinition(v)
	return mime.ParseMediaType(v)
}
