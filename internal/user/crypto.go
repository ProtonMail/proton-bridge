package user

import (
	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"gitlab.protontech.ch/go/liteapi"
)

func unlockKeyRings(user liteapi.User, addresses []liteapi.Address, keyPass []byte) (*crypto.KeyRing, map[string]*crypto.KeyRing, error) {
	userKR, err := user.Keys.Unlock(keyPass, nil)
	if err != nil {
		return nil, nil, err
	}

	addrKRs := make(map[string]*crypto.KeyRing)

	for _, address := range addresses {
		if !address.HasKeys.Bool() {
			continue
		}

		addrKR, err := address.Keys.Unlock(keyPass, userKR)
		if err != nil {
			return nil, nil, err
		}

		addrKRs[address.ID] = addrKR
	}

	return userKR, addrKRs, nil
}
