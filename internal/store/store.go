// Copyright (c) 2020 Proton Technologies AG
//
// This file is part of ProtonMail Bridge.
//
// ProtonMail Bridge is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// ProtonMail Bridge is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with ProtonMail Bridge.  If not, see <https://www.gnu.org/licenses/>.

// Package store communicates with API and caches metadata in a local database.
package store

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/ProtonMail/proton-bridge/pkg/listener"
	"github.com/ProtonMail/proton-bridge/pkg/pmapi"
	imapBackend "github.com/emersion/go-imap/backend"
	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	bolt "go.etcd.io/bbolt"
)

const (
	// PathDelimiter for IMAP
	PathDelimiter = "/"
	// UserLabelsMailboxName for IMAP
	UserLabelsMailboxName = "Labels"
	// UserLabelsPrefix contains name with delimiter for IMAP
	UserLabelsPrefix = UserLabelsMailboxName + PathDelimiter
	// UserFoldersMailboxName for IMAP
	UserFoldersMailboxName = "Folders"
	// UserFoldersPrefix contains name with delimiter for IMAP
	UserFoldersPrefix = UserFoldersMailboxName + PathDelimiter
)

var (
	log = logrus.WithField("pkg", "store") //nolint[gochecknoglobals]

	// Database structure:
	// * metadata
	//   * {messageID} -> message data (subject, from, to, time, headers, body size, ...)
	// * counts
	//   * {mailboxID} -> mailboxCounts: totalOnAPI, unreadOnAPI, labelName, labelColor, labelIsExclusive
	// * address_info
	//   * {index} -> {address, addressID}
	// * address_mode
	//   * mode -> string split or combined
	// * mailboxes_version
	//     * version -> uint32 value
	// * sync_state
	//   * sync_state -> string timestamp when it was last synced (when missing, sync should be ongoing)
	//   * ids_ranges -> json array of groups with start and end message ID (when missing, there is no ongoing sync)
	//   * ids_to_be_deleted -> json array of message IDs to be deleted after sync (when missing, there is no ongoing sync)
	// * mailboxes
	//   * {addressID+mailboxID}
	//     * imap_ids
	//       * {imapUID} -> string messageID
	//     * api_ids
	//       * {messageID} -> uint32 imapUID
	metadataBucket    = []byte("metadata")          //nolint[gochecknoglobals]
	countsBucket      = []byte("counts")            //nolint[gochecknoglobals]
	addressInfoBucket = []byte("address_info")      //nolint[gochecknoglobals]
	addressModeBucket = []byte("address_mode")      //nolint[gochecknoglobals]
	syncStateBucket   = []byte("sync_state")        //nolint[gochecknoglobals]
	mailboxesBucket   = []byte("mailboxes")         //nolint[gochecknoglobals]
	imapIDsBucket     = []byte("imap_ids")          //nolint[gochecknoglobals]
	apiIDsBucket      = []byte("api_ids")           //nolint[gochecknoglobals]
	mboxVersionBucket = []byte("mailboxes_version") //nolint[gochecknoglobals]

	// ErrNoSuchAPIID when mailbox does not have API ID.
	ErrNoSuchAPIID = errors.New("no such api id") //nolint[gochecknoglobals]
	// ErrNoSuchUID when mailbox does not have IMAP UID.
	ErrNoSuchUID = errors.New("no such uid") //nolint[gochecknoglobals]
	// ErrNoSuchSeqNum when mailbox does not have IMAP ID.
	ErrNoSuchSeqNum = errors.New("no such sequence number") //nolint[gochecknoglobals]
)

// Store is local user storage, which handles the synchronization between IMAP and PM API.
type Store struct {
	panicHandler  PanicHandler
	eventLoop     *eventLoop
	user          BridgeUser
	clientManager ClientManager

	log *logrus.Entry

	cache       *Cache
	filePath    string
	db          *bolt.DB
	lock        *sync.RWMutex
	addresses   map[string]*Address
	imapUpdates chan imapBackend.Update

	isSyncRunning bool
	addressMode   addressMode
}

