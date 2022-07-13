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

package pmapi

import (
	"context"
	"errors"
	"strconv"

	"github.com/go-resty/resty/v2"
)

type LabelTypeV4 int

const (
	LabelTypeV4Label        = 1
	LabelTypeV4ContactGroup = 2
	LabelTypeV4Folder       = 3
)

func (c *client) ListLabelsOnly(ctx context.Context) (labels []*Label, err error) {
	return c.listLabelTypeV4(ctx, LabelTypeV4Label)
}

func (c *client) ListFoldersOnly(ctx context.Context) (labels []*Label, err error) {
	return c.listLabelTypeV4(ctx, LabelTypeV4Folder)
}

// listLabelType lists all labels created by the user.
func (c *client) listLabelTypeV4(ctx context.Context, labelType LabelTypeV4) (labels []*Label, err error) {
	var res struct {
		Labels []*Label
	}

	if _, err := c.do(ctx, func(r *resty.Request) (*resty.Response, error) {
		return r.SetQueryParam("Type", strconv.Itoa(int(labelType))).SetResult(&res).Get("/core/v4/labels")
	}); err != nil {
		return nil, err
	}

	return res.Labels, nil
}

// CreateLabel creates a new label.
func (c *client) CreateLabelV4(ctx context.Context, label *Label) (created *Label, err error) {
	if label.Name == "" {
		return nil, errors.New("name is required")
	}

	var res struct {
		Label *Label
	}

	if _, err := c.do(ctx, func(r *resty.Request) (*resty.Response, error) {
		return r.SetBody(&LabelReq{
			Label: label,
		}).SetResult(&res).Post("/core/v4/labels")
	}); err != nil {
		return nil, err
	}

	return res.Label, nil
}

// UpdateLabel updates a label.
func (c *client) UpdateLabelV4(ctx context.Context, label *Label) (updated *Label, err error) {
	if label.Name == "" {
		return nil, errors.New("name is required")
	}

	var res struct {
		Label *Label
	}

	if _, err := c.do(ctx, func(r *resty.Request) (*resty.Response, error) {
		return r.SetBody(&LabelReq{
			Label: label,
		}).SetResult(&res).Put("/core/v4/labels/" + label.ID)
	}); err != nil {
		return nil, err
	}

	return res.Label, nil
}

// DeleteLabel deletes a label.
func (c *client) DeleteLabelV4(ctx context.Context, labelID string) error {
	if _, err := c.do(ctx, func(r *resty.Request) (*resty.Response, error) {
		return r.Delete("/core/v4/labels/" + labelID)
	}); err != nil {
		return err
	}

	return nil
}
