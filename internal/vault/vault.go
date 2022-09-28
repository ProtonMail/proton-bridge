package vault

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"io/fs"
	"math/rand"
	"os"
	"path/filepath"

	"github.com/ProtonMail/gluon/imap"
	"github.com/ProtonMail/proton-bridge/v2/internal/certs"
	"github.com/bradenaw/juniper/xslices"
)

var (
	ErrInsecure = errors.New("the vault is insecure")
	ErrCorrupt  = errors.New("the vault is corrupt")
)

type Vault struct {
	path string
	enc  []byte
	gcm  cipher.AEAD
}

// New constructs a new encrypted data vault at the given filepath using the given encryption key.
func New(vaultDir, gluonDir string, key []byte) (*Vault, bool, error) {
	if err := os.MkdirAll(vaultDir, 0o700); err != nil {
		return nil, false, err
	}

	hash256 := sha256.Sum256(key)

	aes, err := aes.NewCipher(hash256[:])
	if err != nil {
		return nil, false, err
	}

	gcm, err := cipher.NewGCM(aes)
	if err != nil {
		return nil, false, err
	}

	vault, corrupt, err := newVault(filepath.Join(vaultDir, "vault.enc"), gluonDir, gcm)
	if err != nil {
		return nil, false, err
	}

	return vault, corrupt, nil
}

// GetUserIDs returns the user IDs and usernames of all users in the vault.
func (vault *Vault) GetUserIDs() []string {
	return xslices.Map(vault.get().Users, func(user UserData) string {
		return user.UserID
	})
}

// GetUserIDs returns the user IDs and usernames of all users in the vault.
func (vault *Vault) GetUser(userID string) (*User, error) {
	if idx := xslices.IndexFunc(vault.get().Users, func(user UserData) bool {
		return user.UserID == userID
	}); idx < 0 {
		return nil, errors.New("no such user")
	}

	return &User{
		vault:  vault,
		userID: userID,
	}, nil
}

// ForUser executes a callback for each user in the vault.
func (vault *Vault) ForUser(fn func(*User) error) error {
	for _, userID := range vault.GetUserIDs() {
		user, err := vault.GetUser(userID)
		if err != nil {
			return err
		}

		if err := fn(user); err != nil {
			return err
		}
	}

	return nil
}

// AddUser creates a new user in the vault with the given ID and username.
// A bridge password is generated using the package's token generator.
func (vault *Vault) AddUser(userID, username, authUID, authRef string, keyPass []byte) (*User, error) {
	if idx := xslices.IndexFunc(vault.get().Users, func(user UserData) bool {
		return user.UserID == userID
	}); idx >= 0 {
		return nil, errors.New("user already exists")
	}

	if err := vault.mod(func(data *Data) {
		data.Users = append(data.Users, UserData{
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
		})
	}); err != nil {
		return nil, err
	}

	return vault.GetUser(userID)
}

func (vault *Vault) ClearUser(userID string) error {
	return vault.modUser(userID, func(data *UserData) {
		data.AuthUID = ""
		data.AuthRef = ""
		data.KeyPass = nil
	})
}

// DeleteUser removes the given user from the vault.
func (vault *Vault) DeleteUser(userID string) error {
	return vault.mod(func(data *Data) {
		idx := xslices.IndexFunc(data.Users, func(user UserData) bool {
			return user.UserID == userID
		})

		if idx < 0 {
			return
		}

		data.Users = append(data.Users[:idx], data.Users[idx+1:]...)
	})
}

func newVault(path, gluonDir string, gcm cipher.AEAD) (*Vault, bool, error) {
	if _, err := os.Stat(path); errors.Is(err, fs.ErrNotExist) {
		if _, err := initVault(path, gluonDir, gcm); err != nil {
			return nil, false, err
		}
	}

	enc, err := os.ReadFile(path)
	if err != nil {
		return nil, false, err
	}

	var corrupt bool

	if _, err := decrypt(gcm, enc); err != nil {
		corrupt = true

		newEnc, err := initVault(path, gluonDir, gcm)
		if err != nil {
			return nil, false, err
		}

		enc = newEnc
	}

	return &Vault{path: path, enc: enc, gcm: gcm}, corrupt, nil
}

func (vault *Vault) get() Data {
	dec, err := decrypt(vault.gcm, vault.enc)
	if err != nil {
		panic(err)
	}

	var data Data

	if err := json.Unmarshal(dec, &data); err != nil {
		panic(err)
	}

	return data
}

func (vault *Vault) mod(fn func(data *Data)) error {
	data := vault.get()

	fn(&data)

	return vault.set(data)
}

func (vault *Vault) set(data Data) error {
	dec, err := json.Marshal(data)
	if err != nil {
		return err
	}

	enc, err := encrypt(vault.gcm, dec)
	if err != nil {
		return err
	}

	vault.enc = enc

	return os.WriteFile(vault.path, vault.enc, 0o600)
}

func (vault *Vault) getUser(userID string) UserData {
	return vault.get().Users[xslices.IndexFunc(vault.get().Users, func(user UserData) bool {
		return user.UserID == userID
	})]
}

func (vault *Vault) modUser(userID string, fn func(userData *UserData)) error {
	return vault.mod(func(data *Data) {
		idx := xslices.IndexFunc(data.Users, func(user UserData) bool {
			return user.UserID == userID
		})

		fn(&data.Users[idx])
	})
}

func initVault(path, gluonDir string, gcm cipher.AEAD) ([]byte, error) {
	bridgeCert, err := newTLSCert()
	if err != nil {
		return nil, err
	}

	dec, err := json.Marshal(Data{
		Settings: newDefaultSettings(gluonDir),

		Certs: Certs{
			Bridge: bridgeCert,
		},
	})
	if err != nil {
		return nil, err
	}

	enc, err := encrypt(gcm, dec)
	if err != nil {
		return nil, err
	}

	if err := os.WriteFile(path, enc, 0o600); err != nil {
		return nil, err
	}

	return enc, nil
}

func decrypt(gcm cipher.AEAD, enc []byte) ([]byte, error) {
	return gcm.Open(nil, enc[:gcm.NonceSize()], enc[gcm.NonceSize():], nil)
}

func encrypt(gcm cipher.AEAD, data []byte) ([]byte, error) {
	nonce := make([]byte, gcm.NonceSize())

	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}

	return gcm.Seal(nonce, nonce, data, nil), nil
}

func newTLSCert() (Cert, error) {
	template, err := certs.NewTLSTemplate()
	if err != nil {
		return Cert{}, err
	}

	certPEM, keyPEM, err := certs.GenerateCert(template)
	if err != nil {
		return Cert{}, err
	}

	return Cert{
		Cert: certPEM,
		Key:  keyPEM,
	}, nil
}
