// Copyright (c) 2021 Proton Technologies AG
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

// +build build_qt

package qtie

import (
	"errors"
	"strings"

	qtcommon "github.com/ProtonMail/proton-bridge/internal/frontend/qt-common"
	"github.com/ProtonMail/proton-bridge/internal/transfer"
	"github.com/ProtonMail/proton-bridge/pkg/pmapi"
	"github.com/therecipe/qt/core"
)

const (
	GlobalOptionIndex = -1
)

var AllFolderInfoRoles = []int{
	FolderId,
	FolderName,
	FolderColor,
	FolderType,
	FolderEntries,
	IsFolderSelected,
	FolderFromDate,
	FolderToDate,
	TargetFolderID,
	TargetLabelIDs,
}

func getTargetHashes(mboxes []transfer.Mailbox) (targetFolderID, targetLabelIDs string) {
	for _, targetMailbox := range mboxes {
		if targetMailbox.IsExclusive {
			targetFolderID = targetMailbox.Hash()
		} else {
			targetLabelIDs += targetMailbox.Hash() + ";"
		}
	}

	targetLabelIDs = strings.Trim(targetLabelIDs, ";")
	return
}

func newFolderInfo(mbox transfer.Mailbox, rule *transfer.Rule) *FolderInfo {
	targetFolderID, targetLabelIDs := getTargetHashes(rule.TargetMailboxes)

	entry := &FolderInfo{
		mailbox:          mbox,
		FolderEntries:    1,
		FromDate:         rule.FromTime,
		ToDate:           rule.ToTime,
		IsFolderSelected: rule.Active,
		TargetFolderID:   targetFolderID,
		TargetLabelIDs:   targetLabelIDs,
	}

	entry.FolderType = FolderTypeSystem
	if !pmapi.IsSystemLabel(mbox.ID) {
		if mbox.IsExclusive {
			entry.FolderType = FolderTypeFolder
		} else {
			entry.FolderType = FolderTypeLabel
		}
	}

	return entry
}

func (s *FolderStructure) saveRule(info *FolderInfo) error {
	if s.transfer == nil {
		return errors.New("missing transfer")
	}
	sourceMbox := info.mailbox
	if !info.IsFolderSelected {
		s.transfer.UnsetRule(sourceMbox)
		return nil
	}
	allTargetMboxes, err := s.transfer.TargetMailboxes()
	if err != nil {
		return err
	}
	var targetMboxes []transfer.Mailbox
	for _, target := range allTargetMboxes {
		targetHash := target.Hash()
		if info.TargetFolderID == targetHash || strings.Contains(info.TargetLabelIDs, targetHash) {
			targetMboxes = append(targetMboxes, target)
		}
	}

	return s.transfer.SetRule(sourceMbox, targetMboxes, info.FromDate, info.ToDate)
}

func (s *FolderInfo) updateTargetLabelIDs(targetLabelsSet map[string]struct{}) {
	targets := []string{}
	for key := range targetLabelsSet {
		targets = append(targets, key)
	}
	s.TargetLabelIDs = strings.Join(targets, ";")
}

func (s *FolderInfo) AddTargetLabel(targetID string) {
	if targetID == "" {
		return
	}
	targetLabelsSet := s.getSetOfLabels()
	targetLabelsSet[targetID] = struct{}{}
	s.updateTargetLabelIDs(targetLabelsSet)
}

func (s *FolderInfo) RemoveTargetLabel(targetID string) {
	if targetID == "" {
		return
	}
	targetLabelsSet := s.getSetOfLabels()
	delete(targetLabelsSet, targetID)
	s.updateTargetLabelIDs(targetLabelsSet)
}

func (s *FolderInfo) IsType(askType string) bool {
	return s.FolderType == askType
}

func (s *FolderInfo) getSetOfLabels() (uniqSet map[string]struct{}) {
	uniqSet = make(map[string]struct{})
	for _, label := range s.TargetLabelIDList() {
		uniqSet[label] = struct{}{}
	}
	return
}

func (s *FolderInfo) TargetLabelIDList() []string {
	return strings.FieldsFunc(
		s.TargetLabelIDs,
		func(c rune) bool { return c == ';' },
	)
}

