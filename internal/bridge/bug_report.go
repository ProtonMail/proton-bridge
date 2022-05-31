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

package bridge

import (
	"archive/zip"
	"bytes"
	"context"
	"errors"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"

	"github.com/ProtonMail/proton-bridge/v2/internal/logging"
	"github.com/ProtonMail/proton-bridge/v2/pkg/pmapi"
)

const (
	MaxAttachmentSize       = 7 * 1024 * 1024 // 7 MB total limit
	MaxCompressedFilesCount = 6
)

var ErrSizeTooLarge = errors.New("file is too big")

// ReportBug reports a new bug from the user.
func (b *Bridge) ReportBug(osType, osVersion, description, accountName, address, emailClient string, attachLogs bool) error {
	if user, err := b.GetUser(address); err == nil {
		accountName = user.Username()
	} else if users := b.GetUsers(); len(users) > 0 {
		accountName = users[0].Username()
	}

	report := pmapi.ReportBugReq{
		OS:          osType,
		OSVersion:   osVersion,
		Browser:     emailClient,
		Title:       "[Bridge] Bug",
		Description: description,
		Username:    accountName,
		Email:       address,
	}

	if attachLogs {
		logs, err := b.getMatchingLogs(
			func(filename string) bool {
				return logging.MatchLogName(filename) && !logging.MatchStackTraceName(filename)
			},
		)
		if err != nil {
			log.WithError(err).Error("Can't get log files list")
		}
		crashes, err := b.getMatchingLogs(
			func(filename string) bool {
				return logging.MatchLogName(filename) && logging.MatchStackTraceName(filename)
			},
		)
		if err != nil {
			log.WithError(err).Error("Can't get crash files list")
		}

		var matchFiles []string

		matchFiles = append(matchFiles, logs[max(0, len(logs)-(MaxCompressedFilesCount/2)):]...)
		matchFiles = append(matchFiles, crashes[max(0, len(crashes)-(MaxCompressedFilesCount/2)):]...)

		archive, err := zipFiles(matchFiles)
		if err != nil {
			log.WithError(err).Error("Can't zip logs and crashes")
		}

		if archive != nil {
			report.AddAttachment("logs.zip", "application/zip", archive)
		}
	}

	return b.clientManager.ReportBug(context.Background(), report)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func (b *Bridge) getMatchingLogs(filenameMatchFunc func(string) bool) (filenames []string, err error) {
	logsPath, err := b.locations.ProvideLogsPath()
	if err != nil {
		return nil, err
	}

	files, err := ioutil.ReadDir(logsPath)
	if err != nil {
		return nil, err
	}

	var matchFiles []string

	for _, file := range files {
		if filenameMatchFunc(file.Name()) {
			matchFiles = append(matchFiles, filepath.Join(logsPath, file.Name()))
		}
	}
	sort.Strings(matchFiles) // Sorted by timestamp: oldest first.

	return matchFiles, nil
}

type LimitedBuffer struct {
	capacity int
	buf      *bytes.Buffer
}

func NewLimitedBuffer(capacity int) *LimitedBuffer {
	return &LimitedBuffer{
		capacity: capacity,
		buf:      bytes.NewBuffer(make([]byte, 0, capacity)),
	}
}

func (b *LimitedBuffer) Write(p []byte) (n int, err error) {
	if len(p)+b.buf.Len() > b.capacity {
		return 0, ErrSizeTooLarge
	}

	return b.buf.Write(p)
}

func (b *LimitedBuffer) Read(p []byte) (n int, err error) {
	return b.buf.Read(p)
}

func zipFiles(filenames []string) (io.Reader, error) {
	if len(filenames) == 0 {
		return nil, nil
	}

	buf := NewLimitedBuffer(MaxAttachmentSize)

	w := zip.NewWriter(buf)
	defer w.Close() //nolint:errcheck

	for _, file := range filenames {
		err := addFileToZip(file, w)
		if err != nil {
			return nil, err
		}
	}

	if err := w.Close(); err != nil {
		return nil, err
	}

	return buf, nil
}

func addFileToZip(filename string, writer *zip.Writer) error {
	fileReader, err := os.Open(filepath.Clean(filename))
	if err != nil {
		return err
	}
	defer fileReader.Close() //nolint:errcheck,gosec

	fileInfo, err := fileReader.Stat()
	if err != nil {
		return err
	}

	header, err := zip.FileInfoHeader(fileInfo)
	if err != nil {
		return err
	}

	header.Method = zip.Deflate
	header.Name = filepath.Base(filename)

	fileWriter, err := writer.CreateHeader(header)
	if err != nil {
		return err
	}

	_, err = io.Copy(fileWriter, fileReader)
	if err != nil {
		return err
	}

	err = fileReader.Close()

	return err
}
