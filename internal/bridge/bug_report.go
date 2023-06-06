// Copyright (c) 2023 Proton AG
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
	"io"
	"os"
	"path/filepath"
	"sort"

	"github.com/ProtonMail/go-proton-api"
	"github.com/ProtonMail/proton-bridge/v3/internal/constants"
	"github.com/ProtonMail/proton-bridge/v3/internal/logging"
	"github.com/ProtonMail/proton-bridge/v3/internal/vault"
)

const (
	MaxTotalAttachmentSize  = 7 * (1 << 20)
	MaxCompressedFilesCount = 6
)

func (bridge *Bridge) ReportBug(ctx context.Context, osType, osVersion, description, username, email, client string, attachLogs bool) error {
	var account string

	if info, err := bridge.QueryUserInfo(username); err == nil {
		account = info.Username
	} else if userIDs := bridge.GetUserIDs(); len(userIDs) > 0 {
		if err := bridge.vault.GetUser(userIDs[0], func(user *vault.User) {
			account = user.Username()
		}); err != nil {
			return err
		}
	}

	var atts []proton.ReportBugAttachment

	if attachLogs {
		logs, err := getMatchingLogs(bridge.locator, func(filename string) bool {
			return logging.MatchBridgeLogName(filename) && !logging.MatchStackTraceName(filename)
		})
		if err != nil {
			return err
		}

		crashes, err := getMatchingLogs(bridge.locator, func(filename string) bool {
			return logging.MatchBridgeLogName(filename) && logging.MatchStackTraceName(filename)
		})
		if err != nil {
			return err
		}

		guiLogs, err := getMatchingLogs(bridge.locator, func(filename string) bool {
			return logging.MatchGUILogName(filename) && !logging.MatchStackTraceName(filename)
		})
		if err != nil {
			return err
		}

		var matchFiles []string

		// Include bridge logs, up to a maximum amount.
		matchFiles = append(matchFiles, logs[max(0, len(logs)-(MaxCompressedFilesCount/2)):]...)

		// Include crash logs, up to a maximum amount.
		matchFiles = append(matchFiles, crashes[max(0, len(crashes)-(MaxCompressedFilesCount/2)):]...)

		// bridge-gui keeps just one small (~ 1kb) log file; we always include it.
		if len(guiLogs) > 0 {
			matchFiles = append(matchFiles, guiLogs[len(guiLogs)-1])
		}

		archive, err := zipFiles(matchFiles)
		if err != nil {
			return err
		}

		body, err := io.ReadAll(archive)
		if err != nil {
			return err
		}

		atts = append(atts, proton.ReportBugAttachment{
			Name:     "logs.zip",
			Filename: "logs.zip",
			MIMEType: "application/zip",
			Body:     body,
		})
	}

	return bridge.api.ReportBug(ctx, proton.ReportBugReq{
		OS:        osType,
		OSVersion: osVersion,

		Title:       "[Bridge] Bug",
		Description: description,

		Client:        client,
		ClientType:    proton.ClientTypeEmail,
		ClientVersion: constants.AppVersion(bridge.curVersion.Original()),

		Username: account,
		Email:    email,
	}, atts...)
}

func max(a, b int) int {
	if a > b {
		return a
	}

	return b
}

func getMatchingLogs(locator Locator, filenameMatchFunc func(string) bool) (filenames []string, err error) {
	logsPath, err := locator.ProvideLogsPath()
	if err != nil {
		return nil, err
	}

	files, err := os.ReadDir(logsPath)
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

type limitedBuffer struct {
	capacity int
	buf      *bytes.Buffer
}

func newLimitedBuffer(capacity int) *limitedBuffer {
	return &limitedBuffer{
		capacity: capacity,
		buf:      bytes.NewBuffer(make([]byte, 0, capacity)),
	}
}

func (b *limitedBuffer) Write(p []byte) (n int, err error) {
	if len(p)+b.buf.Len() > b.capacity {
		return 0, ErrSizeTooLarge
	}

	return b.buf.Write(p)
}

func (b *limitedBuffer) Read(p []byte) (n int, err error) {
	return b.buf.Read(p)
}

func zipFiles(filenames []string) (io.Reader, error) {
	if len(filenames) == 0 {
		return nil, nil
	}

	buf := newLimitedBuffer(MaxTotalAttachmentSize)

	w := zip.NewWriter(buf)
	defer w.Close() //nolint:errcheck

	for _, file := range filenames {
		if err := addFileToZip(file, w); err != nil {
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

	if _, err := io.Copy(fileWriter, fileReader); err != nil {
		return err
	}

	return fileReader.Close()
}
