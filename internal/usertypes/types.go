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

package usertypes

import (
	"fmt"
	"reflect"
	"runtime"
	"strings"

	"github.com/ProtonMail/gluon/async"
	"github.com/ProtonMail/go-proton-api"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
)

// MapTo converts the slice to the given type.
// This is not runtime safe, so make sure the slice is of the correct type!
// (This is a workaround for the fact that slices cannot be converted to other types generically).
func MapTo[From, To any](from []From) []To {
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

// GroupBy returns a map of the given slice grouped by the given key.
// Duplicate keys are overwritten.
func GroupBy[Key comparable, Value any](items []Value, key func(Value) Key) map[Key]Value {
	groups := make(map[Key]Value)

	for _, item := range items {
		groups[key(item)] = item
	}

	return groups
}

// GetAddrID returns the address ID for the given email address.
func GetAddrID(apiAddrs map[string]proton.Address, email string) (string, error) {
	for _, addr := range apiAddrs {
		if strings.EqualFold(addr.Email, SanitizeEmail(email)) {
			return addr.ID, nil
		}
	}

	return "", fmt.Errorf("address %s not found", email)
}

// GetAddrIdx returns the address with the given index.
func GetAddrIdx(apiAddrs map[string]proton.Address, idx int) (proton.Address, error) {
	sorted := sortSlice(maps.Values(apiAddrs), func(a, b proton.Address) bool {
		return a.Order < b.Order
	})

	if idx < 0 || idx >= len(sorted) {
		return proton.Address{}, fmt.Errorf("address index %d out of range", idx)
	}

	return sorted[idx], nil
}

func GetPrimaryAddr(apiAddrs map[string]proton.Address) (proton.Address, error) {
	sorted := sortSlice(maps.Values(apiAddrs), func(a, b proton.Address) bool {
		return a.Order < b.Order
	})

	if len(sorted) == 0 {
		return proton.Address{}, fmt.Errorf("no addresses available")
	}

	return sorted[0], nil
}

// sortSlice returns the given slice sorted by the given comparator.
func sortSlice[Item any](items []Item, less func(Item, Item) bool) []Item {
	sorted := make([]Item, len(items))

	copy(sorted, items)

	slices.SortFunc(sorted, less)

	return sorted
}

func NewProtonAPIScheduler(panicHandler async.PanicHandler) proton.Scheduler {
	return proton.NewParallelScheduler(runtime.NumCPU()/2, panicHandler)
}

func SanitizeEmail(email string) string {
	splitAt := strings.Split(email, "@")
	if len(splitAt) != 2 {
		return email
	}

	return strings.Split(splitAt[0], "+")[0] + "@" + splitAt[1]
}
