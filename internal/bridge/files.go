package bridge

import (
	"os"
	"path/filepath"
)

func moveDir(from, to string) error {
	entries, err := os.ReadDir(from)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			if err := os.Mkdir(filepath.Join(to, entry.Name()), 0700); err != nil {
				return err
			}

			if err := moveDir(filepath.Join(from, entry.Name()), filepath.Join(to, entry.Name())); err != nil {
				return err
			}

			if err := os.RemoveAll(filepath.Join(from, entry.Name())); err != nil {
				return err
			}
		} else {
			if err := move(filepath.Join(from, entry.Name()), filepath.Join(to, entry.Name())); err != nil {
				return err
			}
		}
	}

	return os.Remove(from)
}

func move(from, to string) error {
	if err := os.MkdirAll(filepath.Dir(to), 0700); err != nil {
		return err
	}

	f, err := os.Open(from) // nolint:gosec
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()

	c, err := os.Create(to) // nolint:gosec
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()

	if err := os.Chmod(to, 0600); err != nil {
		return err
	}

	if _, err := c.ReadFrom(f); err != nil {
		return err
	}

	if err := os.Remove(from); err != nil {
		return err
	}

	return nil
}
