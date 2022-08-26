package vault

func (vault *Vault) GetCookies() ([]byte, error) {
	return vault.get().Cookies, nil
}

func (vault *Vault) SetCookies(cookies []byte) error {
	return vault.mod(func(data *Data) {
		data.Cookies = cookies
	})
}
