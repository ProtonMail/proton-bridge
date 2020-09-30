// Copyright (c) 2020 Proton Technologies AG
//
// This file is part of ProtonMail Bridge.
//
// ProtonMail Bridge is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// ProtonMail Bridge is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with ProtonMail Bridge.  If not, see <https://www.gnu.org/licenses/>.

package pmapi

import "fmt"

// System labels
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

	LabelTypeMailbox      = 1
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
var LabelColors = []string{ //nolint[gochecknoglobals]
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

type LabelAction int

const (
	RemoveLabel LabelAction = iota
	AddLabel
)

// Label for message.
type Label struct {
	ID        string
	Name      string
	Path      string
	Color     string
	Order     int `json:",omitempty"`
	Display   int // Not used for now, leave it empty.
	Exclusive int
	Type      int
	Notify    int
}

type LabelListRes struct {
	Res
	Labels []*Label
}

func (c *client) ListLabels() (labels []*Label, err error) {
	return c.ListLabelType(LabelTypeMailbox)
}

func (c *client) ListContactGroups() (labels []*Label, err error) {
	return c.ListLabelType(LabelTypeContactGroup)
}

// ListLabelType lists all labels created by the user.
func (c *client) ListLabelType(labelType int) (labels []*Label, err error) {
	req, err := c.NewRequest("GET", fmt.Sprintf("/labels?%d", labelType), nil)
	if err != nil {
		return
	}

	var res LabelListRes
	if err = c.DoJSON(req, &res); err != nil {
		return
	}

	labels, err = res.Labels, res.Err()
	return
}

type LabelReq struct {
	*Label
}

type LabelRes struct {
	Res
	Label *Label
}

// CreateLabel creates a new label.
func (c *client) CreateLabel(label *Label) (created *Label, err error) {
	labelReq := &LabelReq{label}
	req, err := c.NewJSONRequest("POST", "/labels", labelReq)
	if err != nil {
		return
	}

	var res LabelRes
	if err = c.DoJSON(req, &res); err != nil {
		return
	}

	created, err = res.Label, res.Err()
	return
}

// UpdateLabel updates a label.
func (c *client) UpdateLabel(label *Label) (updated *Label, err error) {
	labelReq := &LabelReq{label}
	req, err := c.NewJSONRequest("PUT", "/labels/"+label.ID, labelReq)
	if err != nil {
		return
	}

	var res LabelRes
	if err = c.DoJSON(req, &res); err != nil {
		return
	}

	updated, err = res.Label, res.Err()
	return
}

// DeleteLabel deletes a label.
func (c *client) DeleteLabel(id string) (err error) {
	req, err := c.NewRequest("DELETE", "/labels/"+id, nil)
	if err != nil {
		return
	}

	var res Res
	if err = c.DoJSON(req, &res); err != nil {
		return
	}

	err = res.Err()
	return
}

// LeastUsedColor is intended to return color for creating a new inbox or label
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
