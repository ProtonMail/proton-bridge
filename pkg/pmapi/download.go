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

package pmapi

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/ProtonMail/gopenpgp/v2/crypto"
)

// DownloadAndVerify downloads a file and its signature from the given locations `file` and `sig`.
// The file and its signature are verified using the given keyring `kr`.
// If the file is verified successfully, it can be read from the returned reader.
// TLS fingerprinting is used to verify that connections are only made to known servers.
func (c *client) DownloadAndVerify(file, sig string, kr *crypto.KeyRing) (io.Reader, error) {
	var fb, sb []byte

	if err := c.fetchFile(file, func(r io.Reader) (err error) {
		fb, err = ioutil.ReadAll(r)
		return err
	}); err != nil {
		return nil, err
	}

	if err := c.fetchFile(sig, func(r io.Reader) (err error) {
		sb, err = ioutil.ReadAll(r)
		return err
	}); err != nil {
		return nil, err
	}

	if err := kr.VerifyDetached(
		crypto.NewPlainMessage(fb),
		crypto.NewPGPSignature(sb),
		crypto.GetUnixTime(),
	); err != nil {
		return nil, err
	}

	return bytes.NewReader(fb), nil
}

func (c *client) fetchFile(file string, fn func(io.Reader) error) error {
	res, err := c.hc.Get(file)
	if err != nil {
		return err
	}
	defer func() { _ = res.Body.Close() }()

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to get file: http error %v", res.StatusCode)
	}

	return fn(res.Body)
}
