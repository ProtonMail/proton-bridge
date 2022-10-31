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

// System labels.
const (
	InboxLabel     = "0"
	AllDraftsLabel = "1"
	AllSentLabel   = "2"
	TrashLabel     = "3"
	SpamLabel      = "4"
	AllMailLabel   = "5"
	ArchiveLabel   = "6"
	SentLabel      = "7"
	DraftLabel     = "8"
	StarredLabel   = "10"

	LabelTypeMailBox      = 1
	LabelTypeContactGroup = 2
)

// IsSystemLabel checks if a label is a pre-defined system label.
func IsSystemLabel(label string) bool {
	switch label {
	case InboxLabel, DraftLabel, SentLabel, TrashLabel, SpamLabel, ArchiveLabel, StarredLabel, AllMailLabel, AllSentLabel, AllDraftsLabel:
		return true
	}
	return false
}

// LabelColors provides the RGB values of the available label colors.
var LabelColors = []string{ //nolint:gochecknoglobals
	"#7272a7",
	"#cf5858",
	"#c26cc7",
	"#7569d1",
	"#69a9d1",
	"#5ec7b7",
	"#72bb75",
	"#c3d261",
	"#e6c04c",
	"#e6984c",
	"#8989ac",
	"#cf7e7e",
	"#c793ca",
	"#9b94d1",
	"#a8c4d5",
	"#97c9c1",
	"#9db99f",
	"#c6cd97",
	"#e7d292",
	"#dfb286",
}

// Label for message.
type Label struct { //nolint:maligned
	ID        string
	Name      string
	Path      string
	Color     string
	Order     int `json:",omitempty"`
	Display   int // Not used for now, leave it empty.
	Exclusive Boolean
	Type      int
	Notify    Boolean
}

func (c *client) ListLabels(ctx context.Context) (labels []*Label, err error) {
	return c.listLabelType(ctx, LabelTypeMailBox)
}

func (c *client) ListContactGroups(ctx context.Context) (labels []*Label, err error) {
	return c.listLabelType(ctx, LabelTypeContactGroup)
}

// listLabelType lists all labels created by the user.
func (c *client) listLabelType(ctx context.Context, labelType int) (labels []*Label, err error) {
	var res struct {
		Labels []*Label
	}

	if _, err := c.do(ctx, func(r *resty.Request) (*resty.Response, error) {
		return r.SetQueryParam("Type", strconv.Itoa(labelType)).SetResult(&res).Get("/labels")
	}); err != nil {
		return nil, err
	}

	return res.Labels, nil
}

type LabelReq struct {
	*Label
}

// CreateLabel creates a new label.
func (c *client) CreateLabel(ctx context.Context, label *Label) (created *Label, err error) {
	if label.Name == "" {
		return nil, errors.New("name is required")
	}

	var res struct {
		Label *Label
	}

	if _, err := c.do(ctx, func(r *resty.Request) (*resty.Response, error) {
		return r.SetBody(&LabelReq{
			Label: label,
		}).SetResult(&res).Post("/labels")
	}); err != nil {
		return nil, err
	}

	return res.Label, nil
}

// UpdateLabel updates a label.
func (c *client) UpdateLabel(ctx context.Context, label *Label) (updated *Label, err error) {
	if label.Name == "" {
		return nil, errors.New("name is required")
	}

	var res struct {
		Label *Label
	}

	if _, err := c.do(ctx, func(r *resty.Request) (*resty.Response, error) {
		return r.SetBody(&LabelReq{
			Label: label,
		}).SetResult(&res).Put("/labels/" + label.ID)
	}); err != nil {
		return nil, err
	}

	return res.Label, nil
}

// DeleteLabel deletes a label.
func (c *client) DeleteLabel(ctx context.Context, labelID string) error {
	if _, err := c.do(ctx, func(r *resty.Request) (*resty.Response, error) {
		return r.Delete("/labels/" + labelID)
	}); err != nil {
		return err
	}

	return nil
}

// LeastUsedColor is intended to return color for creating a new inbox or label.
func LeastUsedColor(colors []string) (color string) {
	color = LabelColors[0]
	frequency := map[string]int{}

	for _, c := range colors {
		frequency[c]++
	}

	for _, c := range LabelColors {
		if frequency[color] > frequency[c] {
			color = c
		}
	}

	return
}
