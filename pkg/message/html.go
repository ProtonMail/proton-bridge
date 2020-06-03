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
	"errors"
	escape "html"
	"strings"

	"github.com/andybalholm/cascadia"
	"golang.org/x/net/html"
)

func plaintextToHTML(text string) (output string) {
	text = escape.EscapeString(text)
	text = strings.Replace(text, "\n\r", "<br>", -1)
	text = strings.Replace(text, "\r\n", "<br>", -1)
	text = strings.Replace(text, "\n", "<br>", -1)
	text = strings.Replace(text, "\r", "<br>", -1)

	return "<div>" + text + "</div>"
}

func stripHTML(input string) (stripped string, err error) {
	reader := strings.NewReader(input)
	doc, _ := html.Parse(reader)
	body := cascadia.MustCompile("body").MatchFirst(doc)
	if body == nil {
		err = errors.New("failed to find necessary html element")
		return
	}
	var buf1 bytes.Buffer
	if err = html.Render(&buf1, body); err != nil {
		stripped = input
		return
	}
	stripped = buf1.String()
	// Handle double body tags edge case.
	if strings.Index(stripped, "<body") == 0 {
		startIndex := strings.Index(stripped, ">")
		if startIndex < 5 {
			return
		}
		stripped = stripped[startIndex+1:]
		// Closing body tag is optional.
		closingIndex := strings.Index(stripped, "</body>")
		if closingIndex > -1 {
			stripped = stripped[:closingIndex]
		}
	}
	return
}

func addOuterHTMLTags(input string) (output string) {
	return "<html><head></head><body>" + input + "</body></html>"
}

func makeEmbeddedImageHTML(cid, name string) (output string) {
	return "<img class=\"proton-embedded\" alt=\"" + name + "\" src=\"cid:" + cid + "\">"
}
