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

// Change default keychain dialog

import QtQuick 2.8
import BridgeUI 1.0
import ProtonUI 1.0
import QtQuick.Controls 2.2 as QC
import QtQuick.Layouts 1.0

Dialog {
    id: root

    title : "Change which keychain Bridge uses as default"
    subtitle : "Select which keychain is used (Bridge will automatically restart)"
    isDialogBusy: currentIndex==1

    property var selectedKeychain

    Connections {
        target: go.selectedKeychain
        onValueChanged: {
            console.debug("go.selectedKeychain == ", go.selectedKeychain)
        }
    }

    ColumnLayout {
        Layout.fillHeight: true
        Layout.fillWidth: true

        Item {
            Layout.fillWidth: true
            Layout.minimumHeight: root.titleHeight + Style.dialog.heightSeparator
            Layout.maximumHeight: root.titleHeight + Style.dialog.heightSeparator
        }

        Item {
            Layout.fillHeight: true
            Layout.fillWidth: true

            ColumnLayout {
                anchors.centerIn: parent

                Repeater {
                    id: keychainRadioButtons
                    model: go.availableKeychain
                    QC.RadioButton {
                        id: radioDelegate
                        text: modelData
                        checked: go.selectedKeychain === modelData

                        Layout.alignment: Qt.AlignHCenter | Qt.AlignVCenter
                        spacing: Style.main.spacing

                        indicator: Text {
                            text  : radioDelegate.checked ? Style.fa.check_circle : Style.fa.circle_o
                            color : radioDelegate.checked ? Style.main.textBlue   : Style.main.textInactive
                            font {
                                pointSize: Style.dialog.iconSize * Style.pt
                                family: Style.fontawesome.name
                            }
                        }
                        contentItem: Text {
                            text: radioDelegate.text
                            color: Style.main.text
                            font {
                                pointSize: Style.dialog.fontSize * Style.pt
                                bold: checked
                            }
                            horizontalAlignment : Text.AlignHCenter
                            verticalAlignment   : Text.AlignVCenter
                            leftPadding: Style.dialog.iconSize
                        }

                        onCheckedChanged: {
                            if (checked) {
                                root.selectedKeychain = modelData
                            }
                        }
                    }
                }

                Item {
                    Layout.fillWidth: true
                    Layout.minimumHeight: Style.dialog.heightSeparator
                    Layout.maximumHeight: Style.dialog.heightSeparator
                }

                Row {
                    id: buttonRow
                    Layout.alignment: Qt.AlignHCenter | Qt.AlignVCenter
                    spacing: Style.dialog.spacing
                    ButtonRounded {
                        id:buttonNo
                        color_main: Style.dialog.text
                        fa_icon: Style.fa.times
                        text: qsTr("Cancel", "dismisses current action")
                        onClicked : root.hide()
                    }
                    ButtonRounded {
                        id: buttonYes
                        color_main: Style.dialog.text
                        color_minor: Style.main.textBlue
                        isOpaque: true
                        fa_icon: Style.fa.check
                        text: qsTr("Okay", "confirms and dismisses a notification")
                        onClicked : root.confirmed()
                    }
                }
            }
        }
    }

    ColumnLayout {
        Layout.fillHeight: true
        Layout.fillWidth: true

        Item {
            Layout.fillWidth: true
            Layout.minimumHeight: root.titleHeight + Style.dialog.heightSeparator
            Layout.maximumHeight: root.titleHeight + Style.dialog.heightSeparator
        }

        Item {
            Layout.fillHeight: true
            Layout.fillWidth: true
            Layout.alignment: Qt.AlignHCenter | Qt.AlignVCenter

            Text {
                id: answ
                anchors.horizontalCenter: parent.horizontalCenter
                anchors.verticalCenter: parent.verticalCenter
                width : parent.width/2
                color: Style.dialog.text
                font {
                    pointSize : Style.dialog.fontSize * Style.pt
                    bold      : true
                }
                text : "Default keychain is now set to " + root.selectedKeychain +
                "\n\n" +
                qsTr("Settings will be applied after the next start.", "notification about setting being applied after next start") +
                "\n\n" +
                qsTr("Bridge will now restart.", "notification about restarting")
                wrapMode: Text.Wrap
                horizontalAlignment: Text.AlignHCenter
            }
        }
    }

    Shortcut {
        sequence: StandardKey.Cancel
        onActivated: root.hide()
    }

    Shortcut {
        sequence: "Enter"
        onActivated: root.confirmed()
    }

    function confirmed() {
        if (selectedKeychain === go.selectedKeychain) {
            root.hide()
            return
        }

        incrementCurrentIndex()
        timer.start()
    }

    timer.interval : 5000

    Connections {
        target: timer
        onTriggered: {
            // This action triggers restart on the backend side.
            go.selectedKeychain = selectedKeychain
        }
    }
}
