package events

import "github.com/ProtonMail/proton-bridge/v2/internal/updater"

type UpdateAvailable struct {
	eventBase

	Version updater.VersionInfo

	CanInstall bool
}

type UpdateNotAvailable struct {
	eventBase
}

type UpdateInstalled struct {
	eventBase

	Version updater.VersionInfo
}

type UpdateForced struct {
	eventBase
}
