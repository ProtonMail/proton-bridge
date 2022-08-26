package bridge

func (bridge *Bridge) GetBridgeTLSCert() ([]byte, []byte) {
	return bridge.vault.GetBridgeTLSCert(), bridge.vault.GetBridgeTLSKey()
}
