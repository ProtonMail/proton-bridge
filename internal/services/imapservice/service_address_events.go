// Copyright (c) 2025 Proton AG
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
	"fmt"

	"github.com/ProtonMail/go-proton-api"
	"github.com/ProtonMail/proton-bridge/v3/internal/services/useridentity"
	"github.com/ProtonMail/proton-bridge/v3/internal/usertypes"
	"github.com/sirupsen/logrus"
)

func (s *Service) HandleAddressEvents(ctx context.Context, events []proton.AddressEvent) error {
	s.log.Debug("handling address event")

	if s.addressMode == usertypes.AddressModeCombined {
		oldPrimaryAddr, err := s.identityState.GetPrimaryAddress()
		if err != nil {
			return fmt.Errorf("failed to get primary addr: %w", err)
		}

		if err := s.identityState.Write(func(identity *useridentity.State) error {
			identity.OnAddressEvents(events)
			return nil
		}); err != nil {
			s.log.WithError(err).Error("Failed to apply address events to identity state")
			return err
		}

		newPrimaryAddr, err := s.identityState.GetPrimaryAddress()
		if err != nil {
			return fmt.Errorf("failed to get primary addr after update: %w", err)
		}

		if oldPrimaryAddr.ID == newPrimaryAddr.ID {
			return nil
		}

		connector, ok := s.connectors[oldPrimaryAddr.ID]
		if !ok {
			return fmt.Errorf("could not find old primary addr conncetor after default address change")
		}

		s.connectors[newPrimaryAddr.ID] = connector
		delete(s.connectors, oldPrimaryAddr.ID)

		s.log.WithFields(logrus.Fields{
			"old": oldPrimaryAddr.Email,
			"new": newPrimaryAddr.Email,
		}).Debug("Primary address changed")

		return nil
	}

	for _, event := range events {
		switch event.Action {
		case proton.EventCreate:
			var status useridentity.AddressUpdate

			if err := s.identityState.Write(func(identity *useridentity.State) error {
				status = identity.OnAddressCreated(event)

				return nil
			}); err != nil {
				return fmt.Errorf("failed to update identity state: %w", err)
			}

			if status == useridentity.AddressUpdateCreated {
				if err := addNewAddressSplitMode(ctx, s, event.Address.ID); err != nil {
					return err
				}
			}

		case proton.EventUpdateFlags, proton.EventUpdate:
			var addr proton.Address
			var status useridentity.AddressUpdate

			if err := s.identityState.Write(func(identity *useridentity.State) error {
				addr, status = identity.OnAddressUpdated(event)

				return nil
			}); err != nil {
				return fmt.Errorf("failed to update identity state: %w", err)
			}

			// nolint:exhaustive
			switch status {
			case useridentity.AddressUpdateCreated:
				if err := addNewAddressSplitMode(ctx, s, addr.ID); err != nil {
					return err
				}

			case useridentity.AddressUpdateDisabled:
				if err := removeAddressSplitMode(ctx, s, addr.ID); err != nil {
					return err
				}

			case useridentity.AddressUpdateEnabled:
				if err := addNewAddressSplitMode(ctx, s, addr.ID); err != nil {
					return err
				}

			default:
				continue
			}

		case proton.EventDelete:
			var status useridentity.AddressUpdate

			if err := s.identityState.Write(func(identity *useridentity.State) error {
				_, status = identity.OnAddressDeleted(event)

				return nil
			}); err != nil {
				return fmt.Errorf("failed to update identity state: %w", err)
			}
			if status == useridentity.AddressUpdateDeleted {
				if err := removeAddressSplitMode(ctx, s, event.ID); err != nil {
					return err
				}
			}

		default:
			return fmt.Errorf("unknown event action: %v", event.Action)
		}
	}

	return nil
}

func addNewAddressSplitMode(ctx context.Context, s *Service, addrID string) error {
	connector := NewConnector(
		addrID,
		s.client,
		s.labels,
		s.identityState,
		s.addressMode,
		s.sendRecorder,
		s.panicHandler,
		s.reporter,
		s.showAllMail,
		s.syncStateProvider,
	)

	if err := s.serverManager.AddIMAPUser(ctx, connector, connector.addrID, s.gluonIDProvider, s.syncStateProvider); err != nil {
		return fmt.Errorf("failed to add new account to server: %w", err)
	}

	s.connectors[connector.addrID] = connector

	updates, err := syncLabels(ctx, s.labels.GetLabelMap(), []*Connector{connector}, s.labelConflictManager)
	if err != nil {
		return fmt.Errorf("failed to create labels updates for new address: %w", err)
	}

	if err := waitOnIMAPUpdates(ctx, updates); err != nil {
		return fmt.Errorf("failed to sync labels for new address: %w", err)
	}

	return nil
}

func removeAddressSplitMode(ctx context.Context, s *Service, addrID string) error {
	connector, ok := s.connectors[addrID]
	if !ok {
		s.log.Warnf("Could not find connector ")
		return nil
	}

	if err := s.serverManager.RemoveIMAPUser(ctx, true, s.gluonIDProvider, addrID); err != nil {
		return fmt.Errorf("failed to remove user from server: %w", err)
	}

	connector.StateClose()

	delete(s.connectors, addrID)

	return nil
}
