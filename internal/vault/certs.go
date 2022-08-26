package vault

func (vault *Vault) GetBridgeTLSCert() []byte {
	return vault.get().Certs.Bridge.Cert
}

func (vault *Vault) GetBridgeTLSKey() []byte {
	return vault.get().Certs.Bridge.Key
}

func (vault *Vault) GetCertsInstalled() bool {
	return vault.get().Certs.Installed
}

func (vault *Vault) SetCertsInstalled(installed bool) error {
	return vault.mod(func(data *Data) {
		data.Certs.Installed = installed
	})
}
