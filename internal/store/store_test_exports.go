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

package store

import (
	"encoding/json"
	"fmt"

	"github.com/ProtonMail/proton-bridge/v2/pkg/pmapi"
	"github.com/stretchr/testify/assert"
	bolt "go.etcd.io/bbolt"
)

func (loop *eventLoop) IsRunning() bool {
	return loop.isRunning
}

// TestSync triggers a sync of the store.
func (store *Store) TestSync() {
	store.lock.Lock()
	defer store.lock.Unlock()

	// Sync can happen any time. Sync assigns sequence numbers and UIDs
	// in the order of fetching from the server. We expect in the test
	// that sequence numbers and UIDs are assigned in the same order as
	// written in scenario setup. With more than one sync that cannot
	// be guaranteed so once test calls this function, first it has to
	// delete previous any already synced sequence numbers and UIDs.
	_ = store.truncateMailboxesBucket()

	store.triggerSync()
}

// TestPollNow triggers a loop of the event loop.
func (store *Store) TestPollNow() {
	store.eventLoop.pollNow()
}

// TestIsSyncRunning returns whether the sync is currently ongoing.
func (store *Store) TestIsSyncRunning() bool {
	return store.isSyncRunning
}

// TestGetEventLoop returns the store's event loop.
func (store *Store) TestGetEventLoop() *eventLoop { //nolint:revive
	return store.eventLoop
}

// TestGetLastEvent returns last event processed by the store's event loop.
func (store *Store) TestGetLastEvent() *pmapi.Event {
	return store.eventLoop.currentEvent
}

// TestGetStoreFilePath returns the filepath of the store's database file.
func (store *Store) TestGetStoreFilePath() string {
	return store.filePath
}

// TestDumpDB will dump store database content.
func (store *Store) TestDumpDB(tb assert.TestingT) {
	if store == nil || store.db == nil {
		fmt.Printf(">>>>>>>> NIL STORE / DB <<<<<\n\n")
		assert.Fail(tb, "store or database is nil")
		return
	}

	dumpCounts := true
	fmt.Printf(">>>>>>>> DUMP %s <<<<<\n\n", store.db.Path())

	txMails := txDumpMailsFactory(tb)

	txDump := func(tx *bolt.Tx) error {
		if dumpCounts {
			if err := txDumpCounts(tx); err != nil {
				return err
			}
		}
		return txMails(tx)
	}

	assert.NoError(tb, store.db.View(txDump))
}

func txDumpMailsFactory(tb assert.TestingT) func(tx *bolt.Tx) error {
	return func(tx *bolt.Tx) error {
		mailboxes := tx.Bucket(mailboxesBucket)
		metadata := tx.Bucket(metadataBucket)
		err := mailboxes.ForEach(func(mboxName, mboxData []byte) error {
			fmt.Println("mbox:", string(mboxName))
			b := mailboxes.Bucket(mboxName).Bucket(imapIDsBucket)
			deletedMailboxes := mailboxes.Bucket(mboxName).Bucket(deletedIDsBucket)
			c := b.Cursor()
			i := 0
			for imapID, apiID := c.First(); imapID != nil; imapID, apiID = c.Next() {
				i++
				isDeleted := deletedMailboxes != nil && deletedMailboxes.Get(apiID) != nil
				fmt.Println("  ", i, "imap", btoi(imapID), "api", string(apiID), "isDeleted", isDeleted)
				data := metadata.Get(apiID)
				if !assert.NotNil(tb, data) {
					continue
				}
				if !assert.NoError(tb, txMailMeta(data, i)) {
					continue
				}
			}
			fmt.Println("total:", i)
			return nil
		})
		return err
	}
}

func txDumpCounts(tx *bolt.Tx) error {
	counts := tx.Bucket(countsBucket)
	err := counts.ForEach(func(labelID, countsB []byte) error {
		defer fmt.Println()
		fmt.Printf("counts id: %q ", string(labelID))
		counts := &mailboxCounts{}
		if err := json.Unmarshal(countsB, counts); err != nil {
			fmt.Printf(" Error %v", err)
			return nil
		}
		fmt.Printf(" total :%d unread %d", counts.TotalOnAPI, counts.UnreadOnAPI)
		return nil
	})
	return err
}

func txMailMeta(data []byte, i int) error {
	fullMetaDump := false
	msg := &pmapi.Message{}
	if err := json.Unmarshal(data, msg); err != nil {
		return err
	}
	if msg.Body != "" {
		fmt.Printf("   %d body %s\n\n", i, msg.Body)
		panic("NONZERO BODY")
	}
	if i >= 10 {
		return nil
	}
	if fullMetaDump {
		fmt.Printf("   %d meta %s\n\n", i, string(data))
	} else {
		fmt.Println(
			"     Subj", msg.Subject,
			"\n     From", msg.Sender,
			"\n     Time", msg.Time,
			"\n     Labels", msg.LabelIDs,
			"\n     Unread", msg.Unread,
		)
	}

	return nil
}
