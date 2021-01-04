// Copyright (c) 2021 Proton Technologies AG
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

package updates

import (
	"bytes"
	"encoding/hex"
	"errors"
	"io"
	"os"
	"os/exec"
	"runtime"

	"golang.org/x/crypto/openpgp"
)

// gpg --export D51E64D3E63EDC3EEF7864CEE2C75D68E6234B07 | xxd -p | tr -d '\n' | xclip
const (
	keyID     = "D51E64D3E63EDC3EEF7864CEE2C75D68E6234B07"
	pubkeyHex = "99020d045a3d39e1011000be7cfacb714058f9851ce5888cad8250ea882b2563060b4d21f5f02fdcfb2b1e4073d33edc4050f235d35ab689050ed6435f5d79334bf30f093936472398eec259b7fa9c265cc18a8e4ab0681fcc0d4fdb934bfea9477007935af70bfdfa406de25b1e96838c9c4645d6613a13dffca1e70684e416cb8bff5348101d7c9cd3cb1b78229646dce4cc9cec2ecfa78d456547900e3c089c9d592e8cf33fa322c07016ae273880a1bcb8c8178d8fb804f555a8826129b2e2c535631c56a1fe73f476345e0851a5deda508833008b1751b6845e1ff788264350c3792f0932027fabe63dd230dce4da1b45f15eea584f25758355ae9784c32a2bd31d70333a5b6ff0b863cc177bcacfd35774029887551113cec424d9eb1f5ee4ab042b69c8b73a113d6596e88bdac55451e9403ee7944253b26177cbd97f79f22d138010a2e9044f5f16cf8c23ec7755332cf09250d50efbfedfb256426b1dba775af591b1f324b00dd497abeabc681036848954825139c7832902ab9ace0b73d270611d39a222e07e5ce98159acb99dcf8d3d62624458b2883d5feee43a48f981a601f01ebeb8e430181004e9990fc1a9d7d2e746d9aa8d5876ad576bf327399c08e834e6706a73300f9bc258f51510b597b9b34506ff21a993311d9a961ab07cc7c86476088d9aecaab31cc198e1091d62e5bdb161bc784879d4fca5a53fb292ffa89996a77101e10011010001b44c50726f746f6e20546563686e6f6c6f67696573204147202850726f746f6e4d61696c2042726964676520646576656c6f7065727329203c6272696467654070726f746f6e6d61696c2e63683e89025404130108003e021b03050b09080702061508090a0b020416020301021e01021780162104d51e64d3e63edc3eef7864cee2c75d68e6234b0705025bf7efbc050903bcdedb000a0910e2c75d68e6234b07e19e0fff58d859aabd12b6e37829b61d406f8563e68b8a32e18952378b88fac148acfbb51614d31419e3ec054ac5ef33abce8e15a75e85491861a6f7886c3315260e321a5cf037e8d9afd141bc5f3391b3aa8da8e44c5deb3b2a92402a5956079a991b7a7388064133b07f0c41e0ae66560e2bbfae484882a32d4b42676ee0a43e9e26cdb92e1c9942723ae491e91c39156abee84369db5afd2d9f6c2bad428113d851339267d7119a1d892b6313de1cd0c7aeac495036cf7c985c6b9e0cf7d6ecf50ba6dca913d56594e7fd2dac0d15f1e5196bddf3c9d2cdca724972a6ea294601bce9e9ccbff497785d7df9a8969517585ae514f94e2d54252b829cea842a4b62960724901e0f64953b68d69f5ca4d644dec3b9f947e57b25d6d1d37a0dc53fc62c2860a27ffb1e09658b6fa87fe43fb82f7b6a339c4b0b330ea4d906683b51a2844a685e939ecece3af8447f606c78645da77b66e6627ce838e6416c65f0b0c65335a6bd4091db7356f3f638e320505cd739ab762d27b1b8e5bcbba011821b49bfd63f4c0ac06ebacd36adb25448436ba795424d9d66f413854b833f72ab9ed0a6fc52ec3df5d9d655d0d9c0c21c8323ee785c08af5d341bd1f0067dc81bb4a74aa496f575c00c6fd58e067a6b04aed72ec59b263e6fe6707702e59eafb361532241fe881642da91518d5b01c238701d7317062afcf2f4b1b7c01c3c7164417a5e18b9020d045a3d39e1011000b74f40e9514f5a261e58f19cdeb88c5b835d74886facea681bd4c1d6280e957675f466cc6925b02540ed61d70660d66d47bd6af80edce3bb00dc3afc9cf36233fe8f0226a4983b712a93a96f00d54e9361f1975545776bfa1b204b6d97b6c8f41b1436098bb01608c07a2ffa097f6a32907402d81c29aadc18faa3b0417963e956993d7738dcb0b55b98db333463f5224e5d284d236548ba87e62856e244f68ffe4fe8f1a2a151a5736c66a58e10a9821c3cfa4d3b35893a7e0a867fe60d1e2a987a5c9f93b8da2a17de3dd48604167f68c73b2820191ec1363f7c95a65dffbe727a41383ad92c5fc6a1115cecc5bd29a7a5f6f990d09ee68ccd9b16cdf4f251f4ad1adf99d94b014e622a03efd1a9319e79d4aaf09ca1e905314443b4fdf3d3d2995ee38c9edf37e2eb5ded62def8b643e9dc6cfab3e1c83896f3413c05d973d49a3e73683ab6820dfae0fbfdfe60c1a65a5e1241d283bb7d6495568f87269ad606799ff5fae02aad4e712d0136f03ee008a9da1845a962406d44170b231a794d964d6233cb4150e92105172fd133c3b93b51c7b03623b4889cf5027f3ea483d2a37b2cc9648d1f00756bc0c66f798b91a9c4364e0e576c3a779eba69afb27dc8d630a850b4e293b88955d46e635577bdfb1802479f18dfe7da06b732f6ad0898a9c28086b0c8145c64dda245cea8a060301b812c29318f6c25a25d91f69f11001101000189023c041801080026021b0c162104d51e64d3e63edc3eef7864cee2c75d68e6234b0705025bf7f281050903bce1a0000a0910e2c75d68e6234b07b7930ffe31f08ec8ecc415eb767c8434a1262c528b8f5b7d77b293396b672ac06ea36bec9ac82b3817e45d474d4d091779be54692196930895317f8cbabd7f5bf3991ba23efbb1fe322d5d86e71c58619457a4c19917c789d2e62c23a4c0ede82d4445fb6ff99e721caf75999b2be9e35c7c0e9578874a0507a796d66f0d5b2a1d2544b03861b0a278ef7b5abb0b8f5e4b3b72aabcb6c4a05731f344ef8257a1d5d5bdccba2cb9640ef7442d68206613979cc8935334876148c827df7d3044859fdd00c12bcf072881303e100abe3a5ca94e2c36497643471e6c43c4fed35f24777cee259e0f873243a7343adbbce1cbed4ce073d838733f9e2a4e9281a5f43f2aac56ff0c472843a07814a5515ae809ed976d0699ebce1f5e5661fd6752f22af8521cc485ea2925bc8c650865dab398fbd64460fd873f687fd2b7db55d1920fd5787010063eba5d4b08fd9882e9c2244270886f8c6411194d4e55d207e374d6bf9ea3463ce4db2f2e6818f57ac964f76f79b1df0b9dc3e688f0c5d73f33010809f9ed2effc6e4387ce0f1eb634e7a67bf04e9e30126de137999e7fdea05b9ee6088154d8369e5fac81b7c0af16d6be8d636bd84348812333822319cd0e362ab06969032b57f9233e618ec9f67d5d65a52c51f5f942cf83f5522bfe3de3bab39b3de867f5f6a108e0b661789c2a049b990c812f2b1b1722d3e403299f969eb34a9dc21af"
)

