package certs

type Installer struct{}

func NewInstaller() *Installer {
	return &Installer{}
}

func (installer *Installer) InstallCert(certPEM []byte) error {
	return installCert(certPEM)
}

func (installer *Installer) UninstallCert(certPEM []byte) error {
	return uninstallCert(certPEM)
}
