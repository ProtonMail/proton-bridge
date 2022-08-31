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

package fakeapi

import (
	"context"
	"fmt"
	"strings"

	"github.com/ProtonMail/proton-bridge/v2/pkg/pmapi"
)

const (
	route   = "/labels"
	routeV4 = "/core/v4/labels"
)

func (api *FakePMAPI) isLabelFolder(labelID string) bool {
	for _, label := range api.labels {
		if label.ID == labelID {
			return bool(label.Exclusive)
		}
	}
	switch labelID {
	case pmapi.InboxLabel,
		pmapi.TrashLabel,
		pmapi.SpamLabel,
		pmapi.ArchiveLabel,
		pmapi.SentLabel,
		pmapi.DraftLabel:
		return true
	}
	return false
}

func (api *FakePMAPI) ListLabels(ctx context.Context) ([]*pmapi.Label, error) {
	return api.listLabels(ctx, "1", route)
}

func (api *FakePMAPI) CreateLabel(ctx context.Context, label *pmapi.Label) (*pmapi.Label, error) {
	return api.createLabel(ctx, label, route)
}

func (api *FakePMAPI) UpdateLabel(ctx context.Context, label *pmapi.Label) (*pmapi.Label, error) {
	return api.updateLabel(ctx, label, route)
}

func (api *FakePMAPI) DeleteLabel(ctx context.Context, labelID string) error {
	return api.deleteLabel(ctx, labelID, route)
}

func (api *FakePMAPI) ListLabelsOnly(ctx context.Context) ([]*pmapi.Label, error) {
	return api.listLabels(ctx, "1", routeV4)
}

func (api *FakePMAPI) ListFoldersOnly(ctx context.Context) ([]*pmapi.Label, error) {
	return api.listLabels(ctx, "3", routeV4)
}

func (api *FakePMAPI) CreateLabelV4(ctx context.Context, label *pmapi.Label) (*pmapi.Label, error) {
	return api.createLabel(ctx, label, routeV4)
}

func (api *FakePMAPI) UpdateLabelV4(ctx context.Context, label *pmapi.Label) (*pmapi.Label, error) {
	return api.updateLabel(ctx, label, routeV4)
}

func (api *FakePMAPI) DeleteLabelV4(ctx context.Context, labelID string) error {
	return api.deleteLabel(ctx, labelID, routeV4)
}

func (api *FakePMAPI) listLabels(_ context.Context, labeType string, route string) ([]*pmapi.Label, error) {
	if err := api.checkAndRecordCall(GET, route+"/"+labeType, nil); err != nil {
		return nil, err
	}
	return append([]*pmapi.Label{}, api.labels...), nil
}

func (api *FakePMAPI) createLabel(_ context.Context, label *pmapi.Label, route string) (*pmapi.Label, error) {
	if err := api.checkAndRecordCall(POST, route, &pmapi.LabelReq{Label: label}); err != nil {
		return nil, err
	}

	// API blocks certain names
	switch strings.ToLower(label.Name) {
	case "inbox", "drafts", "trash", "spam", "starred":
		return nil, fmt.Errorf("Invalid name") //nolint:stylecheck
	}

	for _, existingLabel := range api.labels {
		if existingLabel.Name == label.Name {
			return nil, fmt.Errorf("A label or folder with this name already exists") //nolint:stylecheck
		}
	}
	prefix := "label"
	if label.Exclusive {
		prefix = "folder"
	}
	label.ID = api.controller.labelIDGenerator.next(prefix)
	if label.Path == "" {
		label.Path = label.Name
	}
	api.labels = append(api.labels, label)
	api.addEventLabel(pmapi.EventCreate, label)
	return label, nil
}

func (api *FakePMAPI) updateLabel(_ context.Context, label *pmapi.Label, route string) (*pmapi.Label, error) {
	if err := api.checkAndRecordCall(PUT, route, &pmapi.LabelReq{Label: label}); err != nil {
		return nil, err
	}
	for idx, existingLabel := range api.labels {
		if existingLabel.ID == label.ID {
			// Request doesn't have to include all properties and these have to stay the same.
			label.Type = existingLabel.Type
			label.Exclusive = existingLabel.Exclusive
			if label.Path == "" {
				label.Path = label.Name
			}
			api.labels[idx] = label
			api.addEventLabel(pmapi.EventUpdate, label)
			return label, nil
		}
	}
	return nil, fmt.Errorf("label %s does not exist", label.ID)
}

func (api *FakePMAPI) deleteLabel(_ context.Context, labelID string, route string) error {
	if err := api.checkAndRecordCall(DELETE, route+labelID, nil); err != nil {
		return err
	}
	for idx, existingLabel := range api.labels {
		if existingLabel.ID == labelID {
			api.labels = append(api.labels[:idx], api.labels[idx+1:]...)
			api.addEventLabel(pmapi.EventDelete, existingLabel)
			return nil
		}
	}
	return fmt.Errorf("label %s does not exist", labelID)
}
