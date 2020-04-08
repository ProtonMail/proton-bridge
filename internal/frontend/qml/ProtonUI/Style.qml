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

// colors, fonts, etc.
pragma Singleton
import QtQuick 2.8

QtObject {
    // Colors, dimensions and font

    property TextMetrics oneInch : TextMetrics { // 72 points is one inch
        id: oneInch
        font.pointSize: 72
        text: "HI"
    }
    property real dpi    : oneInch.height // one inch in pixel
    property real refdpi : 96.0
    property real px     : dpi/refdpi // default unit, scaled to current DPI
    property real pt     : 80 / dpi /// conversion from px to pt (aestetic correction of font size +3 points)

    property color transparent: "transparent"

    //=================//
    //     Stelian     //
    //=================//
    // Main window
    property QtObject main : QtObject {
        property color background   : "#353440"
        property color text         : "#ffffff"
        property color textInactive : "#cdcaf7"
        property color textDisabled : "#bbbbbb"
        property color textBlue     : "#9199cc"
        property color textRed      : "#ef6a5e"
        property color textGreen    : "#41a56c"
        property color textOrange   : "#e6922e"
        property color line         : "#44444f"
        property real dummy              : 10  * px
        property real width              : 650 * px
        property real height             : 420 * px
        property real heightRow          : 54  * px
        property real heightLine         : 2   * px
        property real leftMargin         : 17  * px
        property real rightMargin        : 17  * px
        property real systrayMargin      : 20  * px
        property real fontSize           : 12  * px
        property real iconSize           : 15  * px
        property real leftMarginButton   : 9   * px
        property real verCheckRepeatTime : 15*60*60*1000 // milliseconds
        property real topMargin          : fontSize
        property real bottomMargin       : fontSize
        property real border             : 1   * px
    }

    property QtObject bugreport : QtObject {
        property real width  : 645 * px
        property real height : 495 * px
    }

    property QtObject errorDialog : QtObject {
        property color background : bubble.background
        property color text       : bubble.text
        property real  fontSize   : dialog.fontSize
        property real  radius     : 10 * px
    }

    property QtObject dialog : QtObject {
        property color background : "#ee353440"
        property color text       : "#ffffff"
        property color line       : "#505061"
        property color textBlue   : "#9396cc"
        property color shadow     : "#19505061"
        property real fontSize         : 14  * px
        property real titleSize        : 17  * px
        property real iconSize         : 18  * px
        property real widthButton      : 100 * px
        property real heightButton     : 3*fontSize
        property real radiusButton     : 5   * px
        property real borderButton     : 2   * px
        property real borderInput      : 1.2 * px
        property real heightSeparator  : 21  * px
        property real widthInput       : 280 * px
        property real heightInput      : 30  * px
        property real heightButtonIcon : 150 * px
        property real widthButtonIcon  : 160 * px
        property real rightMargin      : 17  * px
        property real bottomMargin     : 17  * px
        property real topMargin        : 17  * px
        property real leftMargin       : 17  * px
        property real heightInputBox   : 4*fontSize
        property real widthInputBox    : 280 * px
        property real spacing          : 5   * px
    }

    // Title specific
    property QtObject title : QtObject {
        property color background : "#000"
        property color text       : main.text
        property real height      : 26 * px
        property real leftMargin  : 10 * px
        property real fontSize    : 14 * px
    }

    property QtObject titleMacOS : QtObject {
        property color background  : tabbar.background
        property real  height      : 22 * px
        property real  imgHeight   : 12 * px
        property real  leftMargin  : 8  * px
        property real  fontSize    : 14 * px
        property real  radius      : 7 * px
    }

    // Tabline specific
    property QtObject tabbar : QtObject {
        property color background    : "#302f3a"
        property color text          : "#ffffff"
        property color textInactive  : "#9696a7"
        property real height         : 68 * px
        property real widthButton    : 63 * px
        property real heightButton   : 38 * px
        property real spacingButton  : 35 * px
        property real heightTriangle : 7  * px
        property real fontSize       : 12 * px
        property real iconSize       : 17 * px
        property real bottomMargin   : (height-heightButton)/2
        property real widthUpdate    : 138 * px
        property real heightUpdate   : 40 * px
        property real leftMargin     : main.leftMargin
        property string rightButton  : "quit"
    }

    // Bubble specific
    property QtObject bubble: QtObject {
        property color background     : "#454553"
        property color paneBackground : background
        property color text           : dialog.text
        property real height          : 185 * px
        property real width           : 220 * px
        property real radius          : 5   * px
        property real widthPane       : 20  * px
        property real iconSize        : 14  * px
        property real fontSize        : main.fontSize
    }


    property QtObject menu : QtObject {
        property color background : "#454553"
        property color line       : "#505061"
        property color lineAlt    : "#565668"
        property color text       : "#ffffff"
        property real width        : 184 * px
        property real height       : 200 * px
        property real radius       : 7 * px
        property real topMargin    : 21*px + tabbar.height
        property real rightMargin  : 24 * px
        //property real heightLine  : (width - 2*radius) / 3
    }

    property QtObject accounts : QtObject {
        property color line               : main.line
        property color backgroundExpanded : "#444456"
        property color backgroundAddrRow  : "#1cffffff"
        property real heightLine    : main.heightLine
        property real heightHeader  : 38 * px
        property real heightAccount : main.heightRow
        property real heightFooter  : 75 * px

        property real elideWidth  : 285 * px
        property real leftMargin2 : 334 * px
        property real leftMargin3 : 463 * px
        property real leftMargin4 : 567 * px
        property real sizeChevron : 9   * px

        property real heightAddrRow  : 39 * px
        property real heightAddr     : 32 * px
        property real leftMarginAddr : 28 * px
    }

    property QtObject settings : QtObject {
        property real fontSize    : 15 * px
        property real iconSize    : 20 * px
        property real toggleSize  : 28 * px
    }

    property QtObject info : QtObject {
        property real width          : 315 * px
        property real height         : 450 * px
        property real heightHeader   : 32  * px
        property real topMargin      : 18  * px
        property real iconSize       : 16  * px
        property real leftMarginIcon : 8   * px
        property real widthValue     : 180 * px
    }

    property QtObject exporting : QtObject {
        property color background         : dialog.background
        property color rowBackground      : "#f8f8f8"
        property color sliderBackground    : "#e6e6e6"
        property color sliderForeground    : "#515061"
        property color line               : dialog.line
        property color text               : "#333"
        property color progressBackground : "#aaa"
        property color progressStatus     : main.textBlue // need gradient +- 1.2 light
        property real  boxRadius          : 5  * px
        property real  rowHeight          : 32 * px
        property real  leftMargin         : 5  * px
        property real  leftMargin2        : 18 * px
        property real  leftMargin3        : 30 * px
    }

    property int okInfoBar         : 0
    property int warnInfoBar       : 1
    property int warnBubbleMessage : 2
    property int errorInfoBar      : 4



    // old color pick
    property QtObject old : QtObject {
        /// old scheme
        property color darkBackground:  "grey"
        property color darkForground:   "red"
        property color darkInactive:    "grey"
        property color lightBackground: "white"
        property color lightForground:  "grey"
        property color lightInactive:   "grey"
        property color linkColor:       "blue"
        property color warnColor:       "orange"
        //
        property color pm_black:   "#000" // pitch black
        property color pm_ddgrey:  "#333" // dark background
        property color pm_dgrey:   "#555" //
        property color pm_grey:    "#999" // inactive dark, lines
        property color pm_white:   "#fff" // super white
        property color pm_lgrey:   "#acb0bf" // inactive light
        property color pm_llgrey:  "#e6eaf0" // light background
        property color pm_blue:    "#8286c5"
        property color pm_lblue:   "#9397cd"
        property color pm_llblue:  "#e3e4f2"
        property color pm_dred:    "#c26164"
        property color pm_red:     "#af4649"
        property color pm_orange:  "#d9c4a2" // warning
        property color pm_lorange: "#e7d360" // warning
        property color pm_green:   "#a6cc93" // success

        property color web_dark_side_back          : "#333333"
        property color web_dark_side_text_inactive : "#999999"
        property color web_dark_highl_back         : "#3F3F3F"
        property color web_dark_highl_text         : "#FFFFFF"
        property color web_dark_highl_icon         : "#A8ABD7"
        property color web_dark_top_back           : "#555555"
        property color web_dark_top_icon           : "#E9E9E9"
        property color web_butt_blue_back          : "#9397CD"
        property color web_butt_blue_text          : "#FFFFFF"
        property color web_row_inactive_back       : "#DDDDDD"
        property color web_row_inactive_text       : "#55556E"
        property color web_butt_grey_text_line     : "#838897"
        property color web_butt_grey_text_hover    : "#222222"
        property color web_butt_grey_top_back      : "#FDFDFD"
        property color web_butt_grey_low_back      : "#DEDEDE"
        property color web_main_back               : "#FFFFFF"
        property color web_main_text               : "#555555"


        // colors
        property color pmold_dblue:  "#333367"
        property color pmold_blue:   "#58588c"
        property color pmold_red:    "#9e3c3c"
        property color pmold_orange: "#9e7f3c"
        property color pmold_green:  "#5e9162"
        property color pmold_gray:   "#484a61"
        // highlited
        //property color highlBackground:  pm_lblue
        //property color highlForground:   pm_dgrey
        //property color highlInactive:    pm_grey
        //property color buttLine:  pm_grey
        property color buttLight: "#fdfdfd" // button background start
        property color buttDark:  "#dedede" // button background start
    }


    // font
    property FontLoader fontawesome : FontLoader {
        //source: "qrc://fontawesome.ttf"
        source: "fontawesome.ttf"
    }

    property QtObject fa : QtObject {
        property string glass                               : "\uf000"
        property string music                               : "\uf001"
        property string search                              : "\uf002"
        property string envelope_o                          : "\uf003"
        property string heart                               : "\uf004"
        property string star                                : "\uf005"
        property string star_o                              : "\uf006"
        property string user                                : "\uf007"
        property string film                                : "\uf008"
        property string th_large                            : "\uf009"
        property string th                                  : "\uf00a"
        property string th_list                             : "\uf00b"
        property string check                               : "\uf00c"
        property string remove                              : "\uf00d"
        property string close                               : "\uf00d"
        property string times                               : "\uf00d"
        property string search_plus                         : "\uf00e"
        property string search_minus                        : "\uf010"
        property string power_off                           : "\uf011"
        property string signal                              : "\uf012"
        property string gear                                : "\uf013"
        property string cog                                 : "\uf013"
        property string trash_o                             : "\uf014"
        property string home                                : "\uf015"
        property string file_o                              : "\uf016"
        property string clock_o                             : "\uf017"
        property string road                                : "\uf018"
        property string download                            : "\uf019"
        property string arrow_circle_o_down                 : "\uf01a"
        property string arrow_circle_o_up                   : "\uf01b"
        property string inbox                               : "\uf01c"
        property string play_circle_o                       : "\uf01d"
        property string rotate_right                        : "\uf01e"
        property string repeat                              : "\uf01e"
        property string refresh                             : "\uf021"
        property string list_alt                            : "\uf022"
        property string lock                                : "\uf023"
        property string flag                                : "\uf024"
        property string headphones                          : "\uf025"
        property string volume_off                          : "\uf026"
        property string volume_down                         : "\uf027"
        property string volume_up                           : "\uf028"
        property string qrcode                              : "\uf029"
        property string barcode                             : "\uf02a"
        property string tag                                 : "\uf02b"
        property string tags                                : "\uf02c"
        property string book                                : "\uf02d"
        property string bookmark                            : "\uf02e"
        property string printer                             : "\uf02f"
        property string camera                              : "\uf030"
        property string font                                : "\uf031"
        property string bold                                : "\uf032"
        property string italic                              : "\uf033"
        property string text_height                         : "\uf034"
        property string text_width                          : "\uf035"
        property string align_left                          : "\uf036"
        property string align_center                        : "\uf037"
        property string align_right                         : "\uf038"
        property string align_justify                       : "\uf039"
        property string list                                : "\uf03a"
        property string dedent                              : "\uf03b"
        property string outdent                             : "\uf03b"
        property string indent                              : "\uf03c"
        property string video_camera                        : "\uf03d"
        property string photo                               : "\uf03e"
        property string image                               : "\uf03e"
        property string picture_o                           : "\uf03e"
        property string pencil                              : "\uf040"
        property string map_marker                          : "\uf041"
        property string adjust                              : "\uf042"
        property string tint                                : "\uf043"
        property string edit                                : "\uf044"
        property string pencil_square_o                     : "\uf044"
        property string share_square_o                      : "\uf045"
        property string check_square_o                      : "\uf046"
        property string arrows                              : "\uf047"
        property string step_backward                       : "\uf048"
        property string fast_backward                       : "\uf049"
        property string backward                            : "\uf04a"
        property string play                                : "\uf04b"
        property string pause                               : "\uf04c"
        property string stop                                : "\uf04d"
        property string forward                             : "\uf04e"
        property string fast_forward                        : "\uf050"
        property string step_forward                        : "\uf051"
        property string eject                               : "\uf052"
        property string chevron_left                        : "\uf053"
        property string chevron_right                       : "\uf054"
        property string plus_circle                         : "\uf055"
        property string minus_circle                        : "\uf056"
        property string times_circle                        : "\uf057"
        property string check_circle                        : "\uf058"
        property string question_circle                     : "\uf059"
        property string info_circle                         : "\uf05a"
        property string crosshairs                          : "\uf05b"
        property string times_circle_o                      : "\uf05c"
        property string check_circle_o                      : "\uf05d"
        property string ban                                 : "\uf05e"
        property string arrow_left                          : "\uf060"
        property string arrow_right                         : "\uf061"
        property string arrow_up                            : "\uf062"
        property string arrow_down                          : "\uf063"
        property string mail_forward                        : "\uf064"
        property string share                               : "\uf064"
        property string expand                              : "\uf065"
        property string compress                            : "\uf066"
        property string plus                                : "\uf067"
        property string minus                               : "\uf068"
        property string asterisk                            : "\uf069"
        property string exclamation_circle                  : "\uf06a"
        property string gift                                : "\uf06b"
        property string leaf                                : "\uf06c"
        property string fire                                : "\uf06d"
        property string eye                                 : "\uf06e"
        property string eye_slash                           : "\uf070"
        property string warning                             : "\uf071"
        property string exclamation_triangle                : "\uf071"
        property string plane                               : "\uf072"
        property string calendar                            : "\uf073"
        property string random                              : "\uf074"
        property string comment                             : "\uf075"
        property string magnet                              : "\uf076"
        property string chevron_up                          : "\uf077"
        property string chevron_down                        : "\uf078"
        property string retweet                             : "\uf079"
        property string shopping_cart                       : "\uf07a"
        property string folder                              : "\uf07b"
        property string folder_open                         : "\uf07c"
        property string arrows_v                            : "\uf07d"
        property string arrows_h                            : "\uf07e"
        property string bar_chart_o                         : "\uf080"
        property string bar_chart                           : "\uf080"
        property string twitter_square                      : "\uf081"
        property string facebook_square                     : "\uf082"
        property string camera_retro                        : "\uf083"
        property string key                                 : "\uf084"
        property string gears                               : "\uf085"
        property string cogs                                : "\uf085"
        property string comments                            : "\uf086"
        property string thumbs_o_up                         : "\uf087"
        property string thumbs_o_down                       : "\uf088"
        property string star_half                           : "\uf089"
        property string heart_o                             : "\uf08a"
        property string sign_out                            : "\uf08b"
        property string linkedin_square                     : "\uf08c"
        property string thumb_tack                          : "\uf08d"
        property string external_link                       : "\uf08e"
        property string sign_in                             : "\uf090"
        property string trophy                              : "\uf091"
        property string github_square                       : "\uf092"
        property string upload                              : "\uf093"
        property string lemon_o                             : "\uf094"
        property string phone                               : "\uf095"
        property string square_o                            : "\uf096"
        property string bookmark_o                          : "\uf097"
        property string phone_square                        : "\uf098"
        property string twitter                             : "\uf099"
        property string facebook_f                          : "\uf09a"
        property string facebook                            : "\uf09a"
        property string github                              : "\uf09b"
        property string unlock                              : "\uf09c"
        property string credit_card                         : "\uf09d"
        property string feed                                : "\uf09e"
        property string rss                                 : "\uf09e"
        property string hdd_o                               : "\uf0a0"
        property string bullhorn                            : "\uf0a1"
        property string bell                                : "\uf0f3"
        property string certificate                         : "\uf0a3"
        property string hand_o_right                        : "\uf0a4"
        property string hand_o_left                         : "\uf0a5"
        property string hand_o_up                           : "\uf0a6"
        property string hand_o_down                         : "\uf0a7"
        property string arrow_circle_left                   : "\uf0a8"
        property string arrow_circle_right                  : "\uf0a9"
        property string arrow_circle_up                     : "\uf0aa"
        property string arrow_circle_down                   : "\uf0ab"
        property string globe                               : "\uf0ac"
        property string wrench                              : "\uf0ad"
        property string tasks                               : "\uf0ae"
        property string filter                              : "\uf0b0"
        property string briefcase                           : "\uf0b1"
        property string arrows_alt                          : "\uf0b2"
        property string group                               : "\uf0c0"
        property string users                               : "\uf0c0"
        property string chain                               : "\uf0c1"
        property string link                                : "\uf0c1"
        property string cloud                               : "\uf0c2"
        property string flask                               : "\uf0c3"
        property string cut                                 : "\uf0c4"
        property string scissors                            : "\uf0c4"
        property string copy                                : "\uf0c5"
        property string files_o                             : "\uf0c5"
        property string paperclip                           : "\uf0c6"
        property string save                                : "\uf0c7"
        property string floppy_o                            : "\uf0c7"
        property string square                              : "\uf0c8"
        property string navicon                             : "\uf0c9"
        property string reorder                             : "\uf0c9"
        property string bars                                : "\uf0c9"
        property string list_ul                             : "\uf0ca"
        property string list_ol                             : "\uf0cb"
        property string strikethrough                       : "\uf0cc"
        property string underline                           : "\uf0cd"
        property string table                               : "\uf0ce"
        property string magic                               : "\uf0d0"
        property string truck                               : "\uf0d1"
        property string pinterest                           : "\uf0d2"
        property string pinterest_square                    : "\uf0d3"
        property string google_plus_square                  : "\uf0d4"
        property string google_plus                         : "\uf0d5"
        property string money                               : "\uf0d6"
        property string caret_down                          : "\uf0d7"
        property string caret_up                            : "\uf0d8"
        property string caret_left                          : "\uf0d9"
        property string caret_right                         : "\uf0da"
        property string columns                             : "\uf0db"
        property string unsorted                            : "\uf0dc"
        property string sort                                : "\uf0dc"
        property string sort_down                           : "\uf0dd"
        property string sort_desc                           : "\uf0dd"
        property string sort_up                             : "\uf0de"
        property string sort_asc                            : "\uf0de"
        property string envelope                            : "\uf0e0"
        property string linkedin                            : "\uf0e1"
        property string rotate_left                         : "\uf0e2"
        property string undo                                : "\uf0e2"
        property string legal                               : "\uf0e3"
        property string gavel                               : "\uf0e3"
        property string dashboard                           : "\uf0e4"
        property string tachometer                          : "\uf0e4"
        property string comment_o                           : "\uf0e5"
        property string comments_o                          : "\uf0e6"
        property string flash                               : "\uf0e7"
        property string bolt                                : "\uf0e7"
        property string sitemap                             : "\uf0e8"
        property string umbrella                            : "\uf0e9"
        property string paste                               : "\uf0ea"
        property string clipboard                           : "\uf0ea"
        property string lightbulb_o                         : "\uf0eb"
        property string exchange                            : "\uf0ec"
        property string cloud_download                      : "\uf0ed"
        property string cloud_upload                        : "\uf0ee"
        property string user_md                             : "\uf0f0"
        property string stethoscope                         : "\uf0f1"
        property string suitcase                            : "\uf0f2"
        property string bell_o                              : "\uf0a2"
        property string coffee                              : "\uf0f4"
        property string cutlery                             : "\uf0f5"
        property string file_text_o                         : "\uf0f6"
        property string building_o                          : "\uf0f7"
        property string hospital_o                          : "\uf0f8"
        property string ambulance                           : "\uf0f9"
        property string medkit                              : "\uf0fa"
        property string fighter_jet                         : "\uf0fb"
        property string beer                                : "\uf0fc"
        property string h_square                            : "\uf0fd"
        property string plus_square                         : "\uf0fe"
        property string angle_double_left                   : "\uf100"
        property string angle_double_right                  : "\uf101"
        property string angle_double_up                     : "\uf102"
        property string angle_double_down                   : "\uf103"
        property string angle_left                          : "\uf104"
        property string angle_right                         : "\uf105"
        property string angle_up                            : "\uf106"
        property string angle_down                          : "\uf107"
        property string desktop                             : "\uf108"
        property string laptop                              : "\uf109"
        property string tablet                              : "\uf10a"
        property string mobile_phone                        : "\uf10b"
        property string mobile                              : "\uf10b"
        property string circle_o                            : "\uf10c"
        property string quote_left                          : "\uf10d"
        property string quote_right                         : "\uf10e"
        property string spinner                             : "\uf110"
        property string circle                              : "\uf111"
        property string mail_reply                          : "\uf112"
        property string reply                               : "\uf112"
        property string github_alt                          : "\uf113"
        property string folder_o                            : "\uf114"
        property string folder_open_o                       : "\uf115"
        property string smile_o                             : "\uf118"
        property string frown_o                             : "\uf119"
        property string meh_o                               : "\uf11a"
        property string gamepad                             : "\uf11b"
        property string keyboard_o                          : "\uf11c"
        property string flag_o                              : "\uf11d"
        property string flag_checkered                      : "\uf11e"
        property string terminal                            : "\uf120"
        property string code                                : "\uf121"
        property string mail_reply_all                      : "\uf122"
        property string reply_all                           : "\uf122"
        property string star_half_empty                     : "\uf123"
        property string star_half_full                      : "\uf123"
        property string star_half_o                         : "\uf123"
        property string location_arrow                      : "\uf124"
        property string crop                                : "\uf125"
        property string code_fork                           : "\uf126"
        property string unlink                              : "\uf127"
        property string chain_broken                        : "\uf127"
        property string question                            : "\uf128"
        property string info                                : "\uf129"
        property string exclamation                         : "\uf12a"
        property string superscript                         : "\uf12b"
        property string subscript                           : "\uf12c"
        property string eraser                              : "\uf12d"
        property string puzzle_piece                        : "\uf12e"
        property string microphone                          : "\uf130"
        property string microphone_slash                    : "\uf131"
        property string shield                              : "\uf132"
        property string calendar_o                          : "\uf133"
        property string fire_extinguisher                   : "\uf134"
        property string rocket                              : "\uf135"
        property string maxcdn                              : "\uf136"
        property string chevron_circle_left                 : "\uf137"
        property string chevron_circle_right                : "\uf138"
        property string chevron_circle_up                   : "\uf139"
        property string chevron_circle_down                 : "\uf13a"
        property string html5                               : "\uf13b"
        property string css3                                : "\uf13c"
        property string anchor                              : "\uf13d"
        property string unlock_alt                          : "\uf13e"
        property string bullseye                            : "\uf140"
        property string ellipsis_h                          : "\uf141"
        property string ellipsis_v                          : "\uf142"
        property string rss_square                          : "\uf143"
        property string play_circle                         : "\uf144"
        property string ticket                              : "\uf145"
        property string minus_square                        : "\uf146"
        property string minus_square_o                      : "\uf147"
        property string level_up                            : "\uf148"
        property string level_down                          : "\uf149"
        property string check_square                        : "\uf14a"
        property string pencil_square                       : "\uf14b"
        property string external_link_square                : "\uf14c"
        property string share_square                        : "\uf14d"
        property string compass                             : "\uf14e"
        property string toggle_down                         : "\uf150"
        property string caret_square_o_down                 : "\uf150"
        property string toggle_up                           : "\uf151"
        property string caret_square_o_up                   : "\uf151"
        property string toggle_right                        : "\uf152"
        property string caret_square_o_right                : "\uf152"
        property string euro                                : "\uf153"
        property string eur                                 : "\uf153"
        property string gbp                                 : "\uf154"
        property string dollar                              : "\uf155"
        property string usd                                 : "\uf155"
        property string rupee                               : "\uf157"
        property string inr                                 : "\uf156"
        property string cny                                 : "\uf157"
        property string rmb                                 : "\uf157"
        property string yen                                 : "\uf157"
        property string jpy                                 : "\uf157"
        property string ruble                               : "\uf158"
        property string rouble                              : "\uf158"
        property string rub                                 : "\uf158"
        property string won                                 : "\uf159"
        property string krw                                 : "\uf159"
        property string bitcoin                             : "\uf15a"
        property string btc                                 : "\uf15a"
        property string file                                : "\uf15b"
        property string file_text                           : "\uf15c"
        property string sort_alpha_asc                      : "\uf15d"
        property string sort_alpha_desc                     : "\uf15e"
        property string sort_amount_asc                     : "\uf160"
        property string sort_amount_desc                    : "\uf161"
        property string sort_numeric_asc                    : "\uf162"
        property string sort_numeric_desc                   : "\uf163"
        property string thumbs_up                           : "\uf164"
        property string thumbs_down                         : "\uf165"
        property string youtube_square                      : "\uf166"
        property string youtube                             : "\uf167"
        property string xing                                : "\uf168"
        property string xing_square                         : "\uf169"
        property string youtube_play                        : "\uf16a"
        property string dropbox                             : "\uf16b"
        property string stack_overflow                      : "\uf16c"
        property string instagram                           : "\uf16d"
        property string flickr                              : "\uf16e"
        property string adn                                 : "\uf170"
        property string bitbucket                           : "\uf171"
        property string bitbucket_square                    : "\uf172"
        property string tumblr                              : "\uf173"
        property string tumblr_square                       : "\uf174"
        property string long_arrow_down                     : "\uf175"
        property string long_arrow_up                       : "\uf176"
        property string long_arrow_left                     : "\uf177"
        property string long_arrow_right                    : "\uf178"
        property string apple                               : "\uf179"
        property string windows                             : "\uf17a"
        property string android                             : "\uf17b"
        property string linux                               : "\uf17c"
        property string dribbble                            : "\uf17d"
        property string skype                               : "\uf17e"
        property string foursquare                          : "\uf180"
        property string trello                              : "\uf181"
        property string female                              : "\uf182"
        property string male                                : "\uf183"
        property string gittip                              : "\uf184"
        property string gratipay                            : "\uf184"
        property string sun_o                               : "\uf185"
        property string moon_o                              : "\uf186"
        property string archive                             : "\uf187"
        property string bug                                 : "\uf188"
        property string vk                                  : "\uf189"
        property string weibo                               : "\uf18a"
        property string renren                              : "\uf18b"
        property string pagelines                           : "\uf18c"
        property string stack_exchange                      : "\uf18d"
        property string arrow_circle_o_right                : "\uf18e"
        property string arrow_circle_o_left                 : "\uf190"
        property string toggle_left                         : "\uf191"
        property string caret_square_o_left                 : "\uf191"
        property string dot_circle_o                        : "\uf192"
        property string wheelchair                          : "\uf193"
        property string vimeo_square                        : "\uf194"
        property string turkish_lira                        : "\uf195"
        property string fa_try                              : "\uf195"
        property string plus_square_o                       : "\uf196"
        property string space_shuttle                       : "\uf197"
        property string slack                               : "\uf198"
        property string envelope_square                     : "\uf199"
        property string wordpress                           : "\uf19a"
        property string openid                              : "\uf19b"
        property string institution                         : "\uf19c"
        property string bank                                : "\uf19c"
        property string university                          : "\uf19c"
        property string mortar_board                        : "\uf19d"
        property string graduation_cap                      : "\uf19d"
        property string yahoo                               : "\uf19e"
        property string google                              : "\uf1a0"
        property string reddit                              : "\uf1a1"
        property string reddit_square                       : "\uf1a2"
        property string stumbleupon_circle                  : "\uf1a3"
        property string stumbleupon                         : "\uf1a4"
        property string delicious                           : "\uf1a5"
        property string digg                                : "\uf1a6"
        property string pied_piper_pp                       : "\uf1a7"
        property string pied_piper_alt                      : "\uf1a8"
        property string drupal                              : "\uf1a9"
        property string joomla                              : "\uf1aa"
        property string language                            : "\uf1ab"
        property string fax                                 : "\uf1ac"
        property string building                            : "\uf1ad"
        property string child                               : "\uf1ae"
        property string paw                                 : "\uf1b0"
        property string spoon                               : "\uf1b1"
        property string cube                                : "\uf1b2"
        property string cubes                               : "\uf1b3"
        property string behance                             : "\uf1b4"
        property string behance_square                      : "\uf1b5"
        property string steam                               : "\uf1b6"
        property string steam_square                        : "\uf1b7"
        property string recycle                             : "\uf1b8"
        property string automobile                          : "\uf1b9"
        property string car                                 : "\uf1b9"
        property string cab                                 : "\uf1ba"
        property string taxi                                : "\uf1ba"
        property string tree                                : "\uf1bb"
        property string spotify                             : "\uf1bc"
        property string deviantart                          : "\uf1bd"
        property string soundcloud                          : "\uf1be"
        property string database                            : "\uf1c0"
        property string file_pdf_o                          : "\uf1c1"
        property string file_word_o                         : "\uf1c2"
        property string file_excel_o                        : "\uf1c3"
        property string file_powerpoint_o                   : "\uf1c4"
        property string file_photo_o                        : "\uf1c5"
        property string file_picture_o                      : "\uf1c5"
        property string file_image_o                        : "\uf1c5"
        property string file_zip_o                          : "\uf1c6"
        property string file_archive_o                      : "\uf1c6"
        property string file_sound_o                        : "\uf1c7"
        property string file_audio_o                        : "\uf1c7"
        property string file_movie_o                        : "\uf1c8"
        property string file_video_o                        : "\uf1c8"
        property string file_code_o                         : "\uf1c9"
        property string vine                                : "\uf1ca"
        property string codepen                             : "\uf1cb"
        property string jsfiddle                            : "\uf1cc"
        property string life_bouy                           : "\uf1cd"
        property string life_buoy                           : "\uf1cd"
        property string life_saver                          : "\uf1cd"
        property string support                             : "\uf1cd"
        property string life_ring                           : "\uf1cd"
        property string circle_o_notch                      : "\uf1ce"
        property string ra                                  : "\uf1d0"
        property string resistance                          : "\uf1d0"
        property string rebel                               : "\uf1d0"
        property string ge                                  : "\uf1d1"
        property string empire                              : "\uf1d1"
        property string git_square                          : "\uf1d2"
        property string git                                 : "\uf1d3"
        property string y_combinator_square                 : "\uf1d4"
        property string yc_square                           : "\uf1d4"
        property string hacker_news                         : "\uf1d4"
        property string tencent_weibo                       : "\uf1d5"
        property string qq                                  : "\uf1d6"
        property string wechat                              : "\uf1d7"
        property string weixin                              : "\uf1d7"
        property string send                                : "\uf1d8"
        property string paper_plane                         : "\uf1d8"
        property string send_o                              : "\uf1d9"
        property string paper_plane_o                       : "\uf1d9"
        property string history                             : "\uf1da"
        property string circle_thin                         : "\uf1db"
        property string header                              : "\uf1dc"
        property string paragraph                           : "\uf1dd"
        property string sliders                             : "\uf1de"
        property string share_alt                           : "\uf1e0"
        property string share_alt_square                    : "\uf1e1"
        property string bomb                                : "\uf1e2"
        property string soccer_ball_o                       : "\uf1e3"
        property string futbol_o                            : "\uf1e3"
        property string tty                                 : "\uf1e4"
        property string binoculars                          : "\uf1e5"
        property string plug                                : "\uf1e6"
        property string slideshare                          : "\uf1e7"
        property string twitch                              : "\uf1e8"
        property string yelp                                : "\uf1e9"
        property string newspaper_o                         : "\uf1ea"
        property string wifi                                : "\uf1eb"
        property string calculator                          : "\uf1ec"
        property string paypal                              : "\uf1ed"
        property string google_wallet                       : "\uf1ee"
        property string cc_visa                             : "\uf1f0"
        property string cc_mastercard                       : "\uf1f1"
        property string cc_discover                         : "\uf1f2"
        property string cc_amex                             : "\uf1f3"
        property string cc_paypal                           : "\uf1f4"
        property string cc_stripe                           : "\uf1f5"
        property string bell_slash                          : "\uf1f6"
        property string bell_slash_o                        : "\uf1f7"
        property string trash                               : "\uf1f8"
        property string copyright                           : "\uf1f9"
        property string at                                  : "\uf1fa"
        property string eyedropper                          : "\uf1fb"
        property string paint_brush                         : "\uf1fc"
        property string birthday_cake                       : "\uf1fd"
        property string area_chart                          : "\uf1fe"
        property string pie_chart                           : "\uf200"
        property string line_chart                          : "\uf201"
        property string lastfm                              : "\uf202"
        property string lastfm_square                       : "\uf203"
        property string toggle_off                          : "\uf204"
        property string toggle_on                           : "\uf205"
        property string bicycle                             : "\uf206"
        property string bus                                 : "\uf207"
        property string ioxhost                             : "\uf208"
        property string angellist                           : "\uf209"
        property string cc                                  : "\uf20a"
        property string shekel                              : "\uf20b"
        property string sheqel                              : "\uf20b"
        property string ils                                 : "\uf20b"
        property string meanpath                            : "\uf20c"
        property string buysellads                          : "\uf20d"
        property string connectdevelop                      : "\uf20e"
        property string dashcube                            : "\uf210"
        property string forumbee                            : "\uf211"
        property string leanpub                             : "\uf212"
        property string sellsy                              : "\uf213"
        property string shirtsinbulk                        : "\uf214"
        property string simplybuilt                         : "\uf215"
        property string skyatlas                            : "\uf216"
        property string cart_plus                           : "\uf217"
        property string cart_arrow_down                     : "\uf218"
        property string diamond                             : "\uf219"
        property string ship                                : "\uf21a"
        property string user_secret                         : "\uf21b"
        property string motorcycle                          : "\uf21c"
        property string street_view                         : "\uf21d"
        property string heartbeat                           : "\uf21e"
        property string venus                               : "\uf221"
        property string mars                                : "\uf222"
        property string mercury                             : "\uf223"
        property string intersex                            : "\uf224"
        property string transgender                         : "\uf224"
        property string transgender_alt                     : "\uf225"
        property string venus_double                        : "\uf226"
        property string mars_double                         : "\uf227"
        property string venus_mars                          : "\uf228"
        property string mars_stroke                         : "\uf229"
        property string mars_stroke_v                       : "\uf22a"
        property string mars_stroke_h                       : "\uf22b"
        property string neuter                              : "\uf22c"
        property string genderless                          : "\uf22d"
        property string facebook_official                   : "\uf230"
        property string pinterest_p                         : "\uf231"
        property string whatsapp                            : "\uf232"
        property string server                              : "\uf233"
        property string user_plus                           : "\uf234"
        property string user_times                          : "\uf235"
        property string hotel                               : "\uf236"
        property string bed                                 : "\uf236"
        property string viacoin                             : "\uf237"
        property string train                               : "\uf238"
        property string subway                              : "\uf239"
        property string medium                              : "\uf23a"
        property string yc                                  : "\uf23b"
        property string y_combinator                        : "\uf23b"
        property string optin_monster                       : "\uf23c"
        property string opencart                            : "\uf23d"
        property string expeditedssl                        : "\uf23e"
        property string battery_4                           : "\uf240"
        property string battery                             : "\uf240"
        property string battery_full                        : "\uf240"
        property string battery_3                           : "\uf241"
        property string battery_three_quarters              : "\uf241"
        property string battery_2                           : "\uf242"
        property string battery_half                        : "\uf242"
        property string battery_1                           : "\uf243"
        property string battery_quarter                     : "\uf243"
        property string battery_0                           : "\uf244"
        property string battery_empty                       : "\uf244"
        property string mouse_pointer                       : "\uf245"
        property string i_cursor                            : "\uf246"
        property string object_group                        : "\uf247"
        property string object_ungroup                      : "\uf248"
        property string sticky_note                         : "\uf249"
        property string sticky_note_o                       : "\uf24a"
        property string cc_jcb                              : "\uf24b"
        property string cc_diners_club                      : "\uf24c"
        property string clone                               : "\uf24d"
        property string balance_scale                       : "\uf24e"
        property string hourglass_o                         : "\uf250"
        property string hourglass_1                         : "\uf251"
        property string hourglass_start                     : "\uf251"
        property string hourglass_2                         : "\uf252"
        property string hourglass_half                      : "\uf252"
        property string hourglass_3                         : "\uf253"
        property string hourglass_end                       : "\uf253"
        property string hourglass                           : "\uf254"
        property string hand_grab_o                         : "\uf255"
        property string hand_rock_o                         : "\uf255"
        property string hand_stop_o                         : "\uf256"
        property string hand_paper_o                        : "\uf256"
        property string hand_scissors_o                     : "\uf257"
        property string hand_lizard_o                       : "\uf258"
        property string hand_spock_o                        : "\uf259"
        property string hand_pointer_o                      : "\uf25a"
        property string hand_peace_o                        : "\uf25b"
        property string trademark                           : "\uf25c"
        property string registered                          : "\uf25d"
        property string creative_commons                    : "\uf25e"
        property string gg                                  : "\uf260"
        property string gg_circle                           : "\uf261"
        property string tripadvisor                         : "\uf262"
        property string odnoklassniki                       : "\uf263"
        property string odnoklassniki_square                : "\uf264"
        property string get_pocket                          : "\uf265"
        property string wikipedia_w                         : "\uf266"
        property string safari                              : "\uf267"
        property string chrome                              : "\uf268"
        property string firefox                             : "\uf269"
        property string opera                               : "\uf26a"
        property string internet_explorer                   : "\uf26b"
        property string tv                                  : "\uf26c"
        property string television                          : "\uf26c"
        property string contao                              : "\uf26d"
        property string fa_500px                            : "\uf26e"
        property string amazon                              : "\uf270"
        property string calendar_plus_o                     : "\uf271"
        property string calendar_minus_o                    : "\uf272"
        property string calendar_times_o                    : "\uf273"
        property string calendar_check_o                    : "\uf274"
        property string industry                            : "\uf275"
        property string map_pin                             : "\uf276"
        property string map_signs                           : "\uf277"
        property string map_o                               : "\uf278"
        property string map                                 : "\uf279"
        property string commenting                          : "\uf27a"
        property string commenting_o                        : "\uf27b"
        property string houzz                               : "\uf27c"
        property string vimeo                               : "\uf27d"
        property string black_tie                           : "\uf27e"
        property string fonticons                           : "\uf280"
        property string reddit_alien                        : "\uf281"
        property string edge                                : "\uf282"
        property string credit_card_alt                     : "\uf283"
        property string codiepie                            : "\uf284"
        property string modx                                : "\uf285"
        property string fort_awesome                        : "\uf286"
        property string usb                                 : "\uf287"
        property string product_hunt                        : "\uf288"
        property string mixcloud                            : "\uf289"
        property string scribd                              : "\uf28a"
        property string pause_circle                        : "\uf28b"
        property string pause_circle_o                      : "\uf28c"
        property string stop_circle                         : "\uf28d"
        property string stop_circle_o                       : "\uf28e"
        property string shopping_bag                        : "\uf290"
        property string shopping_basket                     : "\uf291"
        property string hashtag                             : "\uf292"
        property string bluetooth                           : "\uf293"
        property string bluetooth_b                         : "\uf294"
        property string percent                             : "\uf295"
        property string gitlab                              : "\uf296"
        property string wpbeginner                          : "\uf297"
        property string wpforms                             : "\uf298"
        property string envira                              : "\uf299"
        property string universal_access                    : "\uf29a"
        property string wheelchair_alt                      : "\uf29b"
        property string question_circle_o                   : "\uf29c"
        property string blind                               : "\uf29d"
        property string audio_description                   : "\uf29e"
        property string volume_control_phone                : "\uf2a0"
        property string braille                             : "\uf2a1"
        property string assistive_listening_systems         : "\uf2a2"
        property string asl_interpreting                    : "\uf2a3"
        property string american_sign_language_interpreting : "\uf2a3"
        property string deafness                            : "\uf2a4"
        property string hard_of_hearing                     : "\uf2a4"
        property string deaf                                : "\uf2a4"
        property string glide                               : "\uf2a5"
        property string glide_g                             : "\uf2a6"
        property string signing                             : "\uf2a7"
        property string sign_language                       : "\uf2a7"
        property string low_vision                          : "\uf2a8"
        property string viadeo                              : "\uf2a9"
        property string viadeo_square                       : "\uf2aa"
        property string snapchat                            : "\uf2ab"
        property string snapchat_ghost                      : "\uf2ac"
        property string snapchat_square                     : "\uf2ad"
        property string pied_piper                          : "\uf2ae"
        property string first_order                         : "\uf2b0"
        property string yoast                               : "\uf2b1"
        property string themeisle                           : "\uf2b2"
        property string google_plus_circle                  : "\uf2b3"
        property string google_plus_official                : "\uf2b3"
        property string fa                                  : "\uf2b4"
        property string font_awesome                        : "\uf2b4"
        property string handshake_o                         : "\uf2b5"
        property string envelope_open                       : "\uf2b6"
        property string envelope_open_o                     : "\uf2b7"
        property string linode                              : "\uf2b8"
        property string address_book                        : "\uf2b9"
        property string address_book_o                      : "\uf2ba"
        property string vcard                               : "\uf2bb"
        property string address_card                        : "\uf2bb"
        property string vcard_o                             : "\uf2bc"
        property string address_card_o                      : "\uf2bc"
        property string user_circle                         : "\uf2bd"
        property string user_circle_o                       : "\uf2be"
        property string user_o                              : "\uf2c0"
        property string id_badge                            : "\uf2c1"
        property string drivers_license                     : "\uf2c2"
        property string id_card                             : "\uf2c2"
        property string drivers_license_o                   : "\uf2c3"
        property string id_card_o                           : "\uf2c3"
        property string quora                               : "\uf2c4"
        property string free_code_camp                      : "\uf2c5"
        property string telegram                            : "\uf2c6"
        property string thermometer_4                       : "\uf2c7"
        property string thermometer                         : "\uf2c7"
        property string thermometer_full                    : "\uf2c7"
        property string thermometer_3                       : "\uf2c8"
        property string thermometer_three_quarters          : "\uf2c8"
        property string thermometer_2                       : "\uf2c9"
        property string thermometer_half                    : "\uf2c9"
        property string thermometer_1                       : "\uf2ca"
        property string thermometer_quarter                 : "\uf2ca"
        property string thermometer_0                       : "\uf2cb"
        property string thermometer_empty                   : "\uf2cb"
        property string shower                              : "\uf2cc"
        property string bathtub                             : "\uf2cd"
        property string s15                                 : "\uf2cd"
        property string bath                                : "\uf2cd"
        property string podcast                             : "\uf2ce"
        property string window_maximize                     : "\uf2d0"
        property string window_minimize                     : "\uf2d1"
        property string window_restore                      : "\uf2d2"
        property string times_rectangle                     : "\uf2d3"
        property string window_close                        : "\uf2d3"
        property string times_rectangle_o                   : "\uf2d4"
        property string window_close_o                      : "\uf2d4"
        property string bandcamp                            : "\uf2d5"
        property string grav                                : "\uf2d6"
        property string etsy                                : "\uf2d7"
        property string imdb                                : "\uf2d8"
        property string ravelry                             : "\uf2d9"
        property string eercast                             : "\uf2da"
        property string microchip                           : "\uf2db"
        property string snowflake_o                         : "\uf2dc"
        property string superpowers                         : "\uf2dd"
        property string wpexplorer                          : "\uf2de"
        property string meetup                              : "\uf2e0"
    }
}