// Get data
func (s *FolderStructure) data(index *core.QModelIndex, role int) *core.QVariant {
	row, isValid := index.Row(), index.IsValid()
	if !isValid || row >= s.getCount() {
		log.Warnln("Wrong index", isValid, row)
		return core.NewQVariant()
	}

	var f = s.get(row)

	switch role {
	case FolderId:
		return qtcommon.NewQVariantString(f.mailbox.Hash())
	case FolderName, int(core.Qt__DisplayRole):
		return qtcommon.NewQVariantString(f.mailbox.Name)
	case FolderColor:
		return qtcommon.NewQVariantString(f.mailbox.Color)
	case FolderType:
		return qtcommon.NewQVariantString(f.FolderType)
	case FolderEntries:
		return qtcommon.NewQVariantInt(f.FolderEntries)
	case FolderFromDate:
		return qtcommon.NewQVariantLong(f.FromDate)
	case FolderToDate:
		return qtcommon.NewQVariantLong(f.ToDate)
	case IsFolderSelected:
		return qtcommon.NewQVariantBool(f.IsFolderSelected)
	case TargetFolderID:
		return qtcommon.NewQVariantString(f.TargetFolderID)
	case TargetLabelIDs:
		return qtcommon.NewQVariantString(f.TargetLabelIDs)
	default:
		log.Warnln("Wrong role", role)
		return core.NewQVariant()
	}
}

// Get header data (table view, tree view)
func (s *FolderStructure) headerData(section int, orientation core.Qt__Orientation, role int) *core.QVariant {
	if role != int(core.Qt__DisplayRole) {
		return core.NewQVariant()
	}

	if orientation == core.Qt__Horizontal {
		return qtcommon.NewQVariantString("Column")
	}

	return qtcommon.NewQVariantString("Row")
}

// Flags is editable
func (s *FolderStructure) flags(index *core.QModelIndex) core.Qt__ItemFlag {
	if !index.IsValid() {
		return core.Qt__ItemIsEnabled
	}

	// can do here also: core.NewQAbstractItemModelFromPointer(s.Pointer()).Flags(index) | core.Qt__ItemIsEditable
	// or s.FlagsDefault(index) | core.Qt__ItemIsEditable
	return core.Qt__ItemIsEnabled | core.Qt__ItemIsSelectable | core.Qt__ItemIsEditable
}

// Set data
func (s *FolderStructure) setData(index *core.QModelIndex, value *core.QVariant, role int) bool {
	log.Debugf("SET DATA %d", role)
	if !index.IsValid() {
		return false
	}
	if index.Row() < GlobalOptionIndex || index.Row() > s.getCount() || index.Column() != 1 {
		return false
	}
	item := s.get(index.Row())
	t := true
	switch role {
	case FolderId, FolderType:
		log.
			WithField("structure", s).
			WithField("row", index.Row()).
			WithField("column", index.Column()).
			WithField("role", role).
			WithField("isEdit", role == int(core.Qt__EditRole)).
			Warn("Set constant role forbiden")
	case FolderName:
		item.mailbox.Name = value.ToString()
	case FolderColor:
		item.mailbox.Color = value.ToString()
	case FolderEntries:
		item.FolderEntries = value.ToInt(&t)
	case FolderFromDate:
		item.FromDate = value.ToLongLong(&t)
	case FolderToDate:
		item.ToDate = value.ToLongLong(&t)
	case IsFolderSelected:
		item.IsFolderSelected = value.ToBool()
	case TargetFolderID:
		item.TargetFolderID = value.ToString()
	case TargetLabelIDs:
		item.TargetLabelIDs = value.ToString()
	default:
		log.Debugln("uknown role ", s, index.Row(), index.Column(), role, role == int(core.Qt__EditRole))
		return false
	}
	s.changedEntityRole(index.Row(), index.Row(), role)
	return true
}

// Dimension of model: number of rows is equivalent to number of items in list
func (s *FolderStructure) rowCount(parent *core.QModelIndex) int {
	return s.getCount()
}

func (s *FolderStructure) getCount() int {
	return len(s.entities)
}

// Returns names of available item properties
func (s *FolderStructure) roleNames() map[int]*core.QByteArray {
	return s.Roles()
}

