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

import QtQuick 2.8
import ImportExportUI 1.0
import ProtonUI 1.0
import QtQuick.Controls 2.1
import QtQuick.Window 2.2

Window {
    id      : testroot
    width   : 100
    height  : 600
    flags   : Qt.Window | Qt.Dialog | Qt.FramelessWindowHint
    visible : true
    title   : "GUI test Window"
    color   : "transparent"
    x       : testgui.winMain.x - 120
    y       : testgui.winMain.y

    property bool newVersion : true

    Accessible.name: testroot.title
    Accessible.description: "Window with buttons testing the GUI events"


    Rectangle {
        id:test_systray
        anchors{
            top: parent.top
            horizontalCenter: parent.horizontalCenter
        }
        height: 40
        width: 100
        color: "yellow"
        Image {
            id: sysImg
            anchors {
                left : test_systray.left
                top  : test_systray.top
            }
            height: test_systray.height
            mipmap: true
            fillMode : Image.PreserveAspectFit
            source: ""
        }
        Text {
            id: systrText
            anchors {
                right : test_systray.right
                verticalCenter: test_systray.verticalCenter
            }
            text: "unset"
        }

        function normal() {
            test_systray.color = "#22ee22"
            systrText.text = "norm"
            sysImg.source= "../share/icons/rounded-systray.png"
        }
        function highlight() {
            test_systray.color = "#eeee22"
            systrText.text = "highl"
            sysImg.source= "../share/icons/rounded-syswarn.png"
        }

        MouseArea {
            property point diff: "0,0"
            anchors.fill: parent
            onPressed: {
                diff = Qt.point(testroot.x, testroot.y)
                var mousePos = mapToGlobal(mouse.x, mouse.y)
                diff.x -= mousePos.x
                diff.y -= mousePos.y
            }
            onPositionChanged: {
                var currPos = mapToGlobal(mouse.x, mouse.y)
                testroot.x = currPos.x + diff.x
                testroot.y = currPos.y + diff.y
            }
        }
    }

    ListModel {
        id: buttons

        ListElement { title : "Show window"        }
        ListElement { title : "Logout cuthix"      }
        ListElement { title : "Internet on"        }
        ListElement { title : "Internet off"       }
        ListElement { title : "Macos"              }
        ListElement { title : "Windows"            }
        ListElement { title : "Linux"              }
        ListElement { title : "New Version"        }
        ListElement { title : "ForceUpgrade"       }
        ListElement { title : "ImportStructure"    }
        ListElement { title : "DraftImpFailed"     }
        ListElement { title : "NoInterImp"         }
        ListElement { title : "ReportImp"          }
        ListElement { title : "NewFolder"          }
        ListElement { title : "EditFolder"         }
        ListElement { title : "EditLabel"          }
        ListElement { title : "ExpProgErr"         }
        ListElement { title : "ImpProgErr"         }
    }

    ListView {
        id: view
        anchors {
            top    : test_systray.bottom
            bottom : parent.bottom
            left   : parent.left
            right  : parent.right
        }

        orientation : ListView.Vertical
        model       : buttons
        focus       : true

        delegate : ButtonRounded {
            text        : title
            color_main  : "orange"
            color_minor : "#aa335588"
            isOpaque    : true
            anchors.horizontalCenter: parent.horizontalCenter
            onClicked : {
                console.log("Clicked on ", title)
                switch (title)  {
                    case "Show window" :
                    go.showWindow();
                    break;
                    case "Logout cuthix" :
                    go.checkLoggedOut("cuthix");
                    break;
                    case "Internet on" :
                    go.setConnectionStatus(true);
                    break;
                    case "Internet off" :
                    go.setConnectionStatus(false);
                    break;
                    case "Macos" :
                    go.goos = "darwin";
                    break;
                    case "Windows" :
                    go.goos = "windows";
                    break;
                    case "Linux" :
                    go.goos = "linux";
                    break;
                    case "New Version" :
                    testroot.newVersion = !testroot.newVersion
                    systrText.text = testroot.newVersion ? "new version" : "uptodate"
                    break
                    case "ForceUpgrade" :
                    go.notifyUpgrade()
                    break;
                    case "ImportStructure" :
                    testgui.winMain.dialogImport.address = "cuto@pm.com"
                    testgui.winMain.dialogImport.show()
                    testgui.winMain.dialogImport.currentIndex=DialogImport.Page.SourceToTarget
                    break
                    case "DraftImpFailed" :
                    testgui.notifyError(testgui.enums.errDraftImportFailed)
                    break
                    case "NoInterImp" :
                    testgui.notifyError(testgui.enums.errNoInternetWhileImport)
                    break
                    case "ReportImp" :
                    testgui.winMain.dialogImport.address = "cuto@pm.com"
                    testgui.winMain.dialogImport.show()
                    testgui.winMain.dialogImport.currentIndex=DialogImport.Page.Report
                    break
                    case "NewFolder" :
                    testgui.winMain.popupFolderEdit.show("currentName", "", "", testgui.enums.folderTypeFolder, "")
                    break
                    case "EditFolder" :
                    testgui.winMain.popupFolderEdit.show("currentName", "", "#7272a7", testgui.enums.folderTypeFolder, "")
                    break
                    case "EditFolder" :
                    testgui.winMain.popupFolderEdit.show("currentName", "", "", testgui.enums.folderTypeFolder, "")
                    break
                    case "ImpProgErr" :
                    go.animateProgressBar.pause()
                    testgui.notifyError(testgui.enums.errEmailImportFailed)
                    break
                    case "ExpProgErr" :
                    go.animateProgressBar.pause()
                    testgui.notifyError(testgui.enums.errEmailExportFailed)
                    break
                    default :
                    console.log("Not implemented " + title)
                }
            }
        }
    }


    Component.onCompleted : {
        testgui.winMain.x = 150
        testgui.winMain.y = 100
    }

    //InstanceExistsWindow { id: ie_test }

    GuiIE {
        id: testgui
        //checkLogTimer.interval: 3*1000
        winMain.visible: true

        ListModel{
            id: accountsModel
            ListElement{ account : "cuthix"                                           ; status : "connected";    isExpanded: false; isCombinedAddressMode: false; hostname : "127.0.0.1"; password : "ZI9tKp+ryaxmbpn2E12"; security : "StarTLS"; portSMTP : 1025; portIMAP : 1143; aliases : "cuto@pm.com;jaku@pm.com;DoYouKnowAboutAMovieCalledTheHorriblySlowMurderWithExtremelyInefficientWeapon@thatYouCanFindForExampleOnyoutube.com" }
            ListElement{ account : "exteremelongnamewhichmustbeeladedinthemiddleoftheaddress@protonmail.com" ; status : "connected";    isExpanded: true;  isCombinedAddressMode: true;  hostname : "127.0.0.1"; password : "ZI9tKp+ryaxmbpn2E12"; security : "StarTLS"; portSMTP : 1025; portIMAP : 1143; aliases : "cuto@pm.com;jaku@pm.com;hu@hu.hu"                                                        }
            ListElement{ account : "cuthix2@protonmail.com"                           ; status : "disconnected"; isExpanded: false; isCombinedAddressMode: false; hostname : "127.0.0.1"; password : "ZI9tKp+ryaxmbpn2E12"; security : "StarTLS"; portSMTP : 1025; portIMAP : 1143; aliases : "cuto@pm.com;jaku@pm.com;hu@hu.hu"                                                        }
            ListElement{ account : "many@protonmail.com" ; status : "connected";    isExpanded: true;  isCombinedAddressMode: true;  hostname : "127.0.0.1"; password : "ZI9tKp+ryaxmbpn2E12"; security : "StarTLS"; portSMTP : 1025; portIMAP : 1143; aliases : "cuto@pm.com;jaku@pm.com;hu@hu.hu;cuto@pm.com;jaku@pm.com;hu@hu.hu;cuto@pm.com;jaku@pm.com;hu@hu.hu;cuto@pm.com;jaku@pm.com;hu@hu.hu;cuto@pm.com;jaku@pm.com;hu@hu.hu;cuto@pm.com;jaku@pm.com;hu@hu.hu;cuto@pm.com;jaku@pm.com;hu@hu.hu;cuto@pm.com;jaku@pm.com;hu@hu.hu;cuto@pm.com;jaku@pm.com;hu@hu.hu;cuto@pm.com;jaku@pm.com;hu@hu.hu;cuto@pm.com;jaku@pm.com;hu@hu.hu;cuto@pm.com;jaku@pm.com;hu@hu.hu;cuto@pm.com;jaku@pm.com;hu@hu.hu;cuto@pm.com;jaku@pm.com;hu@hu.hu;cuto@pm.com;jaku@pm.com;hu@hu.hu;"}
        }

        ListModel{
            id: structureExternal

            property var globalOptions: JSON.parse('{ "folderId" : "global--uniq"  , "folderName" : ""  , "folderColor" : "" , "folderType" : ""  , "folderEntries" : 0, "fromDate": 0, "toDate": 0, "isFolderSelected" : false  , "targetFolderID": "14" , "targetLabelIDs": ";20;29" }')

            ListElement{ folderId : "Inbox" ; folderName : "Inbox" ; folderColor : "black" ; folderType : "" ; folderEntries : 1 ; fromDate : 0 ; toDate : 0 ; isFolderSelected : true  ; targetFolderID : "14" ; targetLabelIDs : ";20;29" }
            ListElement{ folderId : "Sent"  ; folderName : "Sent"  ; folderColor : "black" ; folderType : "" ; folderEntries : 2 ; fromDate : 0 ; toDate : 0 ; isFolderSelected : false ; targetFolderID : ""   ; targetLabelIDs : ""       }
            ListElement{ folderId : "Spam"  ; folderName : "Spam"  ; folderColor : "black" ; folderType : "" ; folderEntries : 3 ; fromDate : 0 ; toDate : 0 ; isFolderSelected : false ; targetFolderID : ""   ; targetLabelIDs : ""       }
            ListElement{ folderId : "Draft" ; folderName : "Draft" ; folderColor : "black" ; folderType : "" ; folderEntries : 4 ; fromDate : 0 ; toDate : 0 ; isFolderSelected : true  ; targetFolderID : "14" ; targetLabelIDs : ";20;29" }

            ListElement{ folderId : "Folder0" ; folderName : "Folder0"                                                                     ; folderColor : "black" ; folderType : "folder" ; folderEntries : 10 ; fromDate : 300000 ; toDate : 15000000 ; isFolderSelected : true  ; targetFolderID : "14" ; targetLabelIDs : ";20;29" }
            ListElement{ folderId : "Folder1" ; folderName : "Folder1"                                                                     ; folderColor : "black" ; folderType : "folder" ; folderEntries : 20 ; fromDate : 300000 ; toDate : 15000000 ; isFolderSelected : false ; targetFolderID : ""   ; targetLabelIDs : ""      }
            ListElement{ folderId : "Folder2" ; folderName : "Folder2"                                                                     ; folderColor : "black" ; folderType : "folder" ; folderEntries : 30 ; fromDate : 300000 ; toDate : 15000000 ; isFolderSelected : true  ; targetFolderID : "14" ; targetLabelIDs : ";20;29" }
            ListElement{ folderId : "Folder3" ; folderName : "Folder3ToolongAndMustBeElidedSimilarToOnOfAccountsItJustNotNeedToBeThatLong" ; folderColor : "black" ; folderType : "folder" ; folderEntries : 40 ; fromDate : 300000 ; toDate : 15000000 ; isFolderSelected : false ; targetFolderID : ""   ; targetLabelIDs : ""      }

            ListElement{ folderId : "Label0" ; folderName : "Label-" ; folderColor : "black" ; folderType : "label" ; folderEntries : 10 ; fromDate : 300000 ; toDate : 15000000 ; isFolderSelected : false ; targetFolderID : ""   ; targetLabelIDs : ""       }
            ListElement{ folderId : "Label1" ; folderName : "Label1" ; folderColor : "black" ; folderType : "label" ; folderEntries : 11 ; fromDate : 300000 ; toDate : 15000000 ; isFolderSelected : true  ; targetFolderID : "14" ; targetLabelIDs : ";20;29" }
            ListElement{ folderId : "Label2" ; folderName : "Label2" ; folderColor : "black" ; folderType : "label" ; folderEntries : 12 ; fromDate : 300000 ; toDate : 15000000 ; isFolderSelected : false ; targetFolderID : ""   ; targetLabelIDs : ""       }
            ListElement{ folderId : "Label3" ; folderName : "Label3" ; folderColor : "black" ; folderType : "label" ; folderEntries : 13 ; fromDate : 300000 ; toDate : 15000000 ; isFolderSelected : true  ; targetFolderID : "14" ; targetLabelIDs : ";20;29" }

            function addTargetLabelID    ( id , label )    { structureFuncs.addTargetLabelID    ( structureExternal , id , label ) }
            function removeTargetLabelID ( id , label )    { structureFuncs.removeTargetLabelID ( structureExternal , id , label ) }
            function setTargetFolderID   ( id , label )    { structureFuncs.setTargetFolderID   ( structureExternal , id , label ) }
            function setFromToDate       ( id , from, to ) { structureFuncs.setFromToDate       ( structureExternal , id , from, to ) }

            function getID             ( row        ) { return row == -1 ? structureExternal.globalOptions.folderId : structureExternal.get(row).folderId }
            function getById           ( folderId   ) { return structureFuncs.getById            ( structureExternal , folderId           ) }
            function getFrom           ( folderId   ) { return structureExternal.getById         ( folderId          ) .fromDate          }
            function getTo             ( folderId   ) { return structureExternal.getById         ( folderId          ) .toDate            }
            function getTargetLabelIDs ( folderId   ) { return structureExternal.getById         ( folderId          ) .getTargetLabelIDs }
            function hasFolderWithName ( folderName ) { return structureFuncs.hasFolderWithName  ( structureExternal , folderName         ) }

            function hasTarget () { return structureFuncs.hasTarget(structureExternal) }
        }

        ListModel{
            id: structurePM

            // group selectors
            property bool selectedLabels  : false
            property bool selectedFolders : false
            property bool atLeastOneSelected : true

            property var globalOptions: JSON.parse('{ "folderId" : "global--uniq"  , "folderName" : "global"  , "folderColor" : "black" , "folderType" : ""  , "folderEntries" : 0 , "fromDate": 300000 , "toDate": 15000000 , "isFolderSelected" : false  , "targetFolderID": "14" , "targetLabelIDs": ";20;29" }')

            ListElement{ folderId : "0" ; folderName : "INBOX" ; folderColor : "blue" ; folderType : "" ; folderEntries : 1 ; isFolderSelected : true  ; targetFolderID : "" ; targetLabelIDs : "" ; }
            ListElement{ folderId : "1" ; folderName : "Sent"  ; folderColor : "blue" ; folderType : "" ; folderEntries : 2 ; isFolderSelected : false ; targetFolderID : "" ; targetLabelIDs : "" ; }
            ListElement{ folderId : "2" ; folderName : "Spam"  ; folderColor : "blue" ; folderType : "" ; folderEntries : 3 ; isFolderSelected : false ; targetFolderID : "" ; targetLabelIDs : "" ; }
            ListElement{ folderId : "3" ; folderName : "Draft" ; folderColor : "blue" ; folderType : "" ; folderEntries : 4 ; isFolderSelected : true  ; targetFolderID : "" ; targetLabelIDs : "" ; }
            ListElement{ folderId : "6" ; folderName : "Archive" ; folderColor : "blue" ; folderType : "" ; folderEntries : 5 ; isFolderSelected : true  ; targetFolderID : "" ; targetLabelIDs : "" ; }

            ListElement{ folderId : "14" ; folderName : "Folder0"                                                                     ; folderColor : "blue"   ; folderType : "folder" ; folderEntries : 10 ; isFolderSelected : true  ; targetFolderID : "" ; targetLabelIDs : "" ; }
            ListElement{ folderId : "15" ; folderName : "Folder1"                                                                     ; folderColor : "green"  ; folderType : "folder" ; folderEntries : 20 ; isFolderSelected : false ; targetFolderID : "" ; targetLabelIDs : "" ; }
            ListElement{ folderId : "16" ; folderName : "Folder2"                                                                     ; folderColor : "pink"   ; folderType : "folder" ; folderEntries : 30 ; isFolderSelected : true  ; targetFolderID : "" ; targetLabelIDs : "" ; }
            ListElement{ folderId : "17" ; folderName : "Folder3ToolongAndMustBeElidedSimilarToOnOfAccountsItJustNotNeedToBeThatLong" ; folderColor : "orange" ; folderType : "folder" ; folderEntries : 40 ; isFolderSelected : true  ; targetFolderID : "" ; targetLabelIDs : "" ; }

            ListElement{ folderId : "28" ; folderName : "Label0"                                                                     ; folderColor : "red"    ; folderType : "label" ; folderEntries : 10 ; isFolderSelected : false ; targetFolderID : "" ; targetLabelIDs : "" ; }
            ListElement{ folderId : "29" ; folderName : "Label1"                                                                     ; folderColor : "blue"   ; folderType : "label" ; folderEntries : 11 ; isFolderSelected : true  ; targetFolderID : "" ; targetLabelIDs : "" ; }
            ListElement{ folderId : "20" ; folderName : "Label2"                                                                     ; folderColor : "green"  ; folderType : "label" ; folderEntries : 12 ; isFolderSelected : false ; targetFolderID : "" ; targetLabelIDs : "" ; }
            ListElement{ folderId : "21" ; folderName : "Label3ToolongAndMustBeElidedSimilarToOnOfAccountsItJustNotNeedToBeThatLong" ; folderColor : "orange" ; folderType : "label" ; folderEntries : 40 ; isFolderSelected : true  ; targetFolderID : "" ; targetLabelIDs : "" ; }


            function setFolderSelection ( folderId   , toSelect ) { structureFuncs.setFolderSelection ( structurePM , folderId   , toSelect ) }
            function selectType         ( folderType , toSelect ) { structureFuncs.setTypeSelected    ( structurePM , folderType , toSelect ) }
            function setFromToDate      ( id , from, to )         { structureFuncs.setFromToDate       ( structureExternal , id , from, to ) }

            function getID             ( row        ) { return row == -1 ? structurePM.globalOptions.folderId : structurePM.get(row) .folderId }
            function getById           ( folderId   ) { return structureFuncs.getById           ( structurePM , folderId           ) }
            function getName           ( folderId   ) { return structurePM.getById              ( folderId    ) .folderName        }
            function getType           ( folderId   ) { return structurePM.getById              ( folderId    ) .folderType        }
            function getColor          ( folderId   ) { return structurePM.getById              ( folderId    ) .folderColor       }
            function getFrom           ( folderId   ) { return structurePM.getById              ( folderId    ) .fromDate          }
            function getTo             ( folderId   ) { return structurePM.getById              ( folderId    ) .toDate            }
            function getTargetLabelIDs ( folderId   ) { return structurePM.getById              ( folderId    ) .getTargetLabelIDs }
            function hasFolderWithName ( folderName ) { return structureFuncs.hasFolderWithName ( structurePM , folderName         ) }

            onDataChanged: {
                structureFuncs.updateSelection(structurePM)
            }
        }

        QtObject {
            id: structureFuncs

            function setFolderSelection   (model, id , toSelect ) {
                console.log(" set folde sel", id, toSelect)
                for (var i= -1; i<model.count; i++) {
                    var entry = i<0 ? model.globalOptions : model.get(i)
                    //console.log(" listing ",i, entry.folderId)
                    if (entry.folderId == id) {
                        entry.isFolderSelected = toSelect
                        if (i<0) model.globalOptionsChanged()
                        else model.set(i,entry)
                        console.log(" match & set", entry.toSelect)
                        return
                    }
                }
            }

            function setTypeSelected      (model, folderType , toSelect ) {
                console.log(" select type ", folderType, toSelect)
                for (var i= -1; i<model.count; i++) {
                    var entry = i<0 ? model.globalOptions : model.get(i)
                    console.log(" listing ",i, entry.folderType)
                    if (entry.folderType == folderType) {
                        entry.isFolderSelected = toSelect
                        if (i<0) model.globalOptionsChanged()
                        else model.set(i,entry)
                        console.log(" match & set", entry.isFolderSelected)
                    }
                }
            }

            function setFromToDate    (model, id , from, to ) {
                console.log(" set from to date id ", id, from, to)
                for (var i= -1; i<model.count; i++) {
                    var entry = i<0 ? model.globalOptions : model.get(i)
                    // console.log(" listing ",i, entry.targetFolderID)
                    if (entry.folderId == id) {
                        entry.fromDate = from
                        entry.toDate = to
                        if (i<0) model.globalOptionsChanged()
                        else model.set(i,entry)
                        console.log(" match & set", entry.fromDate, entry.toDate)
                        break
                    }
                }
            }

            function setTargetFolderID    (model, id , target ) {
                console.log(" set target folder id ", id, target)
                for (var i= -1; i<model.count; i++) {
                    var entry = i<0 ? model.globalOptions : model.get(i)
                    // console.log(" listing ",i, entry.targetFolderID)
                    if (entry.folderId == id) {
                        entry.targetFolderID=target
                        if (target=="") entry.targetLabelIDs=target
                        if (i<0) model.globalOptionsChanged()
                        else model.set(i,entry)
                        console.log(" match & set", entry.targetFolderID)
                        break
                    }
                }
            }

            function getById ( model, folderId  ) {
                console.log("called get object", folderId)
                for (var i= -1; i<model.count; i++) {
                    var entry = i<0 ? model.globalOptions : model.get(i)
                    //console.log(" listing ",i, entry.folderId)
                    if (entry.folderId == folderId) return entry
                }
                return undefined
            }

            function addTargetLabelID     (model, id , label ) {
                console.log(" add target label ", id, label)
                for (var i= -1; i<model.count; i++) {
                    var entry = i<0 ? model.globalOptions : model.get(i)
                    //console.log(" listing ",i, entry.targetLabelIDs)
                    if (entry.folderId == id) {
                        entry.targetLabelIDs += ";" + label
                        if (i<0) model.globalOptionsChanged()
                        else model.set(i,entry)
                        //console.log(" match & set", entry.targetLabelIDs)
                        break
                    }
                }
            }

            function removeTargetLabelID  (model, id , label ) {
                console.log(" remove target label ", id, label)
                for (var i= -1; i<model.count; i++) {
                    var entry = i<0 ? model.globalOptions : model.get(i)
                    //console.log(" listing ",i, entry.targetLabelIDs)
                    if (entry.folderId == id) {
                        var update = entry.targetLabelIDs
                        update = update.replace(new RegExp(';'+label,'gi'), "" )
                        entry.targetLabelIDs = update
                        if (i<0) model.globalOptionsChanged()
                        else model.set(i,entry)
                        //console.log(" match & set", entry.targetLabelIDs)
                        break
                    }
                }
            }

            function updateSelection(model) {
                console.log("Source folders changed", model)
                model.selectedLabels = true
                model.selectedFolders = true
                model.atLeastOneSelected = false
                for (var i= 0; i<model.count; i++) {
                    var item = model.get(i)
                    //console.log(" looping ", item.folderType)

                    if ( item.folderType == testgui.enums.folderTypeFolder ) model.selectedFolders = item.isFolderSelected && model.selectedFolders
                    if ( item.folderType == testgui.enums.folderTypeLabel  ) model.selectedLabels  = item.isFolderSelected && model.selectedLabels
                    if (                    item.isFolderSelected          ) atLeastOneSelected    = true

                    if (!model.selectedLabels && !model.selectedFolders && model.atLeastOneSelected) break
                }
            }

            function hasFolderWithName(model, folderName) {
                for (var i= 0; i<model.count; i++) {
                    if (model.get(i).folderName == folderName) return true
                }
                return false
            }

            function hasTarget(model) {
                for (var i= 0; i<model.count; i++) {
                    if (model.get(i).targetFolderID != "") return true
                }
                return false
            }
        }

        ListModel{
            id: errorList

            ListElement{ mailSubject : "Want some soup"                                                ; mailDate : "March 2 , 2019 12 : 00 : 22" ; inputFolder : "Archive" ; mailFrom : "me@me.me" ; errorMessage : "Something went wrong and import retry was not successful" ; }
            ListElement{ mailSubject : "RE: Office party"                                              ; mailDate : "March 2 , 2019 12 : 00 : 22" ; inputFolder : "Archive" ; mailFrom : "me@me.me" ; errorMessage : "Something went wrong and import retry was not successful" ; }
            ListElement{ mailSubject : "Hello Andy"                                                    ; mailDate : "March 2 , 2019 12 : 00 : 22" ; inputFolder : "Archive" ; mailFrom : "me@me.me" ; errorMessage : "Something went wrong and import retry was not successful" ; }
            ListElement{ mailSubject : "Pop art is cool again"                                         ; mailDate : "March 2 , 2019 12 : 00 : 22" ; inputFolder : "Archive" ; mailFrom : "me@me.me" ; errorMessage : "Something went wrong and import retry was not successful" ; }
            ListElement{ mailSubject : "Check this cute kittens play volleyball on Copacabanana beach" ; mailDate : "March 2 , 2019 12 : 00 : 22" ; inputFolder : "Archive" ; mailFrom : "me@me.me" ; errorMessage : "Something went wrong and import retry was not successful" ; }
        }
    }


    // moc go
    QtObject {
        id: go

        property int isAutoStart : 1
        property bool isFirstStart : false
        property string currentAddress : "none"
        //property string goos : "windows"
        property string goos : "linux"
        //property string goos : "darwin"
        property bool isDefaultPort : false

        property string wrongCredentials
        property string wrongMailboxPassword
        property string canNotReachAPI
        property string versionCheckFailed
        property string credentialsNotRemoved
        property string bugNotSent
        property string bugReportSent

        property string programTitle : "ProtonMail Import/Export Tool"
        property string newversion : "q0.1.0"
        property string landingPage : "https://jakub.cuth.sk/bridge"
        property string changelog  : "• Lorem ipsum dolor sit amet\n• consetetur sadipscing elitr,\n• sed diam nonumy eirmod tempor invidunt ut labore et dolore magna aliquyam erat,\n• sed diam voluptua.\n• At vero eos et accusam et justo duo dolores et ea rebum. Stet clita kasd gubergren, no sea takimata sanctus est Lorem ipsum dolor sit amet."
        //property string changelog  : ""
        property string bugfixes   : "• lorem ipsum dolor sit amet;• consetetur sadipscing elitr;• sed diam nonumy eirmod tempor invidunt ut labore et dolore magna aliquyam erat;• sed diam voluptua;• at vero eos et accusam et justo duo dolores et ea rebum;• stet clita kasd gubergren, no sea takimata sanctus est Lorem ipsum dolor sit amet"
        //property string bugfixes   : ""

        property real progress: 0.0
        property int progressFails: 0
        property string progressDescription: "nothing"
        property string progressInit: "init"
        property int total: 42
        property string importLogFileName: "importLogFileName not set"

        signal toggleMainWin(int systX, int systY, int systW, int systH)

        signal showWindow()
        signal showHelp()
        signal showQuit()

        signal notifyVersionIsTheLatest()
        signal setUpdateState(string updateState)

        signal showMainWin()
        signal hideMainWin()
        signal simpleErrorHappen()
        signal importStructuresLoadFinished(bool okay)
        signal exportStructureLoadFinished(bool okay)
        signal folderUpdateFinished()
        signal loginFinished()

        signal processFinished()
        signal toggleAutoStart()
        signal notifyBubble(int tabIndex, string message)
        signal runCheckVersion(bool showMessage)
        signal setAddAccountWarning(string message)
        signal notifyUpgrade()
        signal updateFinished(bool hasError)

        signal notifyLogout(string accname)

        signal notifyError(int errCode)
        property string errorDescription : ""

        function delay(duration) {
            var timeStart = new Date().getTime();

            while (new Date().getTime() - timeStart < duration) {
                // Do nothing
            }
        }


        function sendBug(desc,client,address){
            console.log("Test: sending ", desc, client, address)
            return desc.includes("fail")
        }

        function deleteAccount(index,remove) {
            console.log ("Test: Delete account ",index," and remove prefences "+remove)
            workAndClose("deleteAccount")
            accountsModel.remove(index)
        }

        function logoutAccount(index) {
            accountsModel.get(index).status="disconnected"
            workAndClose("logout")
        }

        function login(username,password) {
            delay(700)
            if (password=="wrong") {
                setAddAccountWarning("Wrong password")
                return -1
            }
            if (username=="2fa") {
                return 1
            }
            if (username=="mbox") {
                return 2
            }
            return 0
        }

        function auth2FA(twoFACode){
            delay(700)
            if (twoFACode=="wrong") {
                setAddAccountWarning("Wrong 2FA")
                return -1
            }
            if (twoFACode=="mbox") {
                return 1
            }
            return 0
        }

        function addAccount(mailboxPass) {
            delay(700)
            if (mailboxPass=="wrong") {
                setAddAccountWarning("Wrong mailbox password")
                return -1
            }
            accountsModel.append({
                "account" : testgui.winMain.dialogAddUser.username,
                "status" : "connected",
                "isExpanded":true,
                "hostname" : "127.0.0.1",
                "password" : "ZI9tKp+ryaxmbpn2E12",
                "security" : "StarTLS",
                "portSMTP" : 1025,
                "portIMAP" : 1143,
                "aliases" : "cuto@pm.com;jaku@pm.com;theHorriblySlowMurderWithExtremelyInefficientWeapon@youtube.com",
                "isCombinedAddressMode": true
            })
            workAndClose("addAccount")
        }

        property SequentialAnimation animateProgressBarUpgrade : SequentialAnimation {
            // version
            PropertyAnimation{ target: go; properties: "progressDescription"; to: 1; duration: 1; }
            PropertyAnimation{ duration: 2000; }

            // download
            PropertyAnimation{ target: go; properties: "progressDescription"; to: 2; duration: 1; }
            PropertyAnimation{ duration: 500; }
            PropertyAnimation{ target: go; properties: "progress"; to: 0.01; duration: 1; }
            PropertyAnimation{ duration: 500; }
            PropertyAnimation{ target: go; properties: "progress"; to: 0.1; duration: 1; }
            PropertyAnimation{ duration: 500; }
            PropertyAnimation{ target: go; properties: "progress"; to: 0.3; duration: 1; }
            PropertyAnimation{ duration: 500; }
            PropertyAnimation{ target: go; properties: "progress"; to: 0.5; duration: 1; }
            PropertyAnimation{ duration: 500; }
            PropertyAnimation{ target: go; properties: "progress"; to: 0.8; duration: 1; }
            PropertyAnimation{ duration: 500; }
            PropertyAnimation{ target: go; properties: "progress"; to: 1.0; duration: 1; }
            PropertyAnimation{ duration: 1000; }

            // verify
            PropertyAnimation{ target: go; properties: "progress"; to: 0.0; duration: 1; }
            PropertyAnimation{ target: go; properties: "progressDescription"; to: 3; duration: 1; }
            PropertyAnimation{ duration: 2000; }

            // unzip
            PropertyAnimation{ target: go; properties: "progressDescription"; to: 4; duration: 1; }
            PropertyAnimation{ duration: 500; }
            PropertyAnimation{ target: go; properties: "progress"; to: 0.01; duration: 1; }
            PropertyAnimation{ duration: 500; }
            PropertyAnimation{ target: go; properties: "progress"; to: 0.1; duration: 1; }
            PropertyAnimation{ duration: 500; }
            PropertyAnimation{ target: go; properties: "progress"; to: 0.3; duration: 1; }
            PropertyAnimation{ duration: 500; }
            PropertyAnimation{ target: go; properties: "progress"; to: 0.5; duration: 1; }
            PropertyAnimation{ duration: 500; }
            PropertyAnimation{ target: go; properties: "progress"; to: 0.8; duration: 1; }
            PropertyAnimation{ duration: 500; }
            PropertyAnimation{ target: go; properties: "progress"; to: 1.0; duration: 1; }
            PropertyAnimation{ duration: 2000; }

            // update
            PropertyAnimation{ target: go; properties: "progress"; to: 0.0; duration: 1; }
            PropertyAnimation{ target: go; properties: "progressDescription"; to: 5; duration: 1; }
            PropertyAnimation{ duration: 2000; }

            // quit
            PropertyAnimation{ target: go; properties: "progressDescription"; to: 6; duration: 1; }
            PropertyAnimation{ duration: 2000; }
        }

        property SequentialAnimation animateProgressBar : SequentialAnimation {
            id: apb
            property real speedup : 1.0;
            PropertyAnimation{ target: go; properties: "importLogFileName"; to: ""; duration: 1; }
            PropertyAnimation{ target: go; properties: "progressDescription"; to: go.progressInit; duration: 1; }
            PropertyAnimation{ duration: 2000/apb.speedup; }
            PropertyAnimation{ target: go; properties: "importLogFileName"; to: "/home/cuto/.local/state/protonmail/import-export/c0/import_1554732302.log"; duration: 1; }
            PropertyAnimation{ target: go; properties: "total"; to: 11; duration: 1; }
            PropertyAnimation{ duration: 1000/apb.speedup; }
            PropertyAnimation{ target: go; properties: "total"; to: 24; duration: 1; }
            PropertyAnimation{ duration: 1000/apb.speedup; }
            PropertyAnimation{ target: go; properties: "total"; to: 42; duration: 1; }
            PropertyAnimation{ target: go; properties: "progressDescription"; to: "/path/to/export/folder/"; duration: 1; }
            PropertyAnimation{ target: go; properties: "progress"; to: 0.01; duration: 1; }
            PropertyAnimation{ duration: 1000/apb.speedup; }
            PropertyAnimation{ target: go; properties: "progress"; to: 0.1; duration: 1; }
            PropertyAnimation{ duration: 1000/apb.speedup; }
            PropertyAnimation{ target: go; properties: "progress"; to: 0.3; duration: 1; }
            PropertyAnimation{ target: go; properties: "progressFails"; to: 1; duration: 1; }
            PropertyAnimation{ duration: 1000/apb.speedup; }
            PropertyAnimation{ target: go; properties: "progressDescription"; to: "/path/to/Lorem ipsum dolor sit amet, consetetur sadipscing elitr, sed diam nonumy eirmod tempor invidunt ut labore et dolore magna aliquyam erat, sed diam voluptua. At vero eos et accusam et justo duo dolores et ea rebum. Stet clita kasd gubergren, no sea takimata sanctus est Lorem ipsum dolor sit amet/export/folder/"; duration: 1; }
            PropertyAnimation{ target: go; properties: "progress"; to: 0.5; duration: 1; }
            PropertyAnimation{ duration: 1000/apb.speedup; }
            PropertyAnimation{ target: go; properties: "progress"; to: 0.8; duration: 1; }
            PropertyAnimation{ target: go; properties: "progressFails"; to: 13; duration: 1; }
            PropertyAnimation{ duration: 1000/apb.speedup; }
            PropertyAnimation{ target: go; properties: "progressDescription"; to: "/path/to/export/lastfolder/"; duration: 1; }
            PropertyAnimation{ target: go; properties: "progress"; to: 0.9; duration: 1; }
            PropertyAnimation{ duration: 1000/apb.speedup; }
            PropertyAnimation{ target: go; properties: "progress"; to: 1.0; duration: 1; }
        }

        function pauseProcess() {
            console.log("paused at ", go.progress)
            go.animateProgressBar.pause()
        }

        function resumeProcess() {
            console.log("resumed at ", go.progress)
            go.animateProgressBar.resume()
        }

        function cancelProcess(clearUnfinished) {
            console.log("stopped at ", go.progress, " clearing unfinished", clearUnfinished)
            go.animateProgressBar.stop()
        }

        property Timer timer : Timer {
            id: timer
            interval : 1000
            repeat : false
            property string work
            onTriggered : {
                console.log("triggered "+timer.work)
                switch (timer.work) {
                    case "isNewVersionAvailable" :
                    case "clearCache" :
                    case "clearKeychain" :
                    case "logout" :
                    go.processFinished()
                    break;

                    case "addAccount" :
                    case "login" :
                    go.loginFinished()
                    break;

                    case "loadStructureForExport" :
                    go.exportStructureLoadFinished(true)
                    break;

                    case "setupAndLoadForImport" :
                    case "loadStructuresForImport" :
                    go.importStructuresLoadFinished(true)
                    break;

                    case "startExport" :
                    case "startImport" :
                    go.animateProgressBar.start()
                    break;

                    case "startUpgrade":
                    go.animateProgressBarUpgrade.start()
                    go.updateFinished(true)

                    default:
                    console.log("no action")
                }
            }
        }

        function workAndClose(workDescription) {
            go.progress=0.0
            timer.work = workDescription
            timer.start()
        }

        function startUpgrade() {
            timer.work="startUpgrade"
            timer.start()
        }




        function checkPathStatus(path) {
            if ( path == ""                     ) return testgui.enums.pathEmptyPath
            if ( path == "wrong"                ) return testgui.enums.pathWrongPath
            if ( path == "/root"                ) return testgui.enums.pathWrongPermissions
            if ( path == "/home/cuto/file"      ) return testgui.enums.pathOK | testgui.enums.pathNotADir
            if ( path == "/home/cuto/empty/"    ) return testgui.enums.pathOK | testgui.enums.pathDirEmpty
            if ( path == "/home/cuto/Desktop"   ) return testgui.enums.pathOK | testgui.enums.pathDirEmpty
            if ( path == "/home/cuto/nonEmpty/" ) return testgui.enums.pathOK
            if ( path == "/home/cuto/ok/" ) return testgui.enums.pathOK
            return testgui.enums.pathWrongPath
        }


        function strategies() {
            return ["strategy1", "strategy2"]
        }

        function notPresentStrategy() {
            return ["notStrategy1", "notStrategy2"]
        }

        function loadAccounts() {
            console.log("Test: Account loaded")
        }

        function openDownloadLink(){
        }

        function loadStructureForExport(address) {
            workAndClose("loadStructureForExport")
        }

        function loadStructuresForImport(address) {
            workAndClose("loadStructuresForImport")
        }

        function setupAndLoadForImport(address) {
            workAndClose("setupAndLoadForImport")
        }

        function buildStructuresMapping() {
            var model = structureExternal
            console.log(" buildStructuresMapping aka reset all")
            for (var i= -1; i<model.count; i++) {
                console.log(" get ", i)
                var entry = i<0 ? model.globalOptions : model.get(i)
                console.log("     ", entry.folderId, entry.targetFolderID, entry.targetLabelIDs)
                if (entry.folderType == testgui.enums.folderTypeSystem) {
                    entry.targetLabelIDs = ";20;29"
                    entry.targetFolderID = entry.folderId=="global--uniq" ? "" :  (
                        i%2==0 ? "14" : "16"
                    )
                    entry.fromDate = 0
                    entry.toDate = 0
                } else {
                    entry.targetLabelIDs = ""
                    entry.targetFolderID = ""
                    entry.fromDate = 300000
                    entry.toDate = 15000000
                }
                entry.isFolderSelected = false
                console.log(" set ", i, entry.targetFolderID, entry.targetLabelIDs)
                if (i<0) model.globalOptionsChanged()
                else model.set(i,entry)
            }
        }

        function startExport(path,address,format,dateRange,encryptedBodies) {
            console.log ("Starting export: ",path, address, format, dateRange, encryptedBodies)
            workAndClose("startExport")
        }

        function startImport(address) {
            workAndClose("startImport")
        }

        function resetSource() {
        }

        function setupRemoteSource(username, password, host, port) {
            console.log("setup remote source", username, password, host, port)
        }

        function setupLocalSource(path) {
            console.log("setup local source", path)
        }





        function switchAddressMode(username){
            for (var iAcc=0; iAcc < accountsModel.count; iAcc++) {
                if (accountsModel.get(iAcc).account == username ) {
                    accountsModel.get(iAcc).isCombinedAddressMode = !accountsModel.get(iAcc).isCombinedAddressMode
                    break
                }
            }
            workAndClose("switchAddressMode")
        }

        function isNewVersionAvailable(showMessage){
            if (testroot.newVersion)  {
                setUpdateState("oldVersion")
            } else {
                setUpdateState("upToDate")
                if(showMessage) {
                    notifyVersionIsTheLatest()
                }
            }
            workAndClose("isNewVersionAvailable")
            //notifyBubble(2,go.versionCheckFailed)
            return 0
        }

        function getLocalVersionInfo(){}

        function getBackendVersion() {
            return "PIMP 1.0"
        }

        property bool isConnectionOK : true
        signal setConnectionStatus(bool isAvailable)

        function configureAppleMail(iAccount,iAddress) {
            console.log ("Test: autoconfig account ",iAccount," address ",iAddress)
        }

        function openLogs() {
            Qt.openUrlExternally("file:///home/cuto/")
        }

        function highlightSystray() {
            test_systray.highlight()
        }

        function normalSystray() {
            test_systray.normal()
        }

        signal bubbleClosed()

        function getIMAPPort() {
            return 1143
        }
        function getSMTPPort() {
            return 1025
        }

        function isPortOpen(portstring){
            if (isNaN(portstring)) {
                return 1
            }
            var portnum = parseInt(portstring,10)
            if (portnum < 3333) {
                return 1
            }
            return 0
        }

        signal openManual()

        function clearCache() {
            workAndClose("clearCache")
        }

        function clearKeychain() {
            workAndClose("clearKeychain")
        }

        function leastUsedColor() {
            return "#cf5858"
        }


        function answerSkip(skipAll) {
            go.animateProgressBar.resume()
        }

        function answerRetry(){
            go.animateProgressBar.resume()
        }

        function createLabelOrFolder(address,fname,fcolor,isFolder,sourceID){
            console.log("-> createLabelOrFolder", address, fname, fcolor, isFolder, sourceID)
            return (fname!="fail")
        }

        function checkInternet() {
            // nothing to do
        }

        function loadImportReports(fname) {
            console.log("load import reports for ", fname)
        }


        onToggleAutoStart: {
            workAndClose("toggleAutoStart")
            isAutoStart = (isAutoStart!=0) ? 0 : 1
            console.log (" Test: toggleAutoStart "+isAutoStart)
        }

        function openReport() {
            Qt.openUrlExternally("file:///home/cuto/")
        }

        function sendImportReport(address, fname) {
            console.log("sending import report from ", address, " file ", fname)
            return !fname.includes("fail")
        }
    }
}
