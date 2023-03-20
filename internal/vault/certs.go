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

package vault

func (vault *Vault) GetBridgeTLSCert() []byte {
	return vault.get().Certs.Bridge.Cert
}

func (vault *Vault) GetBridgeTLSKey() []byte {
	return vault.get().Certs.Bridge.Key
}

// SetBridgeTLSCertKey sets the path to PEM-encoded certificates for the bridge.
func (vault *Vault) SetBridgeTLSCertKey(cert, key []byte) error {
	return vault.mod(func(data *Data) {
		data.Certs.Bridge.Cert = cert
		data.Certs.Bridge.Key = key
	})
}

func (vault *Vault) GetCertsInstalled() bool {
	return vault.get().Certs.Installed
}

func (vault *Vault) SetCertsInstalled(installed bool) error {
	return vault.mod(func(data *Data) {
		data.Certs.Installed = installed
	})
}
