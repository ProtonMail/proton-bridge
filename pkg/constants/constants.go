// Package constants contains variables that are set via ldflags during build.
package constants

// nolint[gochecknoglobals]
var (
	// Version of the build.
	Version = ""

	// Revision is current hash of the build.
	Revision = ""

	// BuildTime stamp of the build.
	BuildTime = ""

	// AppShortName to make setup.
	AppShortName = "bridge"

	// DSNSentry client keys to be able to report crashes to Sentry.
	DSNSentry = ""

	// LongVersion is derived from Version and Revision.
	LongVersion = Version + " (" + Revision + ")"

	// BuildVersion is derived from LongVersion and BuildTime.
	BuildVersion = LongVersion + " " + BuildTime
)
