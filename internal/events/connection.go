package events

import "gitlab.protontech.ch/go/liteapi"

type TLSIssue struct {
	eventBase
}

type ConnStatus struct {
	eventBase

	Status liteapi.Status
}
