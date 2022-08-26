package vault

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
)

type Keychain struct {
	Helper string
}

func GetHelper(vaultDir string) (string, error) {
	var keychain Keychain

	if _, err := os.Stat(filepath.Join(vaultDir, "keychain.json")); errors.Is(err, fs.ErrNotExist) {
		return "", nil
	}

	b, err := os.ReadFile(filepath.Join(vaultDir, "keychain.json"))
	if err != nil {
		return "", err
	}

	if err := json.Unmarshal(b, &keychain); err != nil {
		return "", err
	}

	return keychain.Helper, nil
}

func SetHelper(vaultDir, helper string) error {
	b, err := json.MarshalIndent(Keychain{Helper: helper}, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(vaultDir, "keychain.json"), b, 0o600)
}
