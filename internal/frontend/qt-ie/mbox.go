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

// +build !nogui

package qtie

import (
	qtcommon "github.com/ProtonMail/proton-bridge/internal/frontend/qt-common"
	"github.com/ProtonMail/proton-bridge/internal/transfer"
	"github.com/sirupsen/logrus"
	"github.com/therecipe/qt/core"
)

// MboxList is an interface between QML and targets for given rule.
type MboxList struct {
	core.QAbstractListModel

	containsFolders bool // Provides only folders if true. On the other hand provides only labels if false
	transfer        *transfer.Transfer
	rule            *transfer.Rule
	log             *logrus.Entry

	_ int `property:"selectedIndex"`

	_ func() `constructor:"init"`
}

func init() {
	// This is needed so the type exists in QML files.
	MboxList_QRegisterMetaType()
}

func newMboxList(t *TransferRules, rule *transfer.Rule, containsFolders bool) *MboxList {
	m := NewMboxList(t)
	m.BeginResetModel()
	m.transfer = t.transfer
	m.rule = rule
	m.containsFolders = containsFolders
	m.log = log.
		WithField("rule", m.rule.SourceMailbox.Hash()).
		WithField("folders", m.containsFolders)
	m.EndResetModel()
	m.itemsChanged(rule)
	return m
}

func (m *MboxList) init() {
	m.ConnectRowCount(m.rowCount)
	m.ConnectRoleNames(m.roleNames)
	m.ConnectData(m.data)
}

func (m *MboxList) rowCount(index *core.QModelIndex) int {
	return len(m.targetMailboxes())
}

func (m *MboxList) roleNames() map[int]*core.QByteArray {
	m.log.
		WithField("isActive", MboxIsActive).
		WithField("id", MboxID).
		WithField("color", MboxColor).
		Debug("role names")
	return map[int]*core.QByteArray{
		MboxIsActive: qtcommon.NewQByteArrayFromString("isActive"),
		MboxID:       qtcommon.NewQByteArrayFromString("mboxID"),
		MboxName:     qtcommon.NewQByteArrayFromString("name"),
		MboxType:     qtcommon.NewQByteArrayFromString("type"),
		MboxColor:    qtcommon.NewQByteArrayFromString("iconColor"),
	}
}

func (m *MboxList) data(index *core.QModelIndex, role int) *core.QVariant {
	allTargets := m.targetMailboxes()

	i, valid := index.Row(), index.IsValid()
	l := m.log.WithField("row", i).WithField("role", role)
	l.Trace("called data()")

	if !valid || i >= len(allTargets) {
		l.WithField("row", i).Warning("Invalid index")
		return core.NewQVariant()
	}

	if m.transfer == nil {
		l.Warning("Requested mbox list data before transfer is connected")
		return qtcommon.NewQVariantString("")
	}

	mbox := allTargets[i]

	switch role {

	case MboxIsActive:
		for _, selectedMailbox := range m.rule.TargetMailboxes {
			if selectedMailbox.Hash() == mbox.Hash() {
				return qtcommon.NewQVariantBool(true)
			}
		}
		return qtcommon.NewQVariantBool(false)

	case MboxID:
		return qtcommon.NewQVariantString(mbox.Hash())

	case MboxName, int(core.Qt__DisplayRole):
		return qtcommon.NewQVariantString(mbox.Name)

	case MboxType:
		t := "label"
		if mbox.IsExclusive {
			t = "folder"
		}
		return qtcommon.NewQVariantString(t)

	case MboxColor:
		return qtcommon.NewQVariantString(mbox.Color)

	default:
		l.Error("Requested mbox list data with unknown role")
		return qtcommon.NewQVariantString("")
	}
}

func (m *MboxList) targetMailboxes() []transfer.Mailbox {
	if m.transfer == nil {
		m.log.Warning("Requested target mailboxes before transfer is connected")
	}

	mailboxes, err := m.transfer.TargetMailboxes()
	if err != nil {
		m.log.WithError(err).Error("Unable to get target mailboxes")
	}

	return m.filter(mailboxes)
}

func (m *MboxList) filter(mailboxes []transfer.Mailbox) (filtered []transfer.Mailbox) {
	for _, mailbox := range mailboxes {
		if mailbox.IsExclusive == m.containsFolders {
			filtered = append(filtered, mailbox)
		}
	}
	return
}

func (m *MboxList) itemsChanged(rule *transfer.Rule) {
	m.rule = rule
	allTargets := m.targetMailboxes()
	l := m.log.WithField("count", len(allTargets))
	l.Trace("called itemChanged()")
	defer func() {
		l.WithField("selected", m.SelectedIndex()).Trace("index updated")
	}()

	// NOTE: Be careful with indices: If they are invalid the DataChanged
	// signal will not be sent to QML e.g. `end == rowCount - 1`
	if len(allTargets) > 0 {
		begin := m.Index(0, 0, core.NewQModelIndex())
		end := m.Index(len(allTargets)-1, 0, core.NewQModelIndex())
		changedRoles := []int{MboxIsActive}
		m.DataChanged(begin, end, changedRoles)
	}

	for index, targetMailbox := range allTargets {
		for _, selectedTarget := range m.rule.TargetMailboxes {
			if targetMailbox.Hash() == selectedTarget.Hash() {
				m.SetSelectedIndex(index)
				return
			}
		}
	}
	m.SetSelectedIndex(-1)
}
