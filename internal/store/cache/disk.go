// Copyright (c) 2022 Proton AG
//
// This file is part of Proton Mail Bridge.
//
// Proton Mail Bridge is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Proton Mail Bridge is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with Proton Mail Bridge. If not, see <https://www.gnu.org/licenses/>.

package cache

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"

	"github.com/ProtonMail/proton-bridge/v2/pkg/algo"
	"github.com/ProtonMail/proton-bridge/v2/pkg/semaphore"
	"github.com/ricochet2200/go-disk-usage/du"
)

var (
	ErrMsgCorrupted = errors.New("ecrypted file was corrupted")
	ErrLowSpace     = errors.New("not enough free space left on device")
)

// IsOnDiskCache will return true if Cache is type of onDiskCache.
func IsOnDiskCache(c Cache) bool {
	_, ok := c.(*onDiskCache)
	return ok
}

type onDiskCache struct {
	path string
	opts Options

	gcm        map[string]cipher.AEAD
	cmp        Compressor
	rsem, wsem semaphore.Semaphore
	pending    *pending

	diskSize uint64
	diskFree uint64
	once     *sync.Once
	lock     sync.Mutex
}

func NewOnDiskCache(path string, cmp Compressor, opts Options) (Cache, error) {
	if err := os.MkdirAll(path, 0o700); err != nil {
		return nil, err
	}

	file, err := ioutil.TempFile(path, "tmp")
	defer func() {
		file.Close()           //nolint:errcheck,gosec
		os.Remove(file.Name()) //nolint:errcheck,gosec
	}()
	if err != nil {
		return nil, fmt.Errorf("cannot open test write target: %w", err)
	}
	if _, err := file.Write([]byte("test-write")); err != nil {
		return nil, fmt.Errorf("cannot write to target: %w", err)
	}

	usage := du.NewDiskUsage(path)

	// NOTE(GODT-1158): use Available() or Free()?
	return &onDiskCache{
		path: path,
		opts: opts,

		gcm:     make(map[string]cipher.AEAD),
		cmp:     cmp,
		rsem:    semaphore.New(opts.ConcurrentRead),
		wsem:    semaphore.New(opts.ConcurrentWrite),
		pending: newPending(),

		diskSize: usage.Size(),
		diskFree: usage.Available(),
		once:     &sync.Once{},
	}, nil
}

func (c *onDiskCache) Lock(userID string) {
	delete(c.gcm, userID)
}

func (c *onDiskCache) Unlock(userID string, passphrase []byte) error {
	aes, err := aes.NewCipher(algo.Hash256(passphrase))
	if err != nil {
		return err
	}

	gcm, err := cipher.NewGCM(aes)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(c.getUserPath(userID), 0o700); err != nil {
		return err
	}

	c.gcm[userID] = gcm

	return nil
}

func (c *onDiskCache) Delete(userID string) error {
	defer c.update()

	return os.RemoveAll(c.getUserPath(userID))
}

// Has returns whether the given message exists in the cache.
func (c *onDiskCache) Has(userID, messageID string) bool {
	c.pending.wait(c.getMessagePath(userID, messageID))

	c.rsem.Lock()
	defer c.rsem.Unlock()

	_, err := os.Stat(c.getMessagePath(userID, messageID))

	switch {
	case err == nil:
		return true

	case os.IsNotExist(err):
		return false

	default:
		// Cannot decide whether the message is cached or not.
		// Potential recover needs to be don in caller function.
		panic(err)
	}
}

func (c *onDiskCache) Get(userID, messageID string) ([]byte, error) {
	gcm, ok := c.gcm[userID]
	if !ok || gcm == nil {
		return nil, ErrCacheNeedsUnlock
	}

	enc, err := c.readFile(c.getMessagePath(userID, messageID))
	if err != nil {
		return nil, err
	}

	// Data stored in file must larger than NonceSize.
	if len(enc) <= gcm.NonceSize() {
		return nil, ErrMsgCorrupted
	}

	cmp, err := gcm.Open(nil, enc[:gcm.NonceSize()], enc[gcm.NonceSize():], nil)
	if err != nil {
		return nil, err
	}

	return c.cmp.Decompress(cmp)
}

func (c *onDiskCache) Set(userID, messageID string, literal []byte) error {
	gcm, ok := c.gcm[userID]
	if !ok {
		return ErrCacheNeedsUnlock
	}
	nonce := make([]byte, gcm.NonceSize())

	if _, err := rand.Read(nonce); err != nil {
		return err
	}

	cmp, err := c.cmp.Compress(literal)
	if err != nil {
		return err
	}

	// NOTE(GODT-1158, GODT-1488): Need to properly handle low space. Don't
	// return error, that's bad. Send event and clean least used message.
	if !c.hasSpace(len(cmp)) {
		return nil
	}

	return c.writeFile(c.getMessagePath(userID, messageID), gcm.Seal(nonce, nonce, cmp, nil))
}

func (c *onDiskCache) Rem(userID, messageID string) error {
	defer c.update()

	return os.Remove(c.getMessagePath(userID, messageID))
}

func (c *onDiskCache) readFile(path string) ([]byte, error) {
	c.rsem.Lock()
	defer c.rsem.Unlock()

	// Wait before reading in case the file is currently being written.
	c.pending.wait(path)

	return ioutil.ReadFile(filepath.Clean(path))
}

func (c *onDiskCache) writeFile(path string, b []byte) error {
	c.wsem.Lock()
	defer c.wsem.Unlock()

	// Mark the file as currently being written.
	// If it's already being written, wait for it to be done and return nil.
	// NOTE(GODT-1158): Let's hope it succeeded...
	if ok := c.pending.add(path); !ok {
		c.pending.wait(path)
		return nil
	}
	defer c.pending.done(path)

	// Reduce the approximate free space (update it exactly later).
	c.lock.Lock()
	c.diskFree -= uint64(len(b))
	c.lock.Unlock()

	// Update the diskFree eventually.
	defer c.update()

	// NOTE(GODT-1158): What happens when this fails? Should be fixed eventually.
	return ioutil.WriteFile(filepath.Clean(path), b, 0o600)
}

func (c *onDiskCache) hasSpace(size int) bool {
	c.lock.Lock()
	defer c.lock.Unlock()

	if c.opts.MinFreeAbs > 0 {
		if c.diskFree-uint64(size) < c.opts.MinFreeAbs {
			return false
		}
	}

	if c.opts.MinFreeRat > 0 {
		if float64(c.diskFree-uint64(size))/float64(c.diskSize) < c.opts.MinFreeRat {
			return false
		}
	}

	return true
}

func (c *onDiskCache) update() {
	go func() {
		c.once.Do(func() {
			c.lock.Lock()
			defer c.lock.Unlock()

			// Update the free space.
			c.diskFree = du.NewDiskUsage(c.path).Available()

			// Reset the Once object (so we can update again).
			c.once = &sync.Once{}
		})
	}()
}

func (c *onDiskCache) getUserPath(userID string) string {
	return filepath.Join(c.path, algo.HashHexSHA256(userID))
}

func (c *onDiskCache) getMessagePath(userID, messageID string) string {
	return filepath.Join(c.getUserPath(userID), algo.HashHexSHA256(messageID))
}
