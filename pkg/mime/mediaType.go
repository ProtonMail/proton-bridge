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

package pmmime

import (
	"errors"
	"fmt"
	"strings"
	"unicode"

	"github.com/sirupsen/logrus"
)

// changeEncodingAndKeepLastParamDefinition is necessary to modify behaviour
// provided by the golang standard libraries.
func changeEncodingAndKeepLastParamDefinition(v string) (out string, err error) {
	log := logrus.WithField("pkg", "pm-mime")

	out = v // By default don't do anything with that.
	keepOrig := true

	i := strings.Index(v, ";")
	if i == -1 {
		i = len(v)
	}
	mediatype := strings.TrimSpace(strings.ToLower(v[0:i]))

	params := map[string]string{}
	var continuation map[string]map[string]string

	v = v[i:]
	for len(v) > 0 {
		v = strings.TrimLeftFunc(v, unicode.IsSpace)
		if len(v) == 0 {
			break
		}
		key, value, rest := consumeMediaParam(v)
		if key == "" {
			break
		}

		pmap := params
		if idx := strings.Index(key, "*"); idx != -1 {
			baseName := key[:idx]
			if continuation == nil {
				continuation = make(map[string]map[string]string)
			}
			var ok bool
			if pmap, ok = continuation[baseName]; !ok {
				continuation[baseName] = make(map[string]string)
				pmap = continuation[baseName]
			}
			if isFirstContinuation(key) {
				charset, _, err := get2231Charset(value)
				if err != nil {
					log.Errorln("Filter params:", err)
					v = rest
					continue
				}
				if charset != "utf-8" && charset != "us-ascii" {
					keepOrig = false
				}
			}
		}
		if _, exists := pmap[key]; exists {
			keepOrig = false
		}
		pmap[key] = value
		v = rest
	}

	if keepOrig {
		return
	}

	for paramKey, contMap := range continuation {
		value, err := mergeContinuations(paramKey, contMap)
		if err == nil {
			params[paramKey+"*"] = value
			continue
		}

		// Fallback.
		log.Errorln("Merge param", paramKey, ":", err)
		for ck, cv := range contMap {
			params[ck] = cv
		}
	}

	// Merge ;
	out = mediatype
	for k, v := range params {
		out += ";"
		out += k
		out += "="
		out += v
	}

	return
}

func isFirstContinuation(key string) bool {
	if idx := strings.Index(key, "*"); idx != -1 {
		return key[idx:] == "*" || key[idx:] == "*0*"
	}
	return false
}

// get2231Charset partially from mime/mediatype.go:211 function `decode2231Enc`.
func get2231Charset(v string) (charset, value string, err error) {
	sv := strings.SplitN(v, "'", 3)
	if len(sv) != 3 {
		err = errors.New("incorrect RFC2231 charset format")
		return
	}
	charset = strings.ToLower(sv[0])
	value = sv[2]
	return
}

func mergeContinuations(paramKey string, contMap map[string]string) (string, error) {
	var err error
	var charset, value string

	// Single value.
	if contValue, ok := contMap[paramKey+"*"]; ok {
		if charset, value, err = get2231Charset(contValue); err != nil {
			return "", err
		}
	} else {
		for n := 0; ; n++ {
			contKey := fmt.Sprintf("%s*%d", paramKey, n)
			contValue, isLast := contMap[contKey]
			if !isLast {
				var ok bool
				contValue, ok = contMap[contKey+"*"]
				if !ok {
					return "", errors.New("not valid RFC2231 continuation")
				}
			}
			if n == 0 {
				if charset, value, err = get2231Charset(contValue); err != nil || charset == "" {
					return "", err
				}
			} else {
				value += contValue
			}
			if isLast {
				break
			}
		}
	}

	return convertHexToUTF(charset, value)
}

// convertHexToUTF converts hex values string with charset to UTF8 in RFC2231 format.
func convertHexToUTF(charset, value string) (string, error) {
	raw, err := percentHexUnescape(value)
	if err != nil {
		return "", err
	}
	utf8, err := DecodeCharset(raw, "text/plain; charset="+charset)
	return "utf-8''" + percentHexEscape(utf8), err
}

// consumeMediaParam copy paste mime/mediatype.go:297.
func consumeMediaParam(v string) (param, value, rest string) {
	rest = strings.TrimLeftFunc(v, unicode.IsSpace)
	if !strings.HasPrefix(rest, ";") {
		return "", "", v
	}

	rest = rest[1:] // Consume semicolon.
	rest = strings.TrimLeftFunc(rest, unicode.IsSpace)
	param, rest = consumeToken(rest)
	param = strings.ToLower(param)
	if param == "" {
		return "", "", v
	}

	rest = strings.TrimLeftFunc(rest, unicode.IsSpace)
	if !strings.HasPrefix(rest, "=") {
		return "", "", v
	}
	rest = rest[1:] // Consume equals sign.
	rest = strings.TrimLeftFunc(rest, unicode.IsSpace)
	value, rest2 := consumeValue(rest)
	if value == "" && rest2 == rest {
		return "", "", v
	}
	rest = rest2
	return param, value, rest
}

