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

	SyncStatus SyncStatus
	EventID    string
}

type AddressMode int

const (
	CombinedMode AddressMode = iota
	SplitMode
)

type SyncStatus struct {
	HasLabels     bool
	HasMessages   bool
	LastMessageID string
}

func (status SyncStatus) IsComplete() bool {
	return status.HasLabels && status.HasMessages
}

func newDefaultUser(userID, username, authUID, authRef string, keyPass []byte) UserData {
	return UserData{
		UserID:   userID,
		Username: username,

		GluonKey:    newRandomToken(32),
		GluonIDs:    make(map[string]string),
		UIDValidity: make(map[string]imap.UID),
		BridgePass:  newRandomToken(16),
		AddressMode: CombinedMode,

		AuthUID: authUID,
		AuthRef: authRef,
		KeyPass: keyPass,
	}
}
