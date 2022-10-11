package user

import (
	"context"
	"encoding/hex"
	"fmt"
	"reflect"

	"gitlab.protontech.ch/go/liteapi"
)

// mapTo converts the slice to the given type.
// This is not runtime safe, so make sure the slice is of the correct type!
// (This is a workaround for the fact that slices cannot be converted to other types generically).
func mapTo[From, To any](from []From) []To {
	to := make([]To, 0, len(from))

	for _, from := range from {
		val, ok := reflect.ValueOf(from).Convert(reflect.TypeOf(to).Elem()).Interface().(To)
		if !ok {
			panic(fmt.Sprintf("cannot convert %T to %T", from, *new(To)))
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
func getAddrID(apiAddrs []liteapi.Address, email string) (string, error) {
	for _, addr := range apiAddrs {
		if addr.Email == email {
			return addr.ID, nil
		}
	}

	return "", fmt.Errorf("address %s not found", email)
}

// getAddrEmail returns the email address of the given address ID.
func getAddrEmail(apiAddrs []liteapi.Address, addrID string) (string, error) {
	for _, addr := range apiAddrs {
		if addr.ID == addrID {
			return addr.Email, nil
		}
	}

	return "", fmt.Errorf("address %s not found", addrID)
}

// contextWithStopCh returns a new context that is cancelled when the stop channel is closed or a value is sent to it.
func contextWithStopCh(ctx context.Context, stopCh <-chan struct{}) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(ctx)

	go func() {
		select {
		case <-stopCh:
			cancel()

		case <-ctx.Done():
			// ...
		}
	}()

	return ctx, cancel
}
