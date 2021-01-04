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

import QtQuick 2.8
import BridgeUI 1.0
import ProtonUI 1.0


Dialog {
    id: root
    title: qsTr("Connection security error", "Title of modal explainning TLS issue")

    property string par1Title  : qsTr("Description:", "Title of paragraph describing the issue")
    property string par1Text   : qsTr (
        "ProtonMail Bridge was not able to establish a secure connection to Proton servers due to a TLS certificate error. "+
        "This means your connection may potentially be insecure and susceptible to monitoring by third parties.",
        "A paragraph describing the issue"
    )

    property string par2Title  : qsTr("Recommendation:", "Title of paragraph describing recommended steps")
    property string par2Text   : qsTr (
        "If you are on a corporate or public network, the network administrator may be monitoring or intercepting all traffic.",
        "A paragraph describing network issue"
    )
    property string par2ul1    : qsTr(
        "If you trust your network operator, you can continue to use ProtonMail as usual.",
        "A list item describing recomendation for trusted network"
    )

    property string par2ul2       : qsTr(
        "If you don't trust your network operator, reconnect to ProtonMail over a VPN (such as ProtonVPN) "+
        "which encrypts your Internet connection, or use a different network to access ProtonMail.",
        "A list item describing recomendation for untrusted network"
    )
    property string par3Text      : qsTr("Learn more on our knowledge base article","A paragraph describing where to find more information")
    property string kbArticleText : qsTr("What is TLS certificate error.", "Link text for knowledge base article")
    property string kbArticleLink : "https://protonmail.com/support/knowledge-base/"


    Item {
        AccessibleText {
            anchors.centerIn: parent
            color: Style.old.pm_white
            linkColor: color
            width: parent.width - 50 * Style.px
            wrapMode: Text.WordWrap
            font.pointSize: Style.main.fontSize*Style.pt
            onLinkActivated: Qt.openUrlExternally(link)
            text: "<h3>"+par1Title+"</h3>"+
            par1Text+"<br>\n"+
            "<h3>"+par2Title+"</h3>"+
            par2Text+
            "<ul>"+
            "<li>"+par2ul1+"</li>"+
            "<li>"+par2ul2+"</li>"+
            "</ul>"+"<br>\n"+
            ""
            //par3Text+
            //" <a href='"+kbArticleLink+"'>"+kbArticleText+"</a>\n"
        }
    }
}

