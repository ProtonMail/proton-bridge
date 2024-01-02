// Copyright (c) 2024 Proton AG
// This file is part of Proton Mail Bridge.
// Proton Mail Bridge is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
// Proton Mail Bridge is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
// You should have received a copy of the GNU General Public License
// along with Proton Mail Bridge. If not, see <https://www.gnu.org/licenses/>.
import QtQuick
import QtQuick.Layouts
import QtQuick.Controls
import Proton
import "SetupWizard"

Rectangle {
    id: root

    property ColorScheme colorScheme

    color: root.colorScheme.background_norm

    signal startSetup()

    ColumnLayout {
        anchors.fill: parent
        spacing: 0

        // we use the setup wizard left pane (onboarding version)
        LeftPane {
            Layout.alignment: Qt.AlignHCenter
            Layout.fillHeight: true
            Layout.preferredWidth: ProtonStyle.wizard_pane_width
            colorScheme: root.colorScheme
            wizard: setupWizard

            Component.onCompleted: {
                showNoAccount();
            }
            onStartSetup: {
                root.startSetup();
            }
        }
        Image {
            id: mailLogoWithWordmark
            Layout.alignment: Qt.AlignHCenter
            Layout.bottomMargin: ProtonStyle.wizard_window_margin
            height: sourceSize.height
            source: root.colorScheme.mail_logo_with_wordmark
            sourceSize.height: 36
            sourceSize.width: 134
            width: sourceSize.width
        }
    }
}