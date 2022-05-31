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
	"context"
	"math"
	"sync"

	"github.com/ProtonMail/proton-bridge/v2/pkg/pmapi"
	"github.com/pkg/errors"
)

const (
	syncMinPagesPerWorker  = 10
	syncMessagesMaxWorkers = 5
	maxFilterPageSize      = 150
)

type storeSynchronizer interface {
	getAllMessageIDs() ([]string, error)
	createOrUpdateMessagesEvent([]*pmapi.Message) error
	deleteMessagesEvent([]string) error
	saveSyncState(finishTime int64, idRanges []*syncIDRange, idsToBeDeleted []string)
}

type messageLister interface {
	ListMessages(context.Context, *pmapi.MessagesFilter) ([]*pmapi.Message, int, error)
}

func syncAllMail(panicHandler PanicHandler, store storeSynchronizer, api messageLister, syncState *syncState) error {
	labelID := pmapi.AllMailLabel

	// When the full sync starts (i.e. is not already in progress), we need to load
	//  - all message IDs in database, so we can see which messages we need to remove at the end of the sync
	//  - ID ranges which indicate how to split work into multiple workers
	if !syncState.isIncomplete() {
		if err := syncState.loadMessageIDsToBeDeleted(); err != nil {
			return errors.Wrap(err, "failed to load message IDs")
		}

		if err := findIDRanges(labelID, api, syncState); err != nil {
			return errors.Wrap(err, "failed to load IDs ranges")
		}
		syncState.save()
	}

	wg := &sync.WaitGroup{}

	shouldStop := 0 // Using integer to have it atomic.
	var resultError error

	for _, idRange := range syncState.idRanges {
		wg.Add(1)
		idRange := idRange // Bind for goroutine.
		go func() {
			defer panicHandler.HandlePanic()
			defer wg.Done()

			err := syncBatch(labelID, store, api, syncState, idRange, &shouldStop)
			if err != nil {
				shouldStop = 1
				resultError = errors.Wrap(err, "failed to sync group")
			}
		}()
	}

	wg.Wait()

	if resultError == nil {
		if err := syncState.deleteMessagesToBeDeleted(); err != nil {
			return errors.Wrap(err, "failed to delete messages")
		}
	}

	return resultError
}

func findIDRanges(labelID string, api messageLister, syncState *syncState) error {
	_, count, err := getSplitIDAndCount(labelID, api, 0)
	if err != nil {
		return errors.Wrap(err, "failed to get first ID and count")
	}
	log.WithField("total", count).Debug("Finding ID ranges")
	if count == 0 {
		return nil
	}

	syncState.initIDRanges()

	pages := int(math.Ceil(float64(count) / float64(maxFilterPageSize)))
	workers := (pages / syncMinPagesPerWorker) + 1
	if workers > syncMessagesMaxWorkers {
		workers = syncMessagesMaxWorkers
	}

	if workers == 1 {
		return nil
	}

	step := int(math.Round(float64(pages) / float64(workers)))
	// Increment steps in case there are more steps than max # of workers (due to rounding).
	if (step*syncMessagesMaxWorkers)+1 < pages {
		step++
	}

	for page := step; page < pages; page += step {
		splitID, _, err := getSplitIDAndCount(labelID, api, page)
		if err != nil {
			return errors.Wrap(err, "failed to get IDs range")
		}
		// Some messages were probably deleted and so the page does not exist anymore.
		// Would be good to start this function again, but let's rather start the sync instead of
		// wasting time of many calls to API to find where to split workers.
		if splitID == "" {
			break
		}
		syncState.addIDRange(splitID)
	}

	return nil
}

func getSplitIDAndCount(labelID string, api messageLister, page int) (string, int, error) {
	sort := "ID"
	desc := false
	filter := &pmapi.MessagesFilter{
		LabelID:  labelID,
		Sort:     sort,
		Desc:     &desc,
		PageSize: maxFilterPageSize,
		Page:     page,
		Limit:    1,
	}
	// If the page does not exist, an empty page instead of an error is returned.
	messages, total, err := api.ListMessages(context.Background(), filter)
	if err != nil {
		return "", 0, errors.Wrap(err, "failed to list messages")
	}
	if len(messages) == 0 {
		return "", 0, nil
	}
	return messages[0].ID, total, nil
}

func syncBatch( //nolint:funlen
	labelID string,
	store storeSynchronizer,
	api messageLister,
	syncState *syncState,
	idRange *syncIDRange,
	shouldStop *int,
) error {
	log.WithField("start", idRange.StartID).WithField("stop", idRange.StopID).Info("Starting sync batch")
	for {
		if *shouldStop == 1 || idRange.isFinished() {
			break
		}

		sort := "ID"
		desc := true
		filter := &pmapi.MessagesFilter{
			LabelID:  labelID,
			Sort:     sort,
			Desc:     &desc,
			PageSize: maxFilterPageSize,
			Page:     0,

			// Messages with BeginID and EndID are included. We will process
			// those messages twice, but that's OK.
			// When message is completely removed, it still works as expected.
			BeginID: idRange.StartID,
			EndID:   idRange.StopID,
		}

		log.WithField("begin", filter.BeginID).WithField("end", filter.EndID).Debug("Fetching page")

		messages, _, err := api.ListMessages(context.Background(), filter)
		if err != nil {
			return errors.Wrap(err, "failed to list messages")
		}

		if len(messages) == 0 {
			break
		}

		for _, m := range messages {
			syncState.doNotDeleteMessageID(m.ID)
		}
		syncState.save()

		if err := store.createOrUpdateMessagesEvent(messages); err != nil {
			return errors.Wrap(err, "failed to create or update messages")
		}

		pageLastMessageID := messages[len(messages)-1].ID
		if !desc {
			idRange.setStartID(pageLastMessageID)
		} else {
			idRange.setStopID(pageLastMessageID)
		}

		if len(messages) < maxFilterPageSize {
			break
		}
	}
	return nil
}
