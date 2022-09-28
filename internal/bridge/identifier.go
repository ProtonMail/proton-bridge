package bridge

func (bridge *Bridge) GetCurrentUserAgent() string {
	return bridge.identifier.GetUserAgent()
}

func (bridge *Bridge) SetCurrentPlatform(platform string) {
	bridge.identifier.SetPlatform(platform)
}
