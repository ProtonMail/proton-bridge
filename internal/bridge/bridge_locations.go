package bridge

func (bridge *Bridge) ProvideLogsPath() (string, error) {
	return bridge.locations.ProvideLogsPath()
}

func (bridge *Bridge) GetLicenseFilePath() string {
	return bridge.locations.GetLicenseFilePath()
}

func (bridge *Bridge) GetDependencyLicensesLink() string {
	return bridge.locations.GetDependencyLicensesLink()
}
