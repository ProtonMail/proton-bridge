package pmapi

type AuthModulus struct {
	Modulus   string
	ModulusID string
}

type GetAuthInfoReq struct {
	Username string
}

type AuthInfo struct {
	Version         int
	Modulus         string
	ServerEphemeral string
	Salt            string
	SRPSession      string
}

type TwoFAInfo struct {
	Enabled TwoFAStatus
}

type TwoFAStatus int

const (
	TwoFADisabled TwoFAStatus = iota
	TOTPEnabled
	// TODO: Support UTF
)

type PasswordMode int

const (
	OnePasswordMode PasswordMode = iota + 1
	TwoPasswordMode
)

type AuthReq struct {
	Username        string
	ClientProof     string
	ClientEphemeral string
	SRPSession      string
}

type Auth struct {
	UserID string

	UID          string
	AccessToken  string
	RefreshToken string
	ExpiresIn    int64

	Scope       string
	ServerProof string

	TwoFA        TwoFAInfo `json:"2FA"`
	PasswordMode PasswordMode
}

type Auth2FAReq struct {
	TwoFactorCode string
}

type AuthRefreshReq struct {
	UID          string
	RefreshToken string
	ResponseType string
	GrantType    string
	RedirectURI  string
	State        string
}
