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

package user

import (
	"encoding/hex"
	"fmt"
	"reflect"
	"strings"

	"gitlab.protontech.ch/go/liteapi"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
)

// mapTo converts the slice to the given type.
// This is not runtime safe, so make sure the slice is of the correct type!
// (This is a workaround for the fact that slices cannot be converted to other types generically).
func mapTo[From, To any](from []From) []To {
	to := make([]To, 0, len(from))

	for _, from := range from {
		val, ok := reflect.ValueOf(from).Convert(reflect.TypeOf(to).Elem()).Interface().(To)
		if !ok {
			panic(fmt.Sprintf("cannot convert %T to %T", from, *new(To))) //nolint:gocritic
		}

		to = append(to, val)
	}

	return to
}

// groupBy returns a map of the given slice grouped by the given key.
// Duplicate keys are overwritten.
func groupBy[Key comparable, Value any](items []Value, key func(Value) Key) map[Key]Value {
	groups := make(map[Key]Value)

	for _, item := range items {
		groups[key(item)] = item
	}

	return groups
}

// sortAddr returns whether the first address should be sorted before the second.
func sortAddr(addrIDA, addrIDB string, apiAddrs map[string]liteapi.Address) bool {
	return apiAddrs[addrIDA].Order < apiAddrs[addrIDB].Order
}

// hexEncode returns the hexadecimal encoding of the given byte slice.
func hexEncode(b []byte) []byte {
	enc := make([]byte, hex.EncodedLen(len(b)))

	hex.Encode(enc, b)

	return enc
}

// hexDecode returns the bytes represented by the hexadecimal encoding of the given byte slice.
func hexDecode(b []byte) ([]byte, error) {
	dec := make([]byte, hex.DecodedLen(len(b)))

	if _, err := hex.Decode(dec, b); err != nil {
		return nil, err
	}

	return dec, nil
}

// getAddrID returns the address ID for the given email address.
func getAddrID(apiAddrs map[string]liteapi.Address, email string) (string, error) {
	for _, addr := range apiAddrs {
		if strings.EqualFold(addr.Email, sanitizeEmail(email)) {
			return addr.ID, nil
		}
	}

	return "", fmt.Errorf("address %s not found", email)
}

// getAddrIdx returns the address with the given index.
func getAddrIdx(apiAddrs map[string]liteapi.Address, idx int) (liteapi.Address, error) {
	sorted := sortSlice(maps.Values(apiAddrs), func(a, b liteapi.Address) bool {
		return a.Order < b.Order
	})

	if idx < 0 || idx >= len(sorted) {
		return liteapi.Address{}, fmt.Errorf("address index %d out of range", idx)
	}

	return sorted[idx], nil
}

// sortSlice returns the given slice sorted by the given comparator.
func sortSlice[Item any](items []Item, less func(Item, Item) bool) []Item {
	sorted := make([]Item, len(items))

	copy(sorted, items)

	slices.SortFunc(sorted, less)

	return sorted
}
