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

package pmapi

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
)

// NewRequest creates a new request.
func (c *client) NewRequest(method, path string, body io.Reader) (req *http.Request, err error) {
	req, err = http.NewRequest(method, c.cm.GetRootURL()+path, body)

	if req != nil {
		req.Header.Set("User-Agent", CurrentUserAgent)
	}
	return
}

// NewJSONRequest create a new JSON request.
func (c *client) NewJSONRequest(method, path string, body interface{}) (*http.Request, error) {
	b, err := json.Marshal(body)
	if err != nil {
		panic(err)
	}

	req, err := c.NewRequest(method, path, bytes.NewReader(b))
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", "application/json")

	return req, nil
}

type MultipartWriter struct {
	*multipart.Writer

	c io.Closer
}

func (w *MultipartWriter) Close() error {
	if err := w.Writer.Close(); err != nil {
		return err
	}
	return w.c.Close()
}

// NewMultipartRequest creates a new multipart request.
//
// The multipart request is written as long as it is sent to the API. That means
// that writing the request and sending it MUST be done in parallel. If the
// request fails, subsequent writes to the multipart writer will fail with an
// io.ErrClosedPipe error.
func (c *client) NewMultipartRequest(method, path string) (req *http.Request, w *MultipartWriter, err error) {
	// The pipe will connect the multipart writer and the HTTP request body.
	pr, pw := io.Pipe()

	// pw needs to be closed once the multipart writer is closed.
	w = &MultipartWriter{
		multipart.NewWriter(pw),
		pw,
	}

	req, err = c.NewRequest(method, path, pr)
	if err != nil {
		return
	}

	req.Header.Add("Content-Type", w.FormDataContentType())
	return
}
