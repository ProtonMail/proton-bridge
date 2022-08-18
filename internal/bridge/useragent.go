package bridge

func (b *Bridge) GetCurrentUserAgent() string {
	return b.userAgent.String()
}

func (b *Bridge) SetCurrentPlatform(platform string) {
	b.userAgent.SetPlatform(platform)
}
