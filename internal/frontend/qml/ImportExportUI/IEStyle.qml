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

// Adjust Bridge Style

import QtQuick 2.8
import ImportExportUI 1.0
import ProtonUI 1.0

Item {
    Component.onCompleted : {
        //Style.refdpi = go.goos == "darwin" ? 86.0 : 96.0
        Style.pt     = go.goos == "darwin" ? 93/Style.dpi : 80/Style.dpi

        Style.main.background   = "#fff"
        Style.main.text         = "#505061"
        Style.main.textInactive = "#686876"
        Style.main.line         = "#dddddd"
        Style.main.width        = 884 * Style.px
        Style.main.height       = 422 * Style.px
        Style.main.leftMargin   = 25 * Style.px
        Style.main.rightMargin  = 25 * Style.px

        Style.title.background = Style.main.text
        Style.title.text       = Style.main.background

        Style.tabbar.background    = "#3D3A47"
        Style.tabbar.rightButton   = "add account"
        Style.tabbar.spacingButton = 45*Style.px

        Style.accounts.backgroundExpanded = "#fafafa"
        Style.accounts.backgroundAddrRow  = "#fff"
        Style.accounts.leftMargin2        = Style.main.width/2
        Style.accounts.leftMargin3        = 5.5*Style.main.width/8


        Style.dialog.background   = "#fff"
        Style.dialog.text         = Style.main.text
        Style.dialog.line         = "#e2e2e2"
        Style.dialog.fontSize     = 12 * Style.px
        Style.dialog.heightInput  = 2.2*Style.dialog.fontSize
        Style.dialog.heightButton = Style.dialog.heightInput
        Style.dialog.borderInput  = 1 * Style.px

        Style.bubble.background     = "#595966"
        Style.bubble.paneBackground = "#454553"
        Style.bubble.text           = "#fff"
        Style.bubble.width          = 310 * Style.px
        Style.bubble.widthPane      = 36  * Style.px
        Style.bubble.iconSize       = 14  * Style.px


        // colors:
        // text: #515061
        // tick: #686876
        // blue icon: #9396cc
        // row bck: #f8f8f9
        // line: #ddddde or #e2e2e2
        //
        // slider bg: #e6e6e6
        // slider fg: #515061
        // info icon: #c3c3c8
        // input border: #ebebeb
        //
        // bubble color: #595966
        // bubble pane: #454553
        // bubble text: #fff
        //
        // indent folder
        //
        // Dimensions:
        // full width: 882px
        // leftMargin: 25px
        // rightMargin: 25px
        // rightMargin: 25px
        // middleSeparator: 69px
        // width folders: 416px or (width - separators) /2
        // width output: 346px or (width - separators) /2
        //
        // height from top to input begin: 78px
        // heightSeparator: 27px
        // height folder input: 26px
        //
        // buble width: 309px
        // buble left pane icon: 14px
        // buble left pane width: 36px or (2.5 icon width)
        // buble height: 46px
        // buble arrow height: 12px
        // buble arrow width: 14px
        // buble radius: 3-4px
    }
}
