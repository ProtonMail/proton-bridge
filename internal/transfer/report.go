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
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/pkg/errors"
)

// fileReport is struct which can write and read message details.
// File report includes private information.
type fileReport struct {
	path string
}

func openLastFileReport(reportsPath, importID string) (*fileReport, error) { //nolint[deadcode]
	allLogFileNames, err := getFilePathsWithSuffix(reportsPath, ".log")
	if err != nil {
		return nil, err
	}

	reportFileNames := []string{}
	for _, fileName := range allLogFileNames {
		if strings.HasPrefix(fileName, fmt.Sprintf("import_%s_", importID)) {
			reportFileNames = append(reportFileNames, fileName)
		}
	}
	if len(reportFileNames) == 0 {
		return nil, errors.New("no report found")
	}

	sort.Strings(reportFileNames)
	reportFileName := reportFileNames[len(reportFileNames)-1]
	path := filepath.Join(reportsPath, reportFileName)
	return &fileReport{
		path: path,
	}, nil
}

func newFileReport(reportsPath, importID string) *fileReport {
	fileName := fmt.Sprintf("import_%s_%d.log", importID, time.Now().Unix())
	path := filepath.Join(reportsPath, fileName)

	return &fileReport{
		path: path,
	}
}

func (r *fileReport) writeMessageStatus(messageStatus *MessageStatus) {
	f, err := os.OpenFile(r.path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		log.WithError(err).Error("Failed to open report file")
	}
	defer f.Close() //nolint[errcheck]

	messageReport := newMessageReportFromMessageStatus(messageStatus, true)
	data, err := json.Marshal(messageReport)
	if err != nil {
		log.WithError(err).Error("Failed to marshall message details")
	}
	data = append(data, '\n')

	if _, err = f.Write(data); err != nil {
		log.WithError(err).Error("Failed to write to report file")
	}
}

// bugReport is struct which can create report for bug reporting.
// Bug report does NOT include private information.
type bugReport struct {
	data bytes.Buffer
}

func (r *bugReport) writeMessageStatus(messageStatus *MessageStatus) {
	messageReport := newMessageReportFromMessageStatus(messageStatus, false)
	data, err := json.Marshal(messageReport)
	if err != nil {
		log.WithError(err).Error("Failed to marshall message details")
	}
	_, _ = r.data.Write(data)
	_, _ = r.data.Write([]byte("\n"))
}

func (r *bugReport) getData() []byte {
	return r.data.Bytes()
}

// messageReport is struct which holds data used by `fileReport` and `bugReport`.
type messageReport struct {
	EventTime       int64
	SourceID        string
	TargetID        string
	BodyHash        string
	SourceMailbox   string
	TargetMailboxes []string
	Error           string

	// Private information for user.
	Subject string
	From    string
	Time    string
}

func newMessageReportFromMessageStatus(messageStatus *MessageStatus, includePrivateInfo bool) messageReport {
	md := messageReport{
		EventTime:       messageStatus.eventTime.Unix(),
		SourceID:        messageStatus.SourceID,
		TargetID:        messageStatus.targetID,
		BodyHash:        messageStatus.bodyHash,
		SourceMailbox:   messageStatus.rule.SourceMailbox.Name,
		TargetMailboxes: messageStatus.rule.TargetMailboxNames(),
		Error:           messageStatus.GetErrorMessage(),
	}

	if includePrivateInfo {
		md.Subject = messageStatus.Subject
		md.From = messageStatus.From
		md.Time = messageStatus.Time.Format(time.RFC1123Z)
	}

	return md
}
