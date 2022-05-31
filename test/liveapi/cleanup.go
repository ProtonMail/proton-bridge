// Copyright (c) 2022 Proton AG
//
// This file is part of Proton Mail Bridge.Bridge.
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

package liveapi

import (
	"context"
	"time"

	"github.com/ProtonMail/proton-bridge/v2/pkg/pmapi"
	"github.com/pkg/errors"
)

func cleanup(client pmapi.Client, addresses *pmapi.AddressList) error {
	if err := cleanSystemFolders(client); err != nil {
		return errors.Wrap(err, "failed to clean system folders")
	}
	if err := cleanCustomLables(client); err != nil {
		return errors.Wrap(err, "failed to clean custom labels")
	}
	if err := cleanTrash(client); err != nil {
		return errors.Wrap(err, "failed to clean trash")
	}
	if err := reorderAddresses(client, addresses); err != nil {
		return errors.Wrap(err, "failed to clean trash")
	}
	return nil
}

func cleanSystemFolders(client pmapi.Client) error {
	for _, labelID := range []string{pmapi.InboxLabel, pmapi.SentLabel, pmapi.ArchiveLabel, pmapi.AllMailLabel, pmapi.DraftLabel} {
		for {
			messages, total, err := client.ListMessages(context.Background(), &pmapi.MessagesFilter{
				PageSize: 150,
				LabelID:  labelID,
			})
			if err != nil {
				return errors.Wrap(err, "failed to list messages")
			}

			if total == 0 {
				break
			}

			messageIDs := []string{}
			for _, message := range messages {
				messageIDs = append(messageIDs, message.ID)
			}

			if err := client.DeleteMessages(context.Background(), messageIDs); err != nil {
				return errors.Wrap(err, "failed to delete messages")
			}

			if total == len(messages) {
				break
			}
		}
	}
	return nil
}

func cleanCustomLables(client pmapi.Client) error {
	labels, err := client.ListLabels(context.Background())
	if err != nil {
		return errors.Wrap(err, "failed to list labels")
	}

	for _, label := range labels {
		if err := emptyFolder(client, label.ID); err != nil {
			return errors.Wrap(err, "failed to empty label")
		}
		if err := client.DeleteLabel(context.Background(), label.ID); err != nil {
			return errors.Wrap(err, "failed to delete label")
		}
	}

	return nil
}

func cleanTrash(client pmapi.Client) error {
	for {
		_, total, err := client.ListMessages(context.Background(), &pmapi.MessagesFilter{
			PageSize: 1,
			LabelID:  pmapi.TrashLabel,
		})
		if err == nil && total == 0 {
			break
		}

		err = emptyFolder(client, pmapi.TrashLabel)
		if err == nil {
			break
		}
		if err.Error() == "Folder or label is currently being emptied" {
			time.Sleep(500 * time.Millisecond)
			continue
		}
		return errors.Wrap(err, "failed to empty trash")
	}
	return nil
}

func emptyFolder(client pmapi.Client, labelID string) error {
	err := client.EmptyFolder(context.Background(), labelID, "")
	if err != nil {
		return err
	}
	for {
		_, total, err := client.ListMessages(context.Background(), &pmapi.MessagesFilter{
			PageSize: 1,
			LabelID:  labelID,
		})
		if err != nil {
			return err
		}
		if total == 0 {
			break
		}
		time.Sleep(1 * time.Second)
	}
	return nil
}

func reorderAddresses(client pmapi.Client, addresses *pmapi.AddressList) error {
	addressIDs := []string{}

	for _, address := range *addresses {
		addressIDs = append(addressIDs, address.ID)
	}

	return client.ReorderAddresses(context.Background(), addressIDs)
}
