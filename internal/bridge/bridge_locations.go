package bridge

func (b *Bridge) ProvideLogsPath() (string, error) {
	return b.locations.ProvideLogsPath()
}

func (b *Bridge) GetLicenseFilePath() string {
	return b.locations.GetLicenseFilePath()
}

func (b *Bridge) GetDependencyLicensesLink() string {
	return b.locations.GetDependencyLicensesLink()
}
