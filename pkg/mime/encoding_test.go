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

package pmmime

import (
	"bytes"
	"strings"
	"testing"

	"golang.org/x/text/encoding/htmlindex"

	a "github.com/stretchr/testify/assert"
)

func TestDecodeHeader(t *testing.T) {
	testData := []struct{ raw, expected string }{
		{
			"",
			"",
		},
		{
			"=?iso-2022-jp?Q?=1B$B!Z=1B(BTimes_Car_PLUS=1B$B![JV5Q>Z=1B(B?=",
			"【Times Car PLUS】返却証",
		},
		{
			`=?iso-2022-jp?Q?iTunes_Movie_=1B$B%K%e!<%j%j!<%9$HCmL\:nIJ=1B(B?=`,
			"iTunes Movie ニューリリースと注目作品",
		},
		{
			"=?UTF-8?B?w4TDi8OPw5bDnA==?= =?UTF-8?B?IMOkw6vDr8O2w7w=?=",
			"ÄËÏÖÜ äëïöü",
		},
		{
			"=?ISO-8859-2?B?xMtJ1tw=?= =?ISO-8859-2?B?IOTrafb8?=",
			"ÄËIÖÜ äëiöü",
		},
		{
			"=?uknown?B?xMtJ1tw=?= =?ISO-8859-2?B?IOTrafb8?=",
			"=?uknown?B?xMtJ1tw=?= =?ISO-8859-2?B?IOTrafb8?=",
		},
	}

	for _, val := range testData {
		if decoded, err := DecodeHeader(val.raw); strings.Compare(val.expected, decoded) != 0 {
			t.Errorf("Incorrect decoding of header %q expected %q but have %q; Error %v", val.raw, val.expected, decoded, err)
		}
	}
}

type testParseMediaTypeData struct {
	arg, wantMediaType string
	wantParams         map[string]string
}

func (d *testParseMediaTypeData) run(t *testing.T) {
	gotMediaType, params, err := ParseMediaType(d.arg)
	a.Nil(t, err)
	a.Equal(t, d.wantMediaType, gotMediaType)
	a.Equal(t, d.wantParams, params)
}

func TestParseMediaType(t *testing.T) {
	testTable := map[string]testParseMediaTypeData{
		"TwiceTheSameParameter": {
			arg:           "attachment; filename=joy.txt; filename=JOY.TXT; title=hi;",
			wantMediaType: "attachment",
			wantParams:    map[string]string{"filename": "JOY.TXT", "title": "hi"},
		},
		"SingleLineUTF8": {
			arg:           "attachment;\nfilename*=utf-8''%F0%9F%98%81%F0%9F%98%82.txt;\n title=smile",
			wantMediaType: "attachment",
			wantParams:    map[string]string{"filename": "😁😂.txt", "title": "smile"},
		},
		"MultiLineUTF8": {
			arg:           "attachment;\nfilename*0*=utf-8''%F0%9F%98%81;   title=smile;\nfilename*1*=%F0%9F%98%82;\nfilename*2=.txt",
			wantMediaType: "attachment",
			wantParams:    map[string]string{"filename": "😁😂.txt", "title": "smile"},
		},
		"MultiLineFirstNoEncNextUTF8": {
			arg:           "attachment;\nfilename*0*=utf-8''joy  ;\n title*=utf-8''smile;  \nfilename*1*=%F0%9F%98%82;\nfilename*2=.txt",
			wantMediaType: "attachment",
			wantParams:    map[string]string{"filename": "joy😂.txt", "title": "smile"},
		},
		"SingleLineBig5": {
			arg:           "attachment;\nfilename*=big5''%B3%C6%A7%D1%BF%FD.m4a; title*=utf8''memorandum",
			wantMediaType: "attachment",
			wantParams:    map[string]string{"filename": "備忘錄.m4a", "title": "memorandum"},
		},
		"MultiLineBig5": {
			arg:           "attachment;\nfilename*0*=big5''%B3%C6a; title*0=utf8''memorandum; filename*2=%BF%FD.m4a; \nfilename*1*=%A7%D1b;",
			wantMediaType: "attachment",
			wantParams:    map[string]string{"filename": "備a忘b錄.m4a", "title": "memorandum"},
		},
		"SingleLineBadEncoding": {
			arg:           "attachment;\nfilename*=utf-8'%F0%9F%98%81%F0%9F%98%82.txt;\n title=smile",
			wantMediaType: "attachment",
			wantParams:    map[string]string{"title": "smile"},
		},
		"MultiLineBadEncoding": {
			arg:           "attachment;\nfilename*0*=utf-8'%F0%9F%98%81;   title=smile;\nfilename*1*=%F0%9F%98%82;\nfilename*2=.txt",
			wantMediaType: "attachment",
			wantParams:    map[string]string{"filename": "😂.txt", "title": "smile"},
		},
	}
	for name, testData := range testTable {
		t.Run(name, testData.run)
	}
}

