package vault

import (
	"github.com/ProtonMail/gopenpgp/v2/crypto"
)

// RandomToken is a function that returns a random token.
var RandomToken func(size int) ([]byte, error)

// By default, we use crypto.RandomToken to generate tokens.
func init() {
	RandomToken = crypto.RandomToken
}
