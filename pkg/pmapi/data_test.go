package pmapi

import "github.com/ProtonMail/gopenpgp/v2/crypto"

var testIdentity = &crypto.Identity{
	Name:  "UserID",
	Email: "",
}

const (
	testUsername    = "jason"
	testAPIPassword = "apple"

	testUID             = "729ad6012421d67ad26950dc898bebe3a6e3caa2" //nolint[gosec]
	testAccessToken     = "de0423049b44243afeec7d9c1d99be7b46da1e8a" //nolint[gosec]
	testAccessTokenOld  = "feb3159ac63fb05119bcf4480d939278aa746926" //nolint[gosec]
	testRefreshToken    = "a49b98256745bb497bec20e9b55f5de16f01fb52" //nolint[gosec]
	testRefreshTokenNew = "b894b4c4f20003f12d486900d8b88c7d68e67235" //nolint[gosec]
)
