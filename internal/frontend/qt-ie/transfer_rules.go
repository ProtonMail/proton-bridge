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
	"github.com/therecipe/qt/core"
)

// TransferRules is an interface between QML and transfer.
type TransferRules struct {
	core.QAbstractListModel

	transfer *transfer.Transfer

	targetFoldersCache map[string]*MboxList
	targetLabelsCache  map[string]*MboxList

	_ func() `constructor:"init"`

	_ func(sourceID string) *MboxList                     `slot:"targetFolders,auto"`
	_ func(sourceID string) *MboxList                     `slot:"targetLabels,auto"`
	_ func(sourceID string, isActive bool)                `slot:"setIsRuleActive,auto"`
	_ func(groupName string, isActive bool)               `slot:"setIsGroupActive,auto"`
	_ func(sourceID string, fromDate int64, toDate int64) `slot:"setFromToDate,auto"`
	_ func(sourceID string, targetID string)              `slot:"addTargetID,auto"`
	_ func(sourceID string, targetID string)              `slot:"removeTargetID,auto"`

	// globalFromDate and globalToDate is just default value for GUI, always zero.
	_ int  `property:"globalFromDate"`
	_ int  `property:"globalToDate"`
	_ bool `property:"isLabelGroupSelected"`
	_ bool `property:"isFolderGroupSelected"`
}

func init() {
	// This is needed so the type exists in QML files.
	TransferRules_QRegisterMetaType()
}

func (t *TransferRules) init() {
	log.Trace("Initializing transfer rules")

	t.targetFoldersCache = make(map[string]*MboxList)
	t.targetLabelsCache = make(map[string]*MboxList)

	t.SetGlobalFromDate(0)
	t.SetGlobalToDate(0)

	t.ConnectRowCount(t.rowCount)
	t.ConnectRoleNames(t.roleNames)
	t.ConnectData(t.data)
}

func (t *TransferRules) rowCount(index *core.QModelIndex) int {
	if t.transfer == nil {
		return 0
	}
	return len(t.transfer.GetRules())
}

func (t *TransferRules) roleNames() map[int]*core.QByteArray {
	return map[int]*core.QByteArray{
		MboxIsActive:          qtcommon.NewQByteArrayFromString("isActive"),
		MboxID:                qtcommon.NewQByteArrayFromString("mboxID"),
		MboxName:              qtcommon.NewQByteArrayFromString("name"),
		MboxType:              qtcommon.NewQByteArrayFromString("type"),
		MboxColor:             qtcommon.NewQByteArrayFromString("iconColor"),
		RuleTargetLabelColors: qtcommon.NewQByteArrayFromString("labelColors"),
		RuleFromDate:          qtcommon.NewQByteArrayFromString("fromDate"),
		RuleToDate:            qtcommon.NewQByteArrayFromString("toDate"),
	}
}

func (t *TransferRules) data(index *core.QModelIndex, role int) *core.QVariant {
	i := index.Row()
	allRules := t.transfer.GetRules()

	log := log.WithField("row", i).WithField("role", role)
	log.Trace("Transfer rules data")

	if i >= len(allRules) {
		log.Warning("Invalid index")
		return core.NewQVariant()
	}

	if t.transfer == nil {
		log.Warning("Requested transfer rules data before transfer is connected")
		return qtcommon.NewQVariantString("")
	}

	rule := allRules[i]

	switch role {
	case MboxIsActive:
		return qtcommon.NewQVariantBool(rule.Active)

	case MboxID:
		return qtcommon.NewQVariantString(rule.SourceMailbox.Hash())

	case MboxName:
		return qtcommon.NewQVariantString(rule.SourceMailbox.Name)

	case MboxType:
		if rule.SourceMailbox.IsSystemFolder() {
			return qtcommon.NewQVariantString(FolderTypeSystem)
		}
		if rule.SourceMailbox.IsExclusive {
			return qtcommon.NewQVariantString(FolderTypeFolder)
		}
		return qtcommon.NewQVariantString(FolderTypeLabel)

	case MboxColor:
		return qtcommon.NewQVariantString(rule.SourceMailbox.Color)

	case RuleTargetLabelColors:
		colors := ""
		for _, m := range rule.TargetMailboxes {
			if m.IsExclusive {
				continue
			}
			if colors != "" {
				colors += ";"
			}
			colors += m.Color
		}
		return qtcommon.NewQVariantString(colors)

	case RuleFromDate:
		return qtcommon.NewQVariantLong(rule.FromTime)

	case RuleToDate:
		return qtcommon.NewQVariantLong(rule.ToTime)

	default:
		log.Error("Requested transfer rules data with unknown role")
		return qtcommon.NewQVariantString("")
	}
}

