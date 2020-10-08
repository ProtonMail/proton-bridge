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

package transfer

import (
	"bufio"
	"bytes"
	"io"
	"os"
	"strings"
)

const xGmailLabelsHeader = "X-Gmail-Labels"

func getGmailLabelsFromMboxFile(filePath string) ([]string, error) {
	f, err := os.Open(filePath) //nolint[gosec]
	if err != nil {
		return nil, err
	}
	return getGmailLabelsFromMboxReader(f)
}

func getGmailLabelsFromMboxReader(f io.Reader) ([]string, error) {
	allLabels := []string{}

	r := bufio.NewReader(f)
	for {
		b, isPrefix, err := r.ReadLine()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		if isPrefix {
			for !isPrefix {
				_, isPrefix, err = r.ReadLine()
				if err != nil {
					break
				}
			}
			continue
		}
		if bytes.HasPrefix(b, []byte(xGmailLabelsHeader)) {
			for _, label := range getGmailLabelsFromValue(string(b)) {
				allLabels = appendIfNew(allLabels, label)
			}
		}
	}

	return allLabels, nil
}

func getGmailLabelsFromMessage(body []byte) ([]string, error) {
	header, err := getMessageHeader(body)
	if err != nil {
		return nil, err
	}
	labels := header.Get(xGmailLabelsHeader)
	return getGmailLabelsFromValue(labels), nil
}

func getGmailLabelsFromValue(value string) []string {
	value = strings.TrimPrefix(value, xGmailLabelsHeader+":")
	labels := []string{}
	for _, label := range strings.Split(value, ",") {
		label = strings.TrimSpace(label)
		if label == "" {
			continue
		}
		labels = appendIfNew(labels, label)
	}
	return labels
}
