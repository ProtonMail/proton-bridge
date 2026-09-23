// Copyright (c) 2026 Proton AG
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
// along with Proton Mail Bridge.  If not, see <https://www.gnu.org/licenses/>.

package imapservice

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sync"

	"github.com/ProtonMail/gluon/reporter"
	"github.com/ProtonMail/proton-bridge/v3/internal/services/syncservice"
	"github.com/ProtonMail/proton-bridge/v3/pkg/atomicfile"
	"github.com/bradenaw/juniper/xmaps"
	"github.com/sirupsen/logrus"
)

type SyncState struct {
	filePath       string
	status         syncservice.Status
	lock           sync.Mutex
	sentryReporter reporter.Reporter
}

var ErrInvalidSyncFileVersion = errors.New("invalid sync file version")

const SyncFileVersion = 1

type syncStateFile struct {
	Version int
	Data    string
}

type syncFileVersion1 struct {
	Status syncservice.Status
}

func NewSyncState(filePath string, reporter reporter.Reporter) (*SyncState, error) {
	s := &SyncState{filePath: filePath, status: syncservice.DefaultStatus(), sentryReporter: reporter}

	if err := s.loadUnsafe(); err != nil {
		return nil, err
	}

	return s, nil
}

func (s *SyncState) AddFailedMessageID(_ context.Context, ids ...string) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	count := len(s.status.FailedMessages)

	for _, id := range ids {
		s.status.FailedMessages.Add(id)
	}

	// Only update if something change.
	if count == len(s.status.FailedMessages) {
		return nil
	}

	return s.storeUnsafe()
}

func (s *SyncState) RemFailedMessageID(_ context.Context, ids ...string) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	count := len(s.status.FailedMessages)

	for _, id := range ids {
		s.status.FailedMessages.Remove(id)
	}

	// Only update if something change.
	if count == len(s.status.FailedMessages) {
		return nil
	}

	return s.storeUnsafe()
}

func (s *SyncState) GetSyncStatus(_ context.Context) (syncservice.Status, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	return s.status, nil
}

func (s *SyncState) ClearSyncStatus(_ context.Context) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	oldStatus := s.status

	s.status = syncservice.DefaultStatus()

	if err := s.storeUnsafe(); err != nil {
		s.status = oldStatus
		return err
	}

	return nil
}

func (s *SyncState) SetHasLabels(_ context.Context, b bool) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	if s.status.HasLabels == b {
		return nil
	}

	s.status.HasLabels = b

	return s.storeUnsafe()
}

func (s *SyncState) SetHasMessages(_ context.Context, b bool) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	if s.status.HasMessages == b {
		return nil
	}

	s.status.HasMessages = b

	return s.storeUnsafe()
}

func (s *SyncState) SetLastMessageID(_ context.Context, s2 string, i int64) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.status.LastSyncedMessageID = s2
	s.status.NumSyncedMessages += i

	return s.storeUnsafe()
}

func (s *SyncState) SetMessageCount(_ context.Context, i int64) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	if s.status.TotalMessageCount == i && s.status.HasMessageCount {
		return nil
	}

	s.status.TotalMessageCount = i
	s.status.HasMessageCount = true

	return s.storeUnsafe()
}

func (s *SyncState) SetStartSyncEventID(_ context.Context, id string) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	if s.status.StartSyncEventID == id {
		return nil
	}

	s.status.StartSyncEventID = id

	return s.storeUnsafe()
}

func (s *SyncState) storeUnsafe() error {
	if err := storeImpl(&s.status, s.filePath); err != nil {
		_ = s.sentryReporter.ReportMessageWithContext(
			"Failed to store sync file state",
			reporter.Context{
				"err": err,
			},
		)

		return err
	}

	return nil
}

func storeImpl(status *syncservice.Status, path string) error {
	data, err := json.Marshal(syncFileVersion1{Status: *status})
	if err != nil {
		return fmt.Errorf("failed to marshal sync state data: %w", err)
	}

	syncFile := syncStateFile{
		Version: SyncFileVersion,
		Data:    string(data),
	}

	syncFileData, err := json.Marshal(syncFile)
	if err != nil {
		return fmt.Errorf("failed to marshal sync state file: %w", err)
	}

	if err := atomicfile.WriteFile(path, syncFileData); err != nil {
		return fmt.Errorf("failed to update sync file state: %w", err)
	}

	return nil
}

func (s *SyncState) loadUnsafe() error {
	data, err := os.ReadFile(s.filePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}

		_ = s.sentryReporter.ReportMessageWithContext(
			"Failed to read sync file",
			reporter.Context{
				"err": err,
			},
		)

		return err
	}

	var syncFile syncStateFile

	if err := json.Unmarshal(data, &syncFile); err != nil {
		_ = s.sentryReporter.ReportMessageWithContext(
			"Failed to unmarshal sync file",
			reporter.Context{
				"err": err,
			},
		)
		logrus.WithError(err).Error("failed to unmarshal sync file")
		return s.resetUnsafe()
	}

	if syncFile.Version != SyncFileVersion {
		return ErrInvalidSyncFileVersion
	}

	var v1 syncFileVersion1

	if err := json.Unmarshal([]byte(syncFile.Data), &v1); err != nil {
		logrus.WithError(err).Error("failed to unmarshal sync data")
		return s.resetUnsafe()
	}

	s.status = v1.Status

	return nil
}

// resetUnsafe resets the sync state and the syncState file to the default state.
func (s *SyncState) resetUnsafe() error {
	s.status = syncservice.DefaultStatus()

	return s.storeUnsafe()
}

func DeleteSyncState(configDir, userID string) error {
	path := GetSyncConfigPath(configDir, userID)

	if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}

	return nil
}

func MigrateVaultSettings(
	configDir, userID string,
	hasLabels, hasMessages bool,
	failedMessageIDs []string,
) (bool, error) {
	filePath := GetSyncConfigPath(configDir, userID)

	_, err := os.ReadFile(filePath) //nolint:gosec
	if err == nil {
		// File already exists, sync has been migrated.
		return false, nil
	}

	if err != nil && !errors.Is(err, os.ErrNotExist) {
		// unexpected error occurred.
		return false, err
	}

	status := syncservice.DefaultStatus()
	status.HasLabels = hasLabels
	status.HasMessages = hasMessages
	status.HasMessageCount = hasMessages
	status.FailedMessages = xmaps.SetFromSlice(failedMessageIDs)

	return true, storeImpl(&status, filePath)
}
