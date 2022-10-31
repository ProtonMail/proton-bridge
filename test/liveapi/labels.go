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
	"fmt"
	"strings"

	"github.com/ProtonMail/proton-bridge/v2/pkg/pmapi"
	"github.com/pkg/errors"
)

var systemLabelNameToID = map[string]string{ //nolint:gochecknoglobals
	"INBOX":    pmapi.InboxLabel,
	"Trash":    pmapi.TrashLabel,
	"Spam":     pmapi.SpamLabel,
	"All Mail": pmapi.AllMailLabel,
	"Archive":  pmapi.ArchiveLabel,
	"Sent":     pmapi.SentLabel,
	"Drafts":   pmapi.DraftLabel,
}

func (ctl *Controller) AddUserLabel(username string, label *pmapi.Label) error {
	client, err := getPersistentClient(username)
	if err != nil {
		return err
	}

	label.Exclusive = getLabelExclusive(label.Name)
	label.Name = getLabelNameWithoutPrefix(label.Name)
	label.Color = pmapi.LabelColors[0]
	if _, err := client.CreateLabel(context.Background(), label); err != nil {
		return errors.Wrap(err, "failed to create label")
	}
	return nil
}

func (ctl *Controller) GetLabelIDs(username string, labelNames []string) ([]string, error) {
	labelIDs := []string{}
	for _, labelName := range labelNames {
		labelID, err := ctl.getLabelID(username, labelName)
		if err != nil {
			return nil, err
		}
		labelIDs = append(labelIDs, labelID)
	}
	return labelIDs, nil
}

func (ctl *Controller) getLabelID(username, labelName string) (string, error) {
	if labelID, ok := systemLabelNameToID[labelName]; ok {
		return labelID, nil
	}

	client, err := getPersistentClient(username)
	if err != nil {
		return "", err
	}

	labels, err := client.ListLabels(context.Background())
	if err != nil {
		return "", errors.Wrap(err, "failed to list labels")
	}

	exclusive := getLabelExclusive(labelName)
	labelName = getLabelNameWithoutPrefix(labelName)
	for _, label := range labels {
		if label.Exclusive == exclusive && label.Name == labelName {
			return label.ID, nil
		}
	}
	return "", fmt.Errorf("label %s:%s does not exist", username, labelName)
}

func getLabelNameWithoutPrefix(name string) string {
	if strings.HasPrefix(name, "Folders/") {
		return strings.TrimPrefix(name, "Folders/")
	}
	if strings.HasPrefix(name, "Labels/") {
		return strings.TrimPrefix(name, "Labels/")
	}
	return name
}

func getLabelExclusive(name string) pmapi.Boolean {
	return pmapi.Boolean(strings.HasPrefix(name, "Folders/"))
}