func TestGetEncoding(t *testing.T) {
	// All MIME charsets with aliases can be found here:
	// https://www.iana.org/assignments/character-sets/character-sets.xhtml
	mimesets := map[string][]string{
		"utf-8": { // MIB 16
			"utf8",
			"csutf8",
			"unicode-1-1-utf-8",
			"iso-utf-8",
			"utf8mb4",
		},
		"gbk": {
			"gb2312", // MIB 2025
			//"euc-cn": []string{
			"euccn",
			"ibm-euccn",
		},
		//"utf7": []string{"utf-7", "unicode-1-1-utf-7"},
		"iso-8859-2": { // MIB 5
			"iso-ir-101",
			"iso_8859-2",
			"iso8859-2",
			"latin2",
			"l2",
			"csisolatin2",
			"ibm852",
			//"FAILEDibm852",
		},
		"iso-8859-3": { // MIB 6
			"iso-ir-109",
			"iso_8859-3",
			"latin3",
			"l3",
			"csisolatin3",
		},
		"iso-8859-4": { // MIB 7
			"iso-ir-110",
			"iso_8859-4",
			"latin4",
			"l4",
			"csisolatin4",
		},
		"iso-8859-5": { // MIB 8
			"iso-ir-144",
			"iso_8859-5",
			"cyrillic",
			"csisolatincyrillic",
		},
		"iso-8859-6": { // MIB 9
			"iso-ir-127",
			"iso_8859-6",
			"ecma-114",
			"asmo-708",
			"arabic",
			"csisolatinarabic",
			//"iso-8859-6e": []string{ // MIB 81 just direction
			"csiso88596e",
			"iso-8859-6-e",
			//"iso-8859-6i": []string{ // MIB 82
			"csiso88596i",
			"iso-8859-6-i",
		},
		"iso-8859-7": { // MIB 10
			"iso-ir-126",
			"iso_8859-7",
			"elot_928",
			"ecma-118",
			"greek",
			"greek8",
			"csisolatingreek",
		},
		"iso-8859-8": { // MIB 11
			"iso-ir-138",
			"iso_8859-8",
			"hebrew",
			"csisolatinhebrew",
			//"iso-8859-8e": []string{ // MIB 84 (directionality
			"csiso88598e",
			"iso-8859-8-e",
		},
		"iso-8859-8-i": { // MIB 85
			"logical",
			"csiso88598i",
			"iso-8859-8-i", // Hebrew, the "i" means right-to-left, probably unnecessary with ISO cleaning above.
		},
		"iso-8859-10": { // MIB 13
			"iso-ir-157",
			"l6",
			"iso_8859-10:1992",
			"csisolatin6",
			"latin6",
		},
		"iso-8859-13": { // MIB 109
			"csiso885913"},
		"iso-8859-14": { // MIB 110
			"iso-ir-199",
			"iso_8859-14:1998",
			"iso_8859-14",
			"latin8",
			"iso-celtic",
			"l8",
			"csiso885914",
		},
		"iso-8859-15": { // MIB 111
			"iso_8859-15",
			"latin-9",
			"csiso885915",
			"ISO8859-15",
		},
		"iso-8859-16": { // MIB 112
			"iso-ir-226",
			"iso_8859-16:2001",
			"iso_8859-16",
			"latin10",
			"l10",
			"csiso885916",
		},
		"windows-874": { // MIB 2109
			"cswindows874",
			"cp874",
			"iso-8859-11",
			"tis-620",
		},
		"windows-1250": { // MIB 2250
			"cswindows1250",
			"cp1250",
		},
		"windows-1251": { // MIB 2251
			"cswindows1251",
			"cp1251",
		},
		"windows-1252": { // MIB 2252
			"cswindows1252",
			"cp1252",
			"3dwindows-1252",
			"we8mswin1252",
			"us-ascii",         // MIB 3
			"ansi_x3.110-1983", // MIB 74 // usascii
			//"iso-8859-1": []string{ // MIB 4 succeed by win1252
			"iso8859-1",
			"iso-ir-100",
			"iso_8859-1",
			"latin1",
			"l1",
			"ibm819",
			"cp819",
			"csisolatin1",
			"ansi_x3.4-1968",
			"ansi_x3.4-1986",
			"cp850",
			"cp858", // "cp850"  Mostly correct except for the Euro sign.
			"iso_646.irv:1991",
			"iso646-us",
			"us",
			"ibm367",
			"cp367",
			"csascii",
			"ascii",
			"iso-ir-6",
			"we8iso8859p1",
		},
		"windows-1253": {"cswindows1253", "cp1253"},        // MIB 2253
		"windows-1254": {"cswindows1254", "cp1254"},        // MIB 2254
		"windows-1255": {"cSwindows1255", "cp1255"},        // MIB 2255
		"windows-1256": {"cswIndows1256", "cp1256"},        // MIB 2256
		"windows-1257": {"cswinDows1257", "cp1257"},        // MIB 2257
		"windows-1258": {"cswindoWs1258", "cp1258"},        // MIB 2257
		"koi8-r":       {"cskoi8r", "koi8r"},               // MIB 2084
		"koi8-u":       {"cskoi8u", "koi8u"},               // MIB 2088
		"macintosh":    {"mac", "macroman", "csmacintosh"}, // MIB 2027
		"big5": {
			"zht16mswin950", // cp950
			"cp950",
		},
		"euc-kr": {
			"euckr", // MIB 38
			"ibm-euckr",
			//"uhc": []string{ // Korea
			"ks_c_5601-1987",
			"ksc5601",
			"cp949",
		},
		"euc-jp": {
			"eucjp",
			"ibm-eucjp",
		},
		"shift_jis": {
			"CP932",
			"MS932",
			"Windows-932",
			"Windows-31J",
			"MS_Kanji",
			"IBM-943",
			"CP943",
		},
		"iso-2022-jp": { // MIB 39
			"iso2022jp",
			"csiso2022jp",
		},
	}

	for expected, names := range mimesets {
		expenc, _ := htmlindex.Get(expected)
		if canonical, err := htmlindex.Name(expenc); canonical != expected || err != nil {
			t.Fatalf("Error while get canonical name. Expected '%v' but have %v `%#v`: %v", expected, canonical, expenc, err)
		}
		for _, name := range names {
			enc, err := getEncoding(name)
			if err != nil || enc == nil {
				t.Errorf("Error while getting encoding for %v returned: '%#v' and error: '%v'", name, enc, err)
			}
			if expenc != enc {
				t.Errorf("For %v expected %v '%v' but have '%v'", name, expected, expenc, enc)
			}
		}
	}
}

