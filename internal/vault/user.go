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

func (user *User) UpdateKeyPass(keyPass []byte) error {
	return user.vault.modUser(user.userID, func(data *UserData) {
		data.KeyPass = keyPass
	})
}

// UpdateAuth updates the auth secrets for the given user.
func (user *User) UpdateAuth(authUID, authRef string) error {
	return user.vault.modUser(user.userID, func(data *UserData) {
		data.AuthUID = authUID
		data.AuthRef = authRef
	})
}

// UpdateGluonData updates the gluon ID and key for the given user.
func (user *User) UpdateGluonData(gluonID string, gluonKey []byte) error {
	return user.vault.modUser(user.userID, func(data *UserData) {
		data.GluonID = gluonID
		data.GluonKey = gluonKey
	})
}

// UpdateEventID updates the event ID for the given user.
func (user *User) UpdateEventID(eventID string) error {
	return user.vault.modUser(user.userID, func(data *UserData) {
		data.EventID = eventID
	})
}

// UpdateSync updates the sync state for the given user.
func (user *User) UpdateSync(hasSync bool) error {
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
