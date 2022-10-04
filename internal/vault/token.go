package vault

import (
	"encoding/hex"

	"github.com/ProtonMail/gopenpgp/v2/crypto"
)

// RandomToken is a function that returns a random token.
// By default, we use crypto.RandomToken to generate tokens.
var RandomToken = crypto.RandomToken

func newRandomToken(size int) []byte {
	token, err := RandomToken(size)
	if err != nil {
		panic(err)
	}

	return token
}

func newRandomString(size int) []byte {
	token, err := RandomToken(size)
	if err != nil {
		panic(err)
	}

	return []byte(hex.EncodeToString(token))
}
