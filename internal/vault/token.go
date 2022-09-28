package vault

import (
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
