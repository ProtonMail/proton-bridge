// Copyright (c) 2021 Proton Technologies AG
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

package updates

import (
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/ProtonMail/proton-bridge/pkg/dialer"
)

func mkdirAllClear(path string) error {
	if err := os.RemoveAll(path); err != nil {
		return err
	}
	return os.MkdirAll(path, 0750)
}

func downloadToBytes(path string) (out []byte, err error) {
	var (
		client   *http.Client
		response *http.Response
	)
	client = dialer.DialTimeoutClient()
	log.WithField("path", path).Trace("Downloading")

	response, err = client.Get(path)
	if err != nil {
		return
	}
	out, err = ioutil.ReadAll(response.Body)
	_ = response.Body.Close()
	if response.StatusCode < http.StatusOK || http.StatusIMUsed < response.StatusCode {
		err = errors.New(path + " " + response.Status)
	}
	return
}

func downloadWithProgress(status *Progress, sourceURL, targetPath string) (err error) {
	targetFile, err := os.Create(targetPath)
	if err != nil {
		log.Warnf("Cannot create update file %s: %v", targetPath, err)
		return
	}
	defer targetFile.Close() //nolint[errcheck]

	var (
		client   *http.Client
		response *http.Response
	)
	client = dialer.DialTimeoutClient()
	response, err = client.Get(sourceURL)
	if err != nil {
		return
	}
	defer response.Body.Close() //nolint[errcheck]

	contentLength, _ := strconv.ParseUint(response.Header.Get("Content-Length"), 10, 64)

	wc := WriteCounter{
		Status: status,
		Target: targetFile,
		Size:   contentLength,
	}

	err = wc.ReadAll(response.Body)
	return
}

func downloadWithSignature(status *Progress, sourceURL, targetDir string) (localPath string, err error) {
	localPath = filepath.Join(targetDir, filepath.Base(sourceURL))

	if err = downloadWithProgress(nil, sourceURL+sigExtension, localPath+sigExtension); err != nil {
		return
	}

	if err = downloadWithProgress(status, sourceURL, localPath); err != nil {
		return
	}
	return
}

type WriteCounter struct {
	Status                   *Progress
	Target                   io.Writer
	processed, Size, counter uint64
}

func (s *WriteCounter) ReadAll(source io.Reader) (err error) {
	s.counter = uint64(0)
	if s.Target == nil {
		return errors.New("can not read all, target unset")
	}
	if source == nil {
		return errors.New("can not read all, source unset")
	}
	_, err = io.Copy(s.Target, io.TeeReader(source, s))
	return
}

func (s *WriteCounter) Write(p []byte) (int, error) {
	if s.Status != nil && s.Size != 0 {
		s.processed += uint64(len(p))
		fraction := float32(s.processed) / float32(s.Size)
		if s.counter%uint64(100) == 0 || fraction == 1. {
			s.Status.UpdateProcessed(fraction)
		}
	}
	s.counter++
	return len(p), nil
}