func (t *TransferRules) setTransfer(transfer *transfer.Transfer) {
	log.Debug("Setting transfer")
	t.BeginResetModel()
	defer t.EndResetModel()

	t.transfer = transfer

	t.targetFoldersCache = make(map[string]*MboxList)
	t.targetLabelsCache = make(map[string]*MboxList)

	t.updateGroupSelection()
}

// Getters

func (t *TransferRules) targetFolders(sourceID string) *MboxList {
	rule := t.getRule(sourceID)
	if rule == nil {
		return nil
	}

	if t.targetFoldersCache[sourceID] == nil {
		log.WithField("source", sourceID).Debug("New target folder")
		t.targetFoldersCache[sourceID] = newMboxList(t, rule, true)
	}

	return t.targetFoldersCache[sourceID]
}

func (t *TransferRules) targetLabels(sourceID string) *MboxList {
	rule := t.getRule(sourceID)
	if rule == nil {
		return nil
	}

	if t.targetLabelsCache[sourceID] == nil {
		log.WithField("source", sourceID).Debug("New target label")
		t.targetLabelsCache[sourceID] = newMboxList(t, rule, false)
	}

	return t.targetLabelsCache[sourceID]
}

// Setters

func (t *TransferRules) setIsGroupActive(groupName string, isActive bool) {
	log.WithField("group", groupName).WithField("active", isActive).Trace("Setting group as active/inactive")

	wantExclusive := (groupName == FolderTypeFolder)
	for _, rule := range t.transfer.GetRules() {
		if rule.SourceMailbox.IsExclusive != wantExclusive {
			continue
		}
		if rule.SourceMailbox.IsSystemFolder() {
			continue
		}
		if rule.Active != isActive {
			t.setIsRuleActive(rule.SourceMailbox.Hash(), isActive)
		}
	}
}

func (t *TransferRules) setIsRuleActive(sourceID string, isActive bool) {
	log.WithField("source", sourceID).WithField("active", isActive).Trace("Setting rule as active/inactive")

	rule := t.getRule(sourceID)
	if rule == nil {
		return
	}
	if isActive {
		t.setRule(rule.SourceMailbox, rule.TargetMailboxes, rule.FromTime, rule.ToTime, []int{MboxIsActive})
	} else {
		t.unsetRule(rule.SourceMailbox)
	}
}

func (t *TransferRules) setFromToDate(sourceID string, fromTime int64, toTime int64) {
	log.WithField("source", sourceID).WithField("fromTime", fromTime).WithField("toTime", toTime).Trace("Setting from and to dates")

	if sourceID == "-1" {
		t.transfer.SetGlobalTimeLimit(fromTime, toTime)
		return
	}

	rule := t.getRule(sourceID)
	if rule == nil {
		return
	}
	t.setRule(rule.SourceMailbox, rule.TargetMailboxes, fromTime, toTime, []int{RuleFromDate, RuleToDate})
}

func (t *TransferRules) addTargetID(sourceID string, targetID string) {
	log.WithField("source", sourceID).WithField("target", targetID).Trace("Adding target")

	rule := t.getRule(sourceID)
	if rule == nil {
		return
	}
	targetMailboxToAdd := t.getMailbox(t.transfer.TargetMailboxes, targetID)
	if targetMailboxToAdd == nil {
		return
	}

	newTargetMailboxes := []transfer.Mailbox{}
	found := false
	for _, targetMailbox := range rule.TargetMailboxes {
		if targetMailbox.Hash() == targetMailboxToAdd.Hash() {
			found = true
		}
		if !targetMailboxToAdd.IsExclusive || (targetMailboxToAdd.IsExclusive && !targetMailbox.IsExclusive) {
			newTargetMailboxes = append(newTargetMailboxes, targetMailbox)
		}
	}
	if !found {
		newTargetMailboxes = append(newTargetMailboxes, *targetMailboxToAdd)
	}
	t.setRule(rule.SourceMailbox, newTargetMailboxes, rule.FromTime, rule.ToTime, []int{RuleTargetLabelColors})
	t.updateTargetSelection(sourceID, targetMailboxToAdd.IsExclusive)
}