// New creates or opens a store for the given `user`.
func New(
	panicHandler PanicHandler,
	user BridgeUser,
	clientManager ClientManager,
	events listener.Listener,
	path string,
	cache *Cache,
) (store *Store, err error) {
	if user == nil || clientManager == nil || events == nil || cache == nil {
		return nil, fmt.Errorf("missing parameters - user: %v, api: %v, events: %v, cache: %v", user, clientManager, events, cache)
	}

	l := log.WithField("user", user.ID())

	var firstInit bool
	if _, existErr := os.Stat(path); os.IsNotExist(existErr) {
		l.Info("Creating new store database file with address mode from user's credentials store")
		firstInit = true
	} else {
		l.Info("Store database file already exists, using mode already set")
		firstInit = false
	}

	bdb, err := openBoltDatabase(path)
	if err != nil {
		err = errors.Wrap(err, "failed to open store database")
		return
	}

	store = &Store{
		panicHandler:  panicHandler,
		clientManager: clientManager,
		user:          user,
		cache:         cache,
		filePath:      path,
		db:            bdb,
		lock:          &sync.RWMutex{},
		log:           l,
	}

	if err = store.init(firstInit); err != nil {
		l.WithError(err).Error("Could not initialise store, attempting to close")
		if storeCloseErr := store.Close(); storeCloseErr != nil {
			l.WithError(storeCloseErr).Warn("Could not close uninitialised store")
		}
		err = errors.Wrap(err, "failed to initialise store")
		return
	}

	if user.IsConnected() {
		store.eventLoop = newEventLoop(cache, store, user, events)
		go func() {
			defer store.panicHandler.HandlePanic()
			store.eventLoop.start()
		}()
	}

	return store, err
}

func openBoltDatabase(filePath string) (db *bolt.DB, err error) {
	l := log.WithField("path", filePath)
	l.Debug("Opening bolt database")

	if db, err = bolt.Open(filePath, 0600, &bolt.Options{Timeout: 1 * time.Second}); err != nil {
		l.WithError(err).Error("Could not open bolt database")
		return
	}

	if val, set := os.LookupEnv("BRIDGESTRICTMODE"); set && val == "1" {
		db.StrictMode = true
	}

	tx := func(tx *bolt.Tx) (err error) {
		if _, err = tx.CreateBucketIfNotExists(metadataBucket); err != nil {
			return
		}

		if _, err = tx.CreateBucketIfNotExists(countsBucket); err != nil {
			return
		}

		if _, err = tx.CreateBucketIfNotExists(addressInfoBucket); err != nil {
			return
		}

		if _, err = tx.CreateBucketIfNotExists(addressModeBucket); err != nil {
			return
		}

		if _, err = tx.CreateBucketIfNotExists(syncStateBucket); err != nil {
			return
		}

		if _, err = tx.CreateBucketIfNotExists(mailboxesBucket); err != nil {
			return
		}

		if _, err = tx.CreateBucketIfNotExists(mboxVersionBucket); err != nil {
			return
		}

		return
	}

	if err = db.Update(tx); err != nil {
		return
	}

	return db, err
}

// init initialises the store for the given addresses.
func (store *Store) init(firstInit bool) (err error) {
	if store.addresses != nil {
		store.log.Warn("Store was already initialised")
		return
	}

	// If it's the first time we are creating the store, use the mode set in the
	// user's credentials, otherwise read it from the DB (if present).
	if firstInit {
		if store.user.IsCombinedAddressMode() {
			err = store.setAddressMode(combinedMode)
		} else {
			err = store.setAddressMode(splitMode)
		}
		if err != nil {
			return errors.Wrap(err, "first init setting store address mode")
		}
	} else if store.addressMode, err = store.getAddressMode(); err != nil {
		store.log.WithError(err).Error("Store address mode is unknown, setting to combined mode")
		if err = store.setAddressMode(combinedMode); err != nil {
			return errors.Wrap(err, "setting store address mode")
		}
	}

	store.log.WithField("mode", store.addressMode).Debug("Initialising store")

	labels, err := store.initCounts()
	if err != nil {
		store.log.WithError(err).Error("Could not initialise label counts")
		return
	}

	if err = store.initAddresses(labels); err != nil {
		store.log.WithError(err).Error("Could not initialise store addresses")
		return
	}

	return err
}

