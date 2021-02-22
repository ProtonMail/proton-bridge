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
	"testing"
)

func hasNumberOfLabels(tb testing.TB, folder *FolderInfo, expected int) {
	if current := len(folder.TargetLabelIDList()); current != expected {
		tb.Error("Folder has wrong number of labels. Expected", expected, "has", current, " labels", folder.TargetLabelIDs)
	}
}

func labelStringEquals(tb testing.TB, folder *FolderInfo, expected string) {
	if current := folder.TargetLabelIDs; current != expected {
		tb.Error("Folder returned wrong labels. Expected", expected, "has", current, " labels", folder.TargetLabelIDs)
	}
}

func TestLabelInfoUniqSet(t *testing.T) {
	folder := &FolderInfo{}
	labelStringEquals(t, folder, "")
	hasNumberOfLabels(t, folder, 0)
	// add label
	folder.AddTargetLabel("blah")
	hasNumberOfLabels(t, folder, 1)
	labelStringEquals(t, folder, "blah")
	//
	folder.AddTargetLabel("blah___")
	hasNumberOfLabels(t, folder, 2)
	labelStringEquals(t, folder, "blah;blah___")
	// add same label
	folder.AddTargetLabel("blah")
	hasNumberOfLabels(t, folder, 2)
	// remove label
	folder.RemoveTargetLabel("blah")
	hasNumberOfLabels(t, folder, 1)
	//
	folder.AddTargetLabel("blah___")
	hasNumberOfLabels(t, folder, 1)
	// remove same label
	folder.RemoveTargetLabel("blah")
	hasNumberOfLabels(t, folder, 1)
	// add again label
	folder.AddTargetLabel("blah")
	hasNumberOfLabels(t, folder, 2)
}
