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
// along with Proton Mail Bridge.  If not, see <https://www.gnu.org/licenses/>.

package vault

import (
	"crypto/cipher"

	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"github.com/vmihailenco/msgpack/v5"
)

// File holds a versioned, serialized data.
type File struct {
	Version Version
	Data    []byte
}

func unmarshalFile[T any](gcm cipher.AEAD, enc []byte, data *T) error {
	dec, err := gcm.Open(nil, enc[:gcm.NonceSize()], enc[gcm.NonceSize():], nil)
	if err != nil {
		return err
	}

	var f File

	if err := msgpack.Unmarshal(dec, &f); err != nil {
		return err
	}

	for v := f.Version; v < Current; v++ {
		b, err := upgrade(v, f.Data)
		if err != nil {
			return err
		}

		f.Data = b
	}

	if err := msgpack.Unmarshal(f.Data, data); err != nil {
		return err
	}

	return nil
}

func marshalFile[T any](gcm cipher.AEAD, t T) ([]byte, error) {
	b, err := msgpack.Marshal(t)
	if err != nil {
		return nil, err
	}

	dec, err := msgpack.Marshal(File{Version: Current, Data: b})
	if err != nil {
		return nil, err
	}

	nonce, err := crypto.RandomToken(gcm.NonceSize())
	if err != nil {
		return nil, err
	}

	return gcm.Seal(nonce, nonce, dec, nil), nil
}
