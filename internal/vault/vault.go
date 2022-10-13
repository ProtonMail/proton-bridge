package vault

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"math/rand"
	"os"
	"path/filepath"
	"sync"

	"github.com/ProtonMail/proton-bridge/v2/internal/certs"
	"github.com/bradenaw/juniper/xslices"
)

// Vault is an encrypted data vault that stores bridge and user data.
type Vault struct {
	path string
	gcm  cipher.AEAD

	enc     []byte
	encLock sync.RWMutex

	ref     map[string]int
	refLock sync.Mutex
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

// HasUser returns true if the vault contains a user with the given ID.
func (vault *Vault) HasUser(userID string) bool {
	return xslices.IndexFunc(vault.get().Users, func(user UserData) bool {
		return user.UserID == userID
	}) >= 0
}

// GetUser provides access to a vault user. It returns an error if the user does not exist.
func (vault *Vault) GetUser(userID string, fn func(*User)) error {
	user, err := vault.NewUser(userID)
	if err != nil {
		return err
	}
	defer func() { _ = user.Close() }()

	fn(user)

	return nil
}

// NewUser returns a new vault user. It must be closed before it can be deleted.
func (vault *Vault) NewUser(userID string) (*User, error) {
	if idx := xslices.IndexFunc(vault.get().Users, func(user UserData) bool {
		return user.UserID == userID
	}); idx < 0 {
		return nil, errors.New("no such user")
	}

	return vault.attachUser(userID), nil
}

// ForUser executes a callback for each user in the vault.
func (vault *Vault) ForUser(fn func(*User) error) error {
	for _, userID := range vault.GetUserIDs() {
		user, err := vault.NewUser(userID)
		if err != nil {
			return err
		}
		defer func() { _ = user.Close() }()

		if err := fn(user); err != nil {
			return err
		}
	}

	return nil
}

// AddUser creates a new user in the vault with the given ID and username.
// A bridge password and gluon key are generated using the package's token generator.
func (vault *Vault) AddUser(userID, username, authUID, authRef string, keyPass []byte) (*User, error) {
	if idx := xslices.IndexFunc(vault.get().Users, func(user UserData) bool {
		return user.UserID == userID
	}); idx >= 0 {
		return nil, errors.New("user already exists")
	}

	if err := vault.mod(func(data *Data) {
		data.Users = append(data.Users, newDefaultUser(userID, username, authUID, authRef, keyPass))
	}); err != nil {
		return nil, err
	}

	return vault.NewUser(userID)
}

// DeleteUser removes the given user from the vault.
func (vault *Vault) DeleteUser(userID string) error {
	vault.refLock.Lock()
	defer vault.refLock.Unlock()

	if _, ok := vault.ref[userID]; ok {
		return fmt.Errorf("user %s is currently in use", userID)
	}

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

func (vault *Vault) Close() error {
	vault.refLock.Lock()
	defer vault.refLock.Unlock()

	if len(vault.ref) > 0 {
		return errors.New("vault is still in use")
	}

	vault.gcm = nil

	return nil
}

func (vault *Vault) attachUser(userID string) *User {
	vault.refLock.Lock()
	defer vault.refLock.Unlock()

	vault.ref[userID] += 1

	return &User{
		vault:  vault,
		userID: userID,
	}
}

func (vault *Vault) detachUser(userID string) error {
	vault.refLock.Lock()
	defer vault.refLock.Unlock()

	if _, ok := vault.ref[userID]; !ok {
		return fmt.Errorf("user %s is not attached", userID)
	}

	vault.ref[userID] -= 1

	if vault.ref[userID] == 0 {
		delete(vault.ref, userID)
	}

	return nil
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

	return &Vault{
		path: path,
		enc:  enc,
		gcm:  gcm,
		ref:  make(map[string]int),
	}, corrupt, nil
}

func (vault *Vault) get() Data {
	vault.encLock.RLock()
	defer vault.encLock.RUnlock()

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
	vault.encLock.Lock()
	defer vault.encLock.Unlock()

	dec, err := decrypt(vault.gcm, vault.enc)
	if err != nil {
		return err
	}

	var data Data

	if err := json.Unmarshal(dec, &data); err != nil {
		return err
	}

	fn(&data)

	mod, err := json.Marshal(data)
	if err != nil {
		return err
	}

	enc, err := encrypt(vault.gcm, mod)
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
