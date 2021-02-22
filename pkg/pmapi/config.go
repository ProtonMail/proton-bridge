package pmapi

type Config struct {
	HostURL    string
	AppVersion string
}

var DefaultConfig = Config{
	HostURL:    "https://api.protonmail.ch",
	AppVersion: "Other",
}
