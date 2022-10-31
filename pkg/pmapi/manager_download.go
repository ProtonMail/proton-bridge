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

package pmapi

import (
	"io"

	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"golang.org/x/net/context"
)

// DownloadAndVerify downloads a file and its signature from the given locations `file` and `sig`.
// The file and its signature are verified using the given keyring `kr`.
// If the file is verified successfully, it can be read from the returned reader.
// TLS fingerprinting is used to verify that connections are only made to known servers.
func (m *manager) DownloadAndVerify(kr *crypto.KeyRing, url, sig string) ([]byte, error) {
	fb, err := m.fetchFile(url)
	if err != nil {
		return nil, err
	}

	sb, err := m.fetchFile(sig)
	if err != nil {
		return nil, err
	}

	if err := kr.VerifyDetached(
		crypto.NewPlainMessage(fb),
		crypto.NewPGPSignature(sb),
		crypto.GetUnixTime(),
	); err != nil {
		return nil, err
	}

	return fb, nil
}

func (m *manager) fetchFile(url string) ([]byte, error) {
	res, err := m.r(ContextWithoutRetry(context.Background())).SetDoNotParseResponse(true).Get(url)
	if err != nil {
		return nil, err
	}

	b, err := io.ReadAll(res.RawBody())
	if err != nil {
		return nil, err
	}

	if err := res.RawBody().Close(); err != nil {
		return nil, err
	}

	return b, nil
}