// sample text for UTF8 http://www.columbia.edu/~fdc/utf8/index.html
func TestEncodeReader(t *testing.T) {
	// define test data
	testData := []struct {
		charset  string
		original []byte
		message  string
	}{
		// russian
		{
			"koi8-r",
			//     а, з, б, у, к, а, а, б, в, г, д, е, ё
			[]byte{0xC1, 0xDA, 0xC2, 0xD5, 0xCB, 0xC1, 0xC1, 0xC2, 0xD7, 0xC7, 0xC4, 0xC5, 0xA3},
			"азбукаабвгдеё",
		},
		{
			"KOI8-R",
			[]byte{0xC1, 0xDA, 0xC2, 0xD5, 0xCB, 0xC1, 0xC1, 0xC2, 0xD7, 0xC7, 0xC4, 0xC5, 0xA3},
			"азбукаабвгдеё",
		},
		{
			"csKOI8R",
			[]byte{0xC1, 0xDA, 0xC2, 0xD5, 0xCB, 0xC1, 0xC1, 0xC2, 0xD7, 0xC7, 0xC4, 0xC5, 0xA3},
			"азбукаабвгдеё",
		},
		{
			"koi8-u",
			[]byte{0xC1, 0xDA, 0xC2, 0xD5, 0xCB, 0xC1, 0xC1, 0xC2, 0xD7, 0xC7, 0xC4, 0xC5, 0xA3},
			"азбукаабвгдеё",
		},
		{
			"iso-8859-5",
			//     а    , з    , б    , у    , к    , а    , а    , б    , в    , г    , д    , е    , ё
			[]byte{0xD0, 0xD7, 0xD1, 0xE3, 0xDA, 0xD0, 0xD0, 0xD1, 0xD2, 0xD3, 0xD4, 0xD5, 0xF1},
			"азбукаабвгдеё",
		},
		{
			"csWrong",
			[]byte{0xD0, 0xD7, 0xD1, 0xE3, 0xDA, 0xD0, 0xD0, 0xD1, 0xD2, 0xD3, 0xD4, 0xD5, 0xD6},
			"",
		},
		{
			"utf8",
			[]byte{0xD0, 0xB0, 0xD0, 0xB7, 0xD0, 0xB1, 0xD1, 0x83, 0xD0, 0xBA, 0xD0, 0xB0, 0xD0, 0xB0, 0xD0, 0xB1, 0xD0, 0xB2, 0xD0, 0xB3, 0xD0, 0xB4, 0xD0, 0xB5, 0xD1, 0x91},
			"азбукаабвгдеё",
		},
		// czechoslovakia
		{
			"windows-1250",
			[]byte{225, 228, 232, 233, 236, 244},
			"áäčéěô",
		},
		// umlauts
		{
			"iso-8859-1",
			[]byte{196, 203, 214, 220, 228, 235, 246, 252},
			"ÄËÖÜäëöü",
		},
		// latvia
		{
			"iso-8859-4",
			[]byte{224, 239, 243, 182, 254},
			"āīķļū",
		},
		{ // encoded by https://www.motobit.com/util/charset-codepage-conversion.asp
			"utf7",
			[]byte("He wes Leovena+APA-es sone -- li+APA-e him be Drihten.+A6QDtw- +A7MDuwPOA8MDwwOx- +A7wDvwPF- +A60DtAPJA8MDsQO9- +A7UDuwO7A7cDvQO5A7oDrg-. +BCcENQRABD0ENQQ7BDg- +BDgENwQxBEs- +BDcENAQ1BEEETA- +BDg- +BEIEMAQ8-,+BCcENQRABD0ENQQ7BDg- +BDgENwQxBEs- +BDcENAQ1BEEETA- +BDg- +BEIEMAQ8-,+C68LvguuC7ELvwuoC80LpA- +C64Lygu0C78LlQuzC78LsgvH- +C6QLrgu/C7QLzQuuC8oLtAu/- +C6oLywuyC80- +C4cLqQu/C6QLvgu1C6QLwQ- +C44LmQvNC5ULwQuuC80- +C5ULvgujC8sLrgvN-."),
			"He wes Leovenaðes sone -- liðe him be Drihten.Τη γλώσσα μου έδωσαν ελληνική. Чернели избы здесь и там,Чернели избы здесь и там,யாமறிந்த மொழிகளிலே தமிழ்மொழி போல் இனிதாவது எங்கும் காணோம்.",
		},

		// iconv -f UTF8 -t GB2312 utf8.txt | hexdump -v -e '"0x" 1/1 "%x, "'
		{ // encoded by iconv; dump by `cat gb2312.txt | hexdump -v -e '"0x" 1/1 "%x "'` and reformat; text from https://zh.wikipedia.org/wiki/GB_2312
			"GB2312",
			[]byte{0x47, 0x42, 0x20, 0x32, 0x33, 0x31, 0x32, 0xb5, 0xc4, 0xb3, 0xf6, 0xcf, 0xd6, 0xa3, 0xac, 0xbb, 0xf9, 0xb1, 0xbe, 0xc2, 0xfa, 0xd7, 0xe3, 0xc1, 0xcb, 0xba, 0xba, 0xd7, 0xd6, 0xb5, 0xc4, 0xbc, 0xc6, 0xcb, 0xe3, 0xbb, 0xfa, 0xb4, 0xa6, 0xc0, 0xed, 0xd0, 0xe8, 0xd2, 0xaa, 0xa3, 0xac, 0xcb, 0xfc, 0xcb, 0xf9, 0xca, 0xd5, 0xc2, 0xbc, 0xb5, 0xc4, 0xba, 0xba, 0xd7, 0xd6, 0xd2, 0xd1, 0xbe, 0xad, 0xb8, 0xb2, 0xb8, 0xc7, 0xd6, 0xd0, 0xb9, 0xfa, 0xb4, 0xf3, 0xc2, 0xbd, 0x39, 0x39, 0x2e, 0x37, 0x35, 0x25, 0xb5, 0xc4, 0xca, 0xb9, 0xd3, 0xc3, 0xc6, 0xb5, 0xc2, 0xca, 0xa1, 0xa3, 0xb5, 0xab, 0xb6, 0xd4, 0xd3, 0xda, 0xc8, 0xcb, 0xc3, 0xfb},
			"GB 2312的出现，基本满足了汉字的计算机处理需要，它所收录的汉字已经覆盖中国大陆99.75%的使用频率。但对于人名",
		},

		{ // encoded by iconv; text from https://jp.wikipedia.org/wiki/Shift_JIS
			"shift-jis",
			[]byte{0x95, 0xb6, 0x8e, 0x9a, 0x95, 0x84, 0x8d, 0x86, 0x89, 0xbb, 0x95, 0xfb, 0x8e, 0xae, 0x53, 0x68, 0x69, 0x66, 0x74, 0x5f, 0x4a, 0x49, 0x53, 0x82, 0xcc, 0x90, 0xdd, 0x8c, 0x76, 0x8e, 0xd2, 0x82, 0xe7, 0x82, 0xcd, 0x81, 0x41, 0x90, 0xe6, 0x8d, 0x73, 0x82, 0xb5, 0x82, 0xc4, 0x82, 0xe6, 0x82, 0xad, 0x97, 0x98, 0x97, 0x70, 0x82, 0xb3, 0x82, 0xea, 0x82, 0xc4, 0x82, 0xa2, 0x82, 0xbd, 0x4a, 0x49, 0x53, 0x20, 0x43, 0x20, 0x36, 0x32, 0x32, 0x30, 0x81, 0x69, 0x8c, 0xbb, 0x8d, 0xdd, 0x82, 0xcc, 0x4a, 0x49, 0x53, 0x20, 0x58, 0x20, 0x30, 0x32, 0x30, 0x31, 0x81, 0x6a, 0x82, 0xcc, 0x38, 0x83, 0x72, 0x83, 0x62, 0x83, 0x67, 0x95, 0x84, 0x8d, 0x86, 0x81, 0x69, 0x88, 0xc8, 0x89, 0xba, 0x81, 0x75, 0x89, 0x70, 0x90, 0x94, 0x8e, 0x9a, 0x81, 0x45, 0x94, 0xbc, 0x8a, 0x70, 0x83, 0x4a, 0x83, 0x69, 0x81, 0x76, 0x81, 0x6a, 0x82, 0xc6, 0x81, 0x41, 0x4a, 0x49, 0x53, 0x20, 0x43, 0x20, 0x36, 0x32, 0x32, 0x36, 0x81, 0x69, 0x8c, 0xbb, 0x8d, 0xdd, 0x82, 0xcc, 0x4a, 0x49, 0x53, 0x20, 0x58, 0x20, 0x30, 0x32, 0x30, 0x38, 0x81, 0x41, 0x88, 0xc8, 0x89, 0xba, 0x81, 0x75, 0x8a, 0xbf, 0x8e, 0x9a, 0x81, 0x76, 0x81, 0x6a, 0x82, 0xcc, 0x97, 0xbc, 0x95, 0xb6, 0x8e, 0x9a, 0x8f, 0x57, 0x8d, 0x87, 0x82, 0xf0, 0x95, 0x5c, 0x8c, 0xbb, 0x82, 0xb5, 0x82, 0xe6, 0x82, 0xa4, 0x82, 0xc6, 0x82, 0xb5, 0x82, 0xbd, 0x81, 0x42, 0x82, 0xdc, 0x82, 0xbd, 0x81, 0x41, 0x83, 0x74, 0x83, 0x40, 0x83, 0x43, 0x83, 0x8b, 0x82, 0xcc, 0x91, 0xe5, 0x82, 0xab, 0x82, 0xb3, 0x82, 0xe2, 0x8f, 0x88, 0x97, 0x9d, 0x8e, 0x9e, 0x8a, 0xd4, 0x82, 0xcc, 0x92, 0x5a, 0x8f, 0x6b, 0x82, 0xf0, 0x90, 0x7d, 0x82, 0xe9, 0x82, 0xbd, 0x82, 0xdf, 0x81, 0x41, 0x83, 0x47, 0x83, 0x58, 0x83, 0x50, 0x81, 0x5b, 0x83, 0x76, 0x83, 0x56, 0x81, 0x5b, 0x83, 0x50, 0x83, 0x93, 0x83, 0x58, 0x82, 0xc8, 0x82, 0xb5, 0x82, 0xc5, 0x8d, 0xac, 0x8d, 0xdd, 0x89, 0xc2, 0x94, 0x5c, 0x82, 0xc9, 0x82, 0xb7, 0x82, 0xe9, 0x82, 0xb1, 0x82, 0xc6, 0x82, 0xf0, 0x8a, 0xe9, 0x90, 0x7d, 0x82, 0xb5, 0x82, 0xbd, 0x81, 0x42},
			"文字符号化方式Shift_JISの設計者らは、先行してよく利用されていたJIS C 6220（現在のJIS X 0201）の8ビット符号（以下「英数字・半角カナ」）と、JIS C 6226（現在のJIS X 0208、以下「漢字」）の両文字集合を表現しようとした。また、ファイルの大きさや処理時間の短縮を図るため、エスケープシーケンスなしで混在可能にすることを企図した。",
		},

		// add more from mutations of https://en.wikipedia.org/wiki/World_Wide_Web

	}

	// run tests
	for _, val := range testData {
		// fmt.Println("Testing ", val)
		expected := []byte(val.message)
		decoded, err := DecodeCharset(val.original, "text/plain; charset="+val.charset)
		if len(expected) == 0 {
			if err == nil {
				t.Error("Expected err but have ", err)
			} else {
				// fmt.Println("Expected err: ", err)
				continue
			}
		} else {
			if err != nil {
				t.Error("Expected ok but have ", err)
			}
		}

		if bytes.Equal(decoded, expected) {
			// fmt.Println("Succesfull decoding of ", val.params, ":", string(decoded))
		} else {
			t.Error("Wrong encoding of ", val.charset, ".Expected\n", expected, "\nbut have\n", decoded)
		}
		if strings.Compare(val.message, string(decoded)) != 0 {
			t.Error("Wrong message for ", val.charset, ".Expected\n", val.message, "\nbut have\n", string(decoded))
		}
	}
}
