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

// Package store communicates with API and caches metadata in a local database.
package store

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/ProtonMail/proton-bridge/v2/internal/sentry"
	"github.com/ProtonMail/proton-bridge/v2/internal/store/cache"
	"github.com/ProtonMail/proton-bridge/v2/pkg/listener"
	"github.com/ProtonMail/proton-bridge/v2/pkg/message"
	"github.com/ProtonMail/proton-bridge/v2/pkg/pmapi"
	"github.com/ProtonMail/proton-bridge/v2/pkg/pool"
	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	bolt "go.etcd.io/bbolt"
)

const (
	// PathDelimiter for IMAP.
	PathDelimiter = "/"
	// UserLabelsMailboxName for IMAP.
	UserLabelsMailboxName = "Labels"
	// UserLabelsPrefix contains name with delimiter for IMAP.
	UserLabelsPrefix = UserLabelsMailboxName + PathDelimiter
	// UserFoldersMailboxName for IMAP.
	UserFoldersMailboxName = "Folders"
	// UserFoldersPrefix contains name with delimiter for IMAP.
	UserFoldersPrefix = UserFoldersMailboxName + PathDelimiter
)

var (
	log = logrus.WithField("pkg", "store") //nolint:gochecknoglobals

	// Database structure:
	// * metadata
	//   * {messageID} -> message data (subject, from, to, time, ...)
	// * headers
	//   * {messageID} -> header bytes
	// * bodystructure
	//   * {messageID} -> message body structure
	// * size
	//   * {messageID} -> uint32 value
	// * counts
	//   * {mailboxID} -> mailboxCounts: totalOnAPI, unreadOnAPI, labelName, labelColor, labelIsExclusive
	// * address_info
	//   * {index} -> {address, addressID}
	// * address_mode
	//   * mode -> string split or combined
	// * cache_passphrase
	//   * passphrase -> cache passphrase (pgp encrypted message)
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
	//     * deleted_ids (can be missing or have no keys)
	//       * {messageID} -> true
	metadataBucket        = []byte("metadata")          //nolint:gochecknoglobals
	headersBucket         = []byte("headers")           //nolint:gochecknoglobals
	bodystructureBucket   = []byte("bodystructure")     //nolint:gochecknoglobals
	sizeBucket            = []byte("size")              //nolint:gochecknoglobals
	countsBucket          = []byte("counts")            //nolint:gochecknoglobals
	addressInfoBucket     = []byte("address_info")      //nolint:gochecknoglobals
	addressModeBucket     = []byte("address_mode")      //nolint:gochecknoglobals
	cachePassphraseBucket = []byte("cache_passphrase")  //nolint:gochecknoglobals
	syncStateBucket       = []byte("sync_state")        //nolint:gochecknoglobals
	mailboxesBucket       = []byte("mailboxes")         //nolint:gochecknoglobals
	imapIDsBucket         = []byte("imap_ids")          //nolint:gochecknoglobals
	apiIDsBucket          = []byte("api_ids")           //nolint:gochecknoglobals
	deletedIDsBucket      = []byte("deleted_ids")       //nolint:gochecknoglobals
	mboxVersionBucket     = []byte("mailboxes_version") //nolint:gochecknoglobals

	// ErrNoSuchAPIID when mailbox does not have API ID.
	ErrNoSuchAPIID = errors.New("no such api id") //nolint:gochecknoglobals
	// ErrNoSuchUID when mailbox does not have IMAP UID.
	ErrNoSuchUID = errors.New("no such uid") //nolint:gochecknoglobals
	// ErrNoSuchSeqNum when mailbox does not have IMAP ID.
	ErrNoSuchSeqNum = errors.New("no such sequence number") //nolint:gochecknoglobals
)

// exposeContextForIMAP should be replaced once with context passed
// as an argument from IMAP package and IMAP library should cancel
// context when IMAP client cancels the request.
func exposeContextForIMAP() context.Context {
	return context.TODO()
}

// exposeContextForSMTP is the same as above but for SMTP.
func exposeContextForSMTP() context.Context {
	return context.TODO()
}

// Store is local user storage, which handles the synchronization between IMAP and PM API.
type Store struct {
	sentryReporter *sentry.Reporter
	panicHandler   PanicHandler
	user           BridgeUser
	eventLoop      *eventLoop
	currentEvents  *Events

	log *logrus.Entry

	filePath  string
	db        *bolt.DB
	lock      *sync.RWMutex
	addresses map[string]*Address
	notifier  ChangeNotifier

	builder      *message.Builder
	cache        cache.Cache
	msgCachePool *MsgCachePool
	done         chan struct{}

	isSyncRunning bool
	syncCooldown  cooldown
	addressMode   addressMode
}

