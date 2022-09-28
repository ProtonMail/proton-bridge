package vault

import (
	"math/rand"

	"github.com/Masterminds/semver/v3"
	"github.com/ProtonMail/gluon/imap"
	"github.com/ProtonMail/proton-bridge/v2/internal/updater"
)

type Data struct {
	Settings Settings
	Users    []UserData
	Cookies  []byte
	Certs    Certs
}

type Certs struct {
	Bridge    Cert
	Installed bool
}

type Cert struct {
	Cert, Key []byte
}

type Settings struct {
	GluonDir string

	IMAPPort int
	SMTPPort int
	IMAPSSL  bool
	SMTPSSL  bool

	UpdateChannel updater.Channel
	UpdateRollout float64

	ColorScheme  string
	ProxyAllowed bool
	ShowAllMail  bool
	Autostart    bool
	AutoUpdate   bool

	LastVersion   *semver.Version
	FirstStart    bool
	FirstStartGUI bool
}

type AddressMode int

const (
	CombinedMode AddressMode = iota
	SplitMode
)

// UserData holds information about a single bridge user.
// The user may or may not be logged in.
type UserData struct {
	UserID   string
	Username string

	GluonKey    []byte
	GluonIDs    map[string]string
	UIDValidity map[string]imap.UID
	BridgePass  []byte
	AddressMode AddressMode

	AuthUID string
	AuthRef string
	KeyPass []byte

	EventID string
	HasSync bool
}

func newDefaultSettings(gluonDir string) Settings {
	return Settings{
		GluonDir: gluonDir,

		IMAPPort: 1143,
		SMTPPort: 1025,
		IMAPSSL:  false,
		SMTPSSL:  false,

		UpdateChannel: updater.DefaultUpdateChannel,
		UpdateRollout: rand.Float64(),

		ColorScheme:  "",
		ProxyAllowed: true,
		ShowAllMail:  true,
		Autostart:    false,
		AutoUpdate:   true,

		LastVersion:   semver.MustParse("0.0.0"),
		FirstStart:    true,
		FirstStartGUI: true,
	}
}