var (
	pubkeyRing = openpgp.EntityList{} //nolint[gochecknoglobals]
)

func singAndVerify(pathToFile string) (err error) {
	err = signFile(pathToFile)
	if err != nil {
		err = verifyFile(pathToFile)
	}
	return
}

func signFile(pathToFile string) (err error) {
	if runtime.GOOS != "linux" { //nolint[goconst]
		return errors.New("tar not implemented only for linux")
	}
	// assuming gpg detach-sign creates file with suffix .sig by default.
	// Lstat does not follow the link i.e. only link is deleted (not link target).
	if _, err := os.Lstat(pathToFile + sigExtension); !os.IsNotExist(err) {
		_ = os.Remove(pathToFile + sigExtension)
	}
	cmd := exec.Command("gpg", "--local-user", keyID, "--detach-sign", pathToFile) //nolint[gosec]
	return cmd.Run()
}

func verifyFile(pathToFile string) error {
	fileReader, err := os.Open(pathToFile) //nolint[gosec]
	if err != nil {
		return err
	}
	defer fileReader.Close() //nolint[errcheck]

	signatureReader, err := os.Open(pathToFile + sigExtension) //nolint[gosec]
	if err != nil {
		return err
	}
	defer signatureReader.Close() //nolint[errcheck]

	return verifyBytes(fileReader, signatureReader)
}

func verifyBytes(fileReader, signatureReader io.Reader) (err error) {
	if _, err = getPubKey(); err != nil {
		return err
	}

	_, err = openpgp.CheckDetachedSignature(pubkeyRing, fileReader, signatureReader, nil)
	/*
		if err != nil {
			return err
		}

		if signer == nil || signer.PrimaryKey.KeyId != keyID {
			return errors.New("Signer with wrong key ID")
		}
	*/
	return
}

// from opengpg/read_test.go
func getPubKey() (el openpgp.EntityList, err error) {
	if pubkeyRing != nil && len(pubkeyRing) != 0 {
		return pubkeyRing, nil
	}
	data, err := hex.DecodeString(pubkeyHex)
	if err != nil {
		return
	}
	pubkeyRing, err = openpgp.ReadKeyRing(bytes.NewBuffer(data))
	return pubkeyRing, err
}