// consumeToken copy paste mime/mediatype.go:238.
// consumeToken consumes a token from the beginning of the provided string,
// per RFC 2045 section 5.1 (referenced from 2183), and returns
// the token consumed and the rest of the string.
// Returns ("", v) on failure to consume at least one character.
func consumeToken(v string) (token, rest string) {
	notPos := strings.IndexFunc(v, isNotTokenChar)
	if notPos == -1 {
		return v, ""
	}
	if notPos == 0 {
		return "", v
	}
	return v[0:notPos], v[notPos:]
}

// consumeValue copy paste mime/mediatype.go:253
// consumeValue consumes a "value" per RFC 2045, where a value is
// either a 'token' or a 'quoted-string'.  On success, consumeValue
// returns the value consumed (and de-quoted/escaped, if a
// quoted-string) and the rest of the string.
// On failure, returns ("", v).
func consumeValue(v string) (value, rest string) {
	if v == "" {
		return
	}
	if v[0] != '"' {
		return consumeToken(v)
	}

	// parse a quoted-string
	buffer := new(strings.Builder)
	for i := 1; i < len(v); i++ {
		r := v[i]
		if r == '"' {
			return buffer.String(), v[i+1:]
		}
		// When MSIE sends a full file path (in "intranet mode"), it does not
		// escape backslashes: "C:\dev\go\foo.txt", not "C:\\dev\\go\\foo.txt".
		//
		// No known MIME generators emit unnecessary backslash escapes
		// for simple token characters like numbers and letters.
		//
		// If we see an unnecessary backslash escape, assume it is from MSIE
		// and intended as a literal backslash. This makes Go servers deal better
		// with MSIE without affecting the way they handle conforming MIME
		// generators.
		if r == '\\' && i+1 < len(v) && !isTokenChar(rune(v[i+1])) {
			buffer.WriteByte(v[i+1])
			i++
			continue
		}
		if r == '\r' || r == '\n' {
			return "", v
		}
		buffer.WriteByte(v[i])
	}
	// Did not find end quote.
	return "", v
}

// isNotTokenChar copy paste from mime/mediatype.go:234.
func isNotTokenChar(r rune) bool {
	return !isTokenChar(r)
}

// isTokenChar copy paste from mime/grammar.go:19.
// isTokenChar reports whether rune is in 'token' as defined by RFC 1521 and RFC 2045.
func isTokenChar(r rune) bool {
	// token := 1*<any (US-ASCII) CHAR except SPACE, CTLs,
	//             or tspecials>
	return r > 0x20 && r < 0x7f && !isTSpecial(r)
}

// isTSpecial copy paste from mime/grammar.go:13
// isTSpecial reports whether rune is in 'tspecials' as defined by RFC
// 1521 and RFC 2045.
func isTSpecial(r rune) bool {
	return strings.ContainsRune(`()<>@,;:\"/[]?=`, r)
}

func percentHexEscape(raw []byte) (out string) {
	for _, v := range raw {
		out += fmt.Sprintf("%%%x", v)
	}
	return
}

// percentHexUnescape copy paste from mime/mediatype.go:325.
func percentHexUnescape(s string) ([]byte, error) {
	// Count %, check that they're well-formed.
	percents := 0
	for i := 0; i < len(s); {
		if s[i] != '%' {
			i++
			continue
		}
		percents++
		if i+2 >= len(s) || !ishex(s[i+1]) || !ishex(s[i+2]) {
			s = s[i:]
			if len(s) > 3 {
				s = s[0:3]
			}
			return []byte{}, fmt.Errorf("mime: bogus characters after %%: %q", s)
		}
		i += 3
	}
	if percents == 0 {
		return []byte(s), nil
	}

	t := make([]byte, len(s)-2*percents)
	j := 0
	for i := 0; i < len(s); {
		switch s[i] {
		case '%':
			t[j] = unhex(s[i+1])<<4 | unhex(s[i+2])
			j++
			i += 3
		default:
			t[j] = s[i]
			j++
			i++
		}
	}
	return t, nil
}

// ishex copy paste from mime/mediatype.go:364.
func ishex(c byte) bool {
	switch {
	case '0' <= c && c <= '9':
		return true
	case 'a' <= c && c <= 'f':
		return true
	case 'A' <= c && c <= 'F':
		return true
	}
	return false
}

// unhex copy paste from mime/mediatype.go:376.
func unhex(c byte) byte {
	switch {
	case '0' <= c && c <= '9':
		return c - '0'
	case 'a' <= c && c <= 'f':
		return c - 'a' + 10
	case 'A' <= c && c <= 'F':
		return c - 'A' + 10
	}
	return 0
}
