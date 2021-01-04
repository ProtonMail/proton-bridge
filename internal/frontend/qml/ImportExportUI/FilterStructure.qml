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

// Filter only selected folders or labels
import QtQuick 2.8
import QtQml.Models 2.2


DelegateModel {
    id: root
    model         : structurePM
    //filterOnGroup : root.folderType
    //delegate      : root.delegate
    groups        : [
        DelegateModelGroup {name: gui.enums.folderTypeFolder ; includeByDefault: false},
        DelegateModelGroup {name: gui.enums.folderTypeLabel  ; includeByDefault: false}
    ]

    function updateFilter() {
        //console.log("FilterModelDelegate::UpdateFilter")
        // filter
        var rowCount = root.items.count;
        for (var iItem = 0; iItem < rowCount; iItem++) {
            var entry = root.items.get(iItem);
            entry.inLabel = (
                root.filterOnGroup        == gui.enums.folderTypeLabel &&
                entry.model.folderType == gui.enums.folderTypeLabel
            )
            entry.inFolder = (
                root.filterOnGroup        == gui.enums.folderTypeFolder &&
                entry.model.folderType != gui.enums.folderTypeLabel
            )
            /*
             if (entry.inFolder && entry.model.folderId == selectedIDs) {
                 view.currentIndex = iItem
             }
             */
            //console.log("::::update filter:::::", iItem, entry.model.folderName, entry.inFolder, entry.inLabel)
        }
    }
}
