package vault

type User struct {
	vault  *Vault
	userID string
}

func (user *User) UserID() string {
	return user.vault.getUser(user.userID).UserID
}

func (user *User) Username() string {
	return user.vault.getUser(user.userID).Username
}

func (user *User) GluonID() string {
	return user.vault.getUser(user.userID).GluonID
}

func (user *User) GluonKey() []byte {
	return user.vault.getUser(user.userID).GluonKey
}

func (user *User) BridgePass() string {
	return user.vault.getUser(user.userID).BridgePass
}

func (user *User) AuthUID() string {
	return user.vault.getUser(user.userID).AuthUID
}

func (user *User) AuthRef() string {
	return user.vault.getUser(user.userID).AuthRef
}

func (user *User) KeyPass() []byte {
	return user.vault.getUser(user.userID).KeyPass
}

func (user *User) EventID() string {
	return user.vault.getUser(user.userID).EventID
}

func (user *User) HasSync() bool {
	return user.vault.getUser(user.userID).HasSync
}

func (user *User) SetKeyPass(keyPass []byte) error {
	return user.vault.modUser(user.userID, func(data *UserData) {
		data.KeyPass = keyPass
	})
}

// SetAuth updates the auth secrets for the given user.
func (user *User) SetAuth(authUID, authRef string) error {
	return user.vault.modUser(user.userID, func(data *UserData) {
		data.AuthUID = authUID
		data.AuthRef = authRef
	})
}

// SetGluonAuth updates the gluon ID and key for the given user.
func (user *User) SetGluonAuth(gluonID string, gluonKey []byte) error {
	return user.vault.modUser(user.userID, func(data *UserData) {
		data.GluonID = gluonID
		data.GluonKey = gluonKey
	})
}

// SetEventID updates the event ID for the given user.
func (user *User) SetEventID(eventID string) error {
	return user.vault.modUser(user.userID, func(data *UserData) {
		data.EventID = eventID
	})
}

// SetSync updates the sync state for the given user.
func (user *User) SetSync(hasSync bool) error {
	return user.vault.modUser(user.userID, func(data *UserData) {
		data.HasSync = hasSync
	})
}

// Clear clears the secrets for the given user.
func (user *User) Clear() error {
	return user.vault.modUser(user.userID, func(data *UserData) {
		data.AuthUID = ""
		data.AuthRef = ""
		data.KeyPass = nil
	})
}