// Clear removes all items in model
func (s *FolderStructure) Clear() {
	s.BeginResetModel()
	if s.getCount() != 0 {
		s.entities = []*FolderInfo{}
	}

	s.GlobalOptions = FolderInfo{
		mailbox: transfer.Mailbox{
			Name: "=",
		},
		FromDate:       0,
		ToDate:         0,
		TargetFolderID: "",
		TargetLabelIDs: "",
	}
	s.EndResetModel()
}

// Method connected to addEntry slot
func (s *FolderStructure) addEntry(entry *FolderInfo) {
	s.insertEntry(entry, s.getCount())
}

// NewUniqId which is not in map yet.
func (s *FolderStructure) newUniqId() (name string) {
	name = s.GlobalOptions.mailbox.Name
	mbox := transfer.Mailbox{Name: name}
	for newVal := byte(name[0]); true; newVal++ {
		mbox.Name = string([]byte{newVal})
		if s.getRowById(mbox.Hash()) < GlobalOptionIndex {
			return
		}
	}
	return
}

// Method connected to addEntry slot
func (s *FolderStructure) insertEntry(entry *FolderInfo, i int) {
	s.BeginInsertRows(core.NewQModelIndex(), i, i)
	s.entities = append(s.entities[:i], append([]*FolderInfo{entry}, s.entities[i:]...)...)
	s.EndInsertRows()
	// update global if conflict
	if entry.mailbox.Hash() == s.GlobalOptions.mailbox.Hash() {
		globalName := s.newUniqId()
		s.GlobalOptions.mailbox.Name = globalName
	}
}

func (s *FolderStructure) GetInfo(row int) FolderInfo {
	return *s.get(row)
}

func (s *FolderStructure) changedEntityRole(rowStart int, rowEnd int, roles ...int) {
	if rowStart < GlobalOptionIndex || rowEnd < GlobalOptionIndex {
		return
	}
	if rowStart < 0 || rowStart >= s.getCount() {
		rowStart = 0
	}
	if rowEnd < 0 || rowEnd >= s.getCount() {
		rowEnd = s.getCount()
	}
	if rowStart > rowEnd {
		tmp := rowStart
		rowStart = rowEnd
		rowEnd = tmp
	}
	indexStart := s.Index(rowStart, 0, core.NewQModelIndex())
	indexEnd := s.Index(rowEnd, 0, core.NewQModelIndex())
	s.updateSelection(indexStart, indexEnd, roles)
	s.DataChanged(indexStart, indexEnd, roles)
}

func (s *FolderStructure) setFolderSelection(id string, toSelect bool) {
	log.Debugf("set folder selection %q %b", id, toSelect)
	i := s.getRowById(id)
	//
	info := s.get(i)
	before := info.IsFolderSelected
	info.IsFolderSelected = toSelect
	if err := s.saveRule(info); err != nil {
		s.get(i).IsFolderSelected = before
		log.WithError(err).WithField("id", id).WithField("toSelect", toSelect).Error("Cannot set selection")
		return
	}
	//
	s.changedEntityRole(i, i, IsFolderSelected)
}

func (s *FolderStructure) setTargetFolderID(id, target string) {
	log.Debugf("set targetFolderID %q %q", id, target)
	i := s.getRowById(id)
	//
	info := s.get(i)
	//s.get(i).TargetFolderID = target
	before := info.TargetFolderID
	info.TargetFolderID = target
	if err := s.saveRule(info); err != nil {
		info.TargetFolderID = before
		log.WithError(err).WithField("id", id).WithField("target", target).Error("Cannot set target")
		return
	}
	//
	s.changedEntityRole(i, i, TargetFolderID)
	if target == "" { // do not import
		before := info.TargetLabelIDs
		info.TargetLabelIDs = ""
		if err := s.saveRule(info); err != nil {
			info.TargetLabelIDs = before
			log.WithError(err).WithField("id", id).WithField("target", target).Error("Cannot set target")
			return
		}
		s.changedEntityRole(i, i, TargetLabelIDs)
	}
}

