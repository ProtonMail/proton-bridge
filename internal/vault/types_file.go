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

func unmarshalFile[T any](gcm cipher.AEAD, b []byte, data *T) error {
	var f File

	if err := msgpack.Unmarshal(b, &f); err != nil {
		return err
	}

	dec, err := gcm.Open(nil, f.Data[:gcm.NonceSize()], f.Data[gcm.NonceSize():], nil)
	if err != nil {
		return err
	}

	for v := f.Version; v < Current; v++ {
		if dec, err = upgrade(v, dec); err != nil {
			return err
		}
	}

	return msgpack.Unmarshal(dec, data)
}

func marshalFile[T any](gcm cipher.AEAD, t T) ([]byte, error) {
	dec, err := msgpack.Marshal(t)
	if err != nil {
		return nil, err
	}

	nonce, err := crypto.RandomToken(gcm.NonceSize())
	if err != nil {
		return nil, err
	}

	return msgpack.Marshal(File{
		Version: Current,
		Data:    gcm.Seal(nonce, nonce, dec, nil),
	})
}