func (store *Store) client() pmapi.Client {
	return store.clientManager.GetClient(store.UserID())
}

// initCounts initialises the counts for each label. It tries to use the API first to fetch the labels but if
// the API is unavailable for whatever reason it tries to fetch the labels locally.
func (store *Store) initCounts() (labels []*pmapi.Label, err error) {
	if labels, err = store.client().ListLabels(); err != nil {
		store.log.WithError(err).Warn("Could not list API labels. Trying with local labels.")
		if labels, err = store.getLabelsFromLocalStorage(); err != nil {
			store.log.WithError(err).Error("Cannot list local labels")
			return
		}
	} else {
		// the labels listed by PMAPI don't include system folders so we need to add them.
		for _, counts := range getSystemFolders() {
			labels = append(labels, counts.getPMLabel())
		}

		if err = store.createOrUpdateMailboxCountsBuckets(labels); err != nil {
			store.log.WithError(err).Error("Cannot create counts")
			return
		}

		if countsErr := store.updateCountsFromServer(); countsErr != nil {
			store.log.WithError(countsErr).Warning("Continue without new counts from server")
		}
	}

	sortByOrder(labels)

	return
}

// initAddresses creates address objects in the store for each necessary address.
// In combined mode this means just one mailbox for all addresses but in split mode this means one mailbox per address.
func (store *Store) initAddresses(labels []*pmapi.Label) (err error) {
	store.addresses = make(map[string]*Address)

	addrInfo, err := store.GetAddressInfo()
	if err != nil {
		store.log.WithError(err).Error("Could not get addresses and address IDs")
		return
	}

	// We need at least one address to continue.
	if len(addrInfo) < 1 {
		err = errors.New("no addresses to initialise")
		store.log.WithError(err).Warn("There are no addresses to initialise")
		return
	}

	// If in combined mode, we only need the user's primary address.
	if store.addressMode == combinedMode {
		addrInfo = addrInfo[:1]
	}

	for _, addr := range addrInfo {
		if err = store.addAddress(addr.Address, addr.AddressID, labels); err != nil {
			store.log.WithField("address", addr.Address).WithError(err).Error("Could not add address to store")
		}
	}

	return err
}

// addAddress adds a new address to the store. If the address exists already it is overwritten.
func (store *Store) addAddress(address, addressID string, labels []*pmapi.Label) (err error) {
	if _, ok := store.addresses[addressID]; ok {
		store.log.WithField("addressID", addressID).Debug("Overwriting store address")
	}

	addr, err := newAddress(store, address, addressID, labels)
	if err != nil {
		return errors.Wrap(err, "failed to create store address object")
	}

	store.addresses[addressID] = addr

	return
}

// Close stops the event loop and closes the database to free the file.
func (store *Store) Close() error {
	store.lock.Lock()
	defer store.lock.Unlock()

	return store.close()
}

// CloseEventLoop stops the eventloop (if it is present).
func (store *Store) CloseEventLoop() {
	if store.eventLoop != nil {
		store.eventLoop.stop()
	}
}

func (store *Store) close() error {
	store.CloseEventLoop()
	return store.db.Close()
}

// Remove closes and removes the database file and clears the cache file.
func (store *Store) Remove() (err error) {
	store.lock.Lock()
	defer store.lock.Unlock()

	store.log.Trace("Removing store")

	var result *multierror.Error

	if err = store.close(); err != nil {
		result = multierror.Append(result, errors.Wrap(err, "failed to close store"))
	}

	if err = RemoveStore(store.cache, store.filePath, store.user.ID()); err != nil {
		result = multierror.Append(result, errors.Wrap(err, "failed to remove store"))
	}

	return result.ErrorOrNil()
}

// RemoveStore removes the database file and clears the cache file.
func RemoveStore(cache *Cache, path, userID string) error {
	var result *multierror.Error

	if err := cache.clearCacheUser(userID); err != nil {
		result = multierror.Append(result, errors.Wrap(err, "failed to clear event loop user cache"))
	}

	// RemoveAll will not return an error if the path does not exist.
	if err := os.RemoveAll(path); err != nil {
		result = multierror.Append(result, errors.Wrap(err, "failed to remove database file"))
	}

	return result.ErrorOrNil()
}