func (t *TransferRules) removeTargetID(sourceID string, targetID string) {
	log.WithField("source", sourceID).WithField("target", targetID).Trace("Removing target")

	rule := t.getRule(sourceID)
	if rule == nil {
		return
	}
	targetMailboxToRemove := t.getMailbox(t.transfer.TargetMailboxes, targetID)
	if targetMailboxToRemove == nil {
		return
	}

	newTargetMailboxes := []transfer.Mailbox{}
	for _, targetMailbox := range rule.TargetMailboxes {
		if targetMailbox.Hash() != targetMailboxToRemove.Hash() {
			newTargetMailboxes = append(newTargetMailboxes, targetMailbox)
		}
	}
	t.setRule(rule.SourceMailbox, newTargetMailboxes, rule.FromTime, rule.ToTime, []int{RuleTargetLabelColors})
	t.updateTargetSelection(sourceID, targetMailboxToRemove.IsExclusive)
}

// Helpers

// getRule returns rule for given source ID.
// WARN: Always get new rule after change because previous pointer points to
// outdated struct with old data.
func (t *TransferRules) getRule(sourceID string) *transfer.Rule {
	mailbox := t.getMailbox(t.transfer.SourceMailboxes, sourceID)
	if mailbox == nil {
		return nil
	}
	return t.transfer.GetRule(*mailbox)
}

func (t *TransferRules) getMailbox(mailboxesGetter func() ([]transfer.Mailbox, error), sourceID string) *transfer.Mailbox {
	if t.transfer == nil {
		log.Warn("Getting mailbox without avaiable transfer")
		return nil
	}

	mailboxes, err := mailboxesGetter()
	if err != nil {
		log.WithError(err).Error("Failed to get source mailboxes")
		return nil
	}
	for _, mailbox := range mailboxes {
		if mailbox.Hash() == sourceID {
			return &mailbox
		}
	}
	log.WithField("source", sourceID).Error("Mailbox not found for source")
	return nil
}

func (t *TransferRules) setRule(sourceMailbox transfer.Mailbox, targetMailboxes []transfer.Mailbox, fromTime, toTime int64, changedRoles []int) {
	if err := t.transfer.SetRule(sourceMailbox, targetMailboxes, fromTime, toTime); err != nil {
		log.WithError(err).WithField("source", sourceMailbox.Hash()).Error("Failed to set rule")
	}
	t.ruleChanged(sourceMailbox, changedRoles)
}

func (t *TransferRules) unsetRule(sourceMailbox transfer.Mailbox) {
	t.transfer.UnsetRule(sourceMailbox)
	t.ruleChanged(sourceMailbox, []int{MboxIsActive})
}

func (t *TransferRules) ruleChanged(sourceMailbox transfer.Mailbox, changedRoles []int) {
	allRules := t.transfer.GetRules()
	for row, rule := range allRules {
		if rule.SourceMailbox.Hash() != sourceMailbox.Hash() {
			continue
		}

		index := t.Index(row, 0, core.NewQModelIndex())
		if !index.IsValid() || row >= len(allRules) {
			log.WithField("row", row).Warning("Invalid index")
			return
		}

		log.WithField("row", row).Trace("Transfer rule changed")
		t.DataChanged(index, index, changedRoles)
		break
	}

	t.updateGroupSelection()
}

func (t *TransferRules) updateGroupSelection() {
	areAllLabelsSelected, areAllFoldersSelected := true, true
	for _, rule := range t.transfer.GetRules() {
		if rule.Active {
			continue
		}
		if rule.SourceMailbox.IsSystemFolder() {
			continue
		}
		if rule.SourceMailbox.IsExclusive {
			areAllFoldersSelected = false
		} else {
			areAllLabelsSelected = false
		}

		if !areAllLabelsSelected && !areAllFoldersSelected {
			break
		}
	}

	t.SetIsLabelGroupSelected(areAllLabelsSelected)
	t.SetIsFolderGroupSelected(areAllFoldersSelected)
}

func (t *TransferRules) updateTargetSelection(sourceID string, updateFolderSelect bool) {
	rule := t.getRule(sourceID)
	if rule == nil {
		return
	}

	if updateFolderSelect {
		t.targetFolders(rule.SourceMailbox.Hash()).itemsChanged(rule)
	} else {
		t.targetLabels(rule.SourceMailbox.Hash()).itemsChanged(rule)
	}
}
