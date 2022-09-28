package tests

import (
	"github.com/ProtonMail/proton-bridge/v2/internal/certs"
	"gitlab.protontech.ch/go/liteapi/server/account"
)

func init() {
	// Use the fast key generation for tests.
	account.GenerateKey = FastGenerateKey

	// Use the fast cert generation for tests.
	certs.GenerateCert = FastGenerateCert
}
