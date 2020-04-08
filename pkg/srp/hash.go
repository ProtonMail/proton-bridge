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

package srp

import (
	"bytes"
	"crypto/md5" //nolint[gosec]
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"strings"

	"github.com/jameskeane/bcrypt"
)

// BCryptHash function bcrypt algorithm to hash password with salt
func BCryptHash(password string, salt string) (string, error) {
	return bcrypt.Hash(password, salt)
}

// ExpandHash extends the byte data for SRP flow
func ExpandHash(data []byte) []byte {
	part0 := sha512.Sum512(append(data, 0))
	part1 := sha512.Sum512(append(data, 1))
	part2 := sha512.Sum512(append(data, 2))
	part3 := sha512.Sum512(append(data, 3))
	return bytes.Join([][]byte{
		part0[:],
		part1[:],
		part2[:],
		part3[:],
	}, []byte{})
}

// HashPassword returns the hash of password argument. Based on version number
// following arguments are used in addition to password:
// * 0, 1, 2: userName and modulus
// * 3, 4: salt and modulus
func HashPassword(authVersion int, password, userName string, salt, modulus []byte) ([]byte, error) {
	switch authVersion {
	case 4, 3:
		return hashPasswordVersion3(password, salt, modulus)
	case 2:
		return hashPasswordVersion2(password, userName, modulus)
	case 1:
		return hashPasswordVersion1(password, userName, modulus)
	case 0:
		return hashPasswordVersion0(password, userName, modulus)
	default:
		return nil, errors.New("pmapi: unsupported auth version")
	}
}

// CleanUserName returns the input string in lower-case without characters `_`,
// `.` and `-`.
func CleanUserName(userName string) string {
	userName = strings.Replace(userName, "-", "", -1)
	userName = strings.Replace(userName, ".", "", -1)
	userName = strings.Replace(userName, "_", "", -1)
	return strings.ToLower(userName)
}

func hashPasswordVersion3(password string, salt, modulus []byte) (res []byte, err error) {
	encodedSalt := base64.NewEncoding("./ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789").WithPadding(base64.NoPadding).EncodeToString(append(salt, []byte("proton")...))
	crypted, err := BCryptHash(password, "$2y$10$"+encodedSalt)
	if err != nil {
		return
	}

	return ExpandHash(append([]byte(crypted), modulus...)), nil
}

func hashPasswordVersion2(password, userName string, modulus []byte) (res []byte, err error) {
	return hashPasswordVersion1(password, CleanUserName(userName), modulus)
}

func hashPasswordVersion1(password, userName string, modulus []byte) (res []byte, err error) {
	prehashed := md5.Sum([]byte(strings.ToLower(userName))) //nolint[gosec]
	encodedSalt := hex.EncodeToString(prehashed[:])
	crypted, err := BCryptHash(password, "$2y$10$"+encodedSalt)
	if err != nil {
		return
	}

	return ExpandHash(append([]byte(crypted), modulus...)), nil
}

func hashPasswordVersion0(password, userName string, modulus []byte) (res []byte, err error) {
	prehashed := sha512.Sum512([]byte(password))
	return hashPasswordVersion1(base64.StdEncoding.EncodeToString(prehashed[:]), userName, modulus)
}