func (s *FolderStructure) addTargetLabelID(id, label string) {
	log.Debugf("add target label id %q %q", id, label)
	if label == "" {
		return
	}
	i := s.getRowById(id)
	info := s.get(i)
	before := info.TargetLabelIDs
	info.AddTargetLabel(label)
	if err := s.saveRule(info); err != nil {
		info.TargetLabelIDs = before
		log.WithError(err).WithField("id", id).WithField("label", label).Error("Cannot add label")
		return
	}
	s.changedEntityRole(i, i, TargetLabelIDs)
}

func (s *FolderStructure) removeTargetLabelID(id, label string) {
	log.Debugf("remove label id %q %q", id, label)
	if label == "" {
		return
	}
	i := s.getRowById(id)
	info := s.get(i)
	before := info.TargetLabelIDs
	info.RemoveTargetLabel(label)
	if err := s.saveRule(info); err != nil {
		info.TargetLabelIDs = before
		log.WithError(err).WithField("id", id).WithField("label", label).Error("Cannot remove label")
		return
	}
	s.changedEntityRole(i, i, TargetLabelIDs)
}

func (s *FolderStructure) setFromToDate(id string, from, to int64) {
	log.Debugf("set from to date %q %d %d", id, from, to)
	i := s.getRowById(id)
	info := s.get(i)
	beforeFrom := info.FromDate
	beforeTo := info.ToDate
	info.FromDate = from
	info.ToDate = to
	if err := s.saveRule(info); err != nil {
		info.FromDate = beforeFrom
		info.ToDate = beforeTo
		log.WithError(err).WithField("id", id).WithField("from", from).WithField("to", to).Error("Cannot set date")
		return
	}
	s.changedEntityRole(i, i, FolderFromDate, FolderToDate)
}

func (s *FolderStructure) selectType(folderType string, toSelect bool) {
	log.Debugf("set type %q %b", folderType, toSelect)
	iFirst, iLast := -1, -1
	for i, entity := range s.entities {
		if entity.IsType(folderType) {
			if iFirst == -1 {
				iFirst = i
			}
			before := entity.IsFolderSelected
			entity.IsFolderSelected = toSelect
			if err := s.saveRule(entity); err != nil {
				entity.IsFolderSelected = before
				log.WithError(err).WithField("i", i).WithField("type", folderType).WithField("toSelect", toSelect).Error("Cannot select type")
			}
			iLast = i
		}
	}
	if iFirst != -1 {
		s.changedEntityRole(iFirst, iLast, IsFolderSelected)
	}
}

func (s *FolderStructure) updateSelection(topLeft *core.QModelIndex, bottomRight *core.QModelIndex, roles []int) {
	for _, role := range roles {
		switch role {
		case IsFolderSelected:
			s.SetSelectedFolders(true)
			s.SetSelectedLabels(true)
			s.SetAtLeastOneSelected(false)
			for _, entity := range s.entities {
				if entity.IsFolderSelected {
					s.SetAtLeastOneSelected(true)
				} else {
					if entity.IsType(FolderTypeFolder) {
						s.SetSelectedFolders(false)
					}
					if entity.IsType(FolderTypeLabel) {
						s.SetSelectedLabels(false)
					}
				}
				if !s.IsSelectedFolders() && !s.IsSelectedLabels() && s.IsAtLeastOneSelected() {
					break
				}
			}
		default:
		}
	}
}

func (s *FolderStructure) hasFolderWithName(name string) bool {
	for _, entity := range s.entities {
		if entity.mailbox.Name == name {
			return true
		}
	}
	return false
}

func (s *FolderStructure) getRowById(id string) (row int) {
	for row = GlobalOptionIndex; row < s.getCount(); row++ {
		if id == s.get(row).mailbox.Hash() {
			return
		}
	}
	row = GlobalOptionIndex - 1
	return
}

func (s *FolderStructure) hasTarget() bool {
	for row := 0; row < s.getCount(); row++ {
		if s.get(row).TargetFolderID != "" {
			return true
		}
	}
	return false
}

// Getter for account info pointer
// index out of array length returns empty folder info to avoid segfault
// index == GlobalOptionIndex is set to access global options
func (s *FolderStructure) get(index int) *FolderInfo {
	if index < GlobalOptionIndex || index >= s.getCount() {
		return &FolderInfo{}
	}
	if index == GlobalOptionIndex {
		return &s.GlobalOptions
	}
	return s.entities[index]
}
