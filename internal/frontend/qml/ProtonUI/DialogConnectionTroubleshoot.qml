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

// Dialog with Yes/No buttons

import QtQuick 2.8
import ProtonUI 1.0

Dialog {
    id: root

    title : qsTr(
        "Common connection problems and solutions",
        "Title of the network troubleshooting modal"
    )
    isDialogBusy: false // can close
    property var parContent : [
        [
            qsTr("Allow alternative routing"  , "Paragraph title"),
            qsTr(
                "In case Proton sites are blocked, this setting allows Bridge "+
                "to try alternative network routing to reach Proton, which can "+
                "be useful for bypassing firewalls or network issues. We recommend "+
                "keeping this setting on for greater reliability. "+
                '<a href="https://protonmail.com/blog/anti-censorship-alternative-routing/">Learn more</a>'+
                " and "+
                '<a href="showProxy">enable here</a>'+
                ".",
                "Paragraph content"
            ),
        ],

        [
            qsTr("No internet connection"                   , "Paragraph title"),
            qsTr(
                "Please make sure that your internet connection is working.",
                "Paragraph content"
            ),
        ],

        [
            qsTr("Internet Service Provider (ISP) problem"  , "Paragraph title"),
            qsTr(
                "Try connecting to Proton from a different network (or use "+
                '<a href="https://protonvpn.com/">ProtonVPN</a>'+
                " or "+
                '<a href="https://torproject.org/">Tor</a>'+
                ").",
                "Paragraph content"
            ),
        ],

        [
            qsTr("Government block"                         , "Paragraph title"),
            qsTr(
                "Your country may be blocking access to Proton. Try using "+
                '<a href="https://protonvpn.com/">ProtonVPN</a>'+
                " (or any other VPN) or "+
                '<a href="https://torproject.org/">Tor</a>'+
                ".",
                "Paragraph content"
            ),
        ],

        [
            qsTr("Antivirus interference"                   , "Paragraph title"),
            qsTr(
                "Temporarily disable or remove your antivirus software.",
                "Paragraph content"
            ),
        ],

        [
            qsTr("Proxy/Firewall interference"              , "Paragraph title"),
            qsTr(
                "Disable any proxies or firewalls, or contact your network administrator.",
                "Paragraph content"
            ),
        ],

        [
            qsTr("Still can’t find a solution"              , "Paragraph title"),
            qsTr(
                "Contact us directly through our "+
                '<a href="https://protonmail.com/support-form">support form</a>'+
                ", email (support@protonmail.com), or "+
                '<a href="https://twitter.com/ProtonMail">Twitter</a>'+
                ".",
                "Paragraph content"
            ),
        ],

        [
            qsTr("Proton is down"                           , "Paragraph title"),
            qsTr(
                "Check "+
                '<a href="https://protonstatus.com/">Proton Status</a>'+
                " for our system status.",
                "Paragraph content"
            ),
        ],

    ]

    Item {
        AccessibleText {
            anchors.centerIn: parent
            color: Style.old.pm_white
            linkColor: color
            width: parent.width - 50 * Style.px
            wrapMode: Text.WordWrap
            font.pointSize: Style.main.fontSize*Style.pt
            onLinkActivated: {
                if (link=="showProxy") {
                    dialogGlobal.state= "toggleAllowProxy"
                    dialogGlobal.show()
                } else {
                    Qt.openUrlExternally(link)
                }
            }
            text: {
                var content=""
                for (var i=0; i<root.parContent.length; i++) {
                    var par = root.parContent[i]
                    content += "<p>"
                    content += "<b>"+par[0]+":</b> "
                    content += par[1]
                    content += "</p>\n"
                }
                return content
            }
        }
    }
}
