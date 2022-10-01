package events

import "time"

type SyncStarted struct {
	eventBase

	UserID string
}

type SyncProgress struct {
	eventBase

	UserID    string
	Progress  float64
	Elapsed   time.Duration
	Remaining time.Duration
}

type SyncFinished struct {
	eventBase

	UserID string
}

type SyncFailed struct {
	eventBase

	UserID string
	Err    error
}