// New creates or opens a store for the given `user`.
func New( //nolint:funlen
	sentryReporter *sentry.Reporter,
	panicHandler PanicHandler,
	user BridgeUser,
	listener listener.Listener,
	cache cache.Cache,
	builder *message.Builder,
	path string,
	currentEvents *Events,
) (store *Store, err error) {
	if user == nil || listener == nil || currentEvents == nil {
		return nil, fmt.Errorf("missing parameters - user: %v, listener: %v, currentEvents: %v", user, listener, currentEvents)
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
		return nil, errors.Wrap(err, "failed to open store database")
	}

	store = &Store{
		sentryReporter: sentryReporter,
		panicHandler:   panicHandler,
		user:           user,
		currentEvents:  currentEvents,

		log: l,

		filePath: path,
		db:       bdb,
		lock:     &sync.RWMutex{},

		builder: builder,
		cache:   cache,
	}

	// Create a new cacher. It's not started yet.
	// NOTE(GODT-1158): I hate this circular dependency store->cacher->store :(
	store.msgCachePool = newMsgCachePool(store)

	// Minimal increase is event pollInterval, doubles every failed retry up to 5 minutes.
	store.syncCooldown.setExponentialWait(pollInterval, 2, 5*time.Minute)

	if err = store.init(firstInit); err != nil {
		l.WithError(err).Error("Could not initialise store, attempting to close")
		if storeCloseErr := store.Close(); storeCloseErr != nil {
			l.WithError(storeCloseErr).Warn("Could not close uninitialised store")
		}
		err = errors.Wrap(err, "failed to initialise store")
		return
	}

	if user.IsConnected() {
		store.eventLoop = newEventLoop(currentEvents, store, user, listener)
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

	if db, err = bolt.Open(filePath, 0o600, &bolt.Options{Timeout: 1 * time.Second}); err != nil {
		l.WithError(err).Error("Could not open bolt database")
		return
	}

	if val, set := os.LookupEnv("BRIDGESTRICTMODE"); set && val == "1" {
		db.StrictMode = true
	}

	tx := func(tx *bolt.Tx) (err error) {
		buckets := [][]byte{
			metadataBucket,
			headersBucket,
			bodystructureBucket,
			sizeBucket,
			countsBucket,
			addressInfoBucket,
			addressModeBucket,
			cachePassphraseBucket,
			syncStateBucket,
			mailboxesBucket,
			mboxVersionBucket,
		}

		for _, bucket := range buckets {
			if _, err = tx.CreateBucketIfNotExists(bucket); err != nil {
				err = errors.Wrap(err, string(bucket))
				return
			}
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

	store.log.WithField("mode", store.addressMode).Info("Initialising store")

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
	return store.user.GetClient()
}

// initCounts initialises the counts for each label. It tries to use the API first to fetch the labels but if
// the API is unavailable for whatever reason it tries to fetch the labels locally.
func (store *Store) initCounts() (labels []*pmapi.Label, err error) {
	if labels, err = store.client().ListLabels(pmapi.ContextWithoutRetry(context.Background())); err != nil {
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

// newBuildJob returns a new build job for the given message using the store's message builder.
func (store *Store) newBuildJob(ctx context.Context, messageID string, priority int) (*message.Job, pool.DoneFunc) {
	return store.builder.NewJobWithOptions(
		ctx,
		store.client(),
		messageID,
		message.JobOptions{
			IgnoreDecryptionErrors: true, // Whether to ignore decryption errors and create a "custom message" instead.
			SanitizeDate:           true, // Whether to replace all dates before 1970 with RFC822's birthdate.
			AddInternalID:          true, // Whether to include MessageID as X-Pm-Internal-Id.
			AddExternalID:          true, // Whether to include ExternalID as X-Pm-External-Id.
			AddMessageDate:         true, // Whether to include message time as X-Pm-Date.
			AddMessageIDReference:  true, // Whether to include the MessageID in References.
		},
		priority,
	)
}

// Close stops the event loop and closes the database to free the file.
func (store *Store) Close() error {
	store.lock.Lock()
	defer store.lock.Unlock()

	return store.close()
}

// CloseEventLoopAndCacher stops the eventloop (if it is present).
func (store *Store) CloseEventLoopAndCacher() {
	if store.eventLoop != nil {
		store.eventLoop.stop()
	}

	store.stopWatcher()

	store.msgCachePool.stop()
}

func (store *Store) close() error {
	// Stop the event loop and cacher first before closing the DB.
	store.CloseEventLoopAndCacher()

	// Close the database.
	return store.db.Close()
}

// Remove closes and removes the database file and clears the cache file.
func (store *Store) Remove() error {
	store.lock.Lock()
	defer store.lock.Unlock()

	store.log.Trace("Removing store")

	var result *multierror.Error

	if err := store.close(); err != nil {
		result = multierror.Append(result, errors.Wrap(err, "failed to close store"))
	}

	if err := RemoveStore(store.currentEvents, store.filePath, store.user.ID()); err != nil {
		result = multierror.Append(result, errors.Wrap(err, "failed to remove store"))
	}

	if err := store.RemoveCache(); err != nil {
		result = multierror.Append(result, errors.Wrap(err, "failed to remove cache"))
	}

	return result.ErrorOrNil()
}

func (store *Store) RemoveCache() error {
	store.stopWatcher()

	if err := store.clearCachePassphrase(); err != nil {
		logrus.WithError(err).Error("Failed to clear cache passphrase")
	}

	return store.cache.Delete(store.user.ID())
}

// RemoveStore removes the database file and clears the cache file.
func RemoveStore(currentEvents *Events, path, userID string) error {
	var result *multierror.Error

	if err := currentEvents.clearUserEvents(userID); err != nil {
		result = multierror.Append(result, errors.Wrap(err, "failed to clear event loop user cache"))
	}

	// RemoveAll will not return an error if the path does not exist.
	if err := os.RemoveAll(path); err != nil {
		result = multierror.Append(result, errors.Wrap(err, "failed to remove database file"))
	}

	return result.ErrorOrNil()
}
