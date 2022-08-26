package bridge

func (bridge *Bridge) GetLogsPath() (string, error) {
	return bridge.locator.ProvideLogsPath()
}

func (bridge *Bridge) GetLicenseFilePath() string {
	return bridge.locator.GetLicenseFilePath()
}

func (bridge *Bridge) GetDependencyLicensesLink() string {
	return bridge.locator.GetDependencyLicensesLink()
}
