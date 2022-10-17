# Copyright (c) 2022 Proton AG
#
# This file is part of Proton Mail Bridge.
#
# Proton Mail Bridge is free software: you can redistribute it and/or modify
# it under the terms of the GNU General Public License as published by
# the Free Software Foundation, either version 3 of the License, or
# (at your option) any later version.
#
# Proton Mail Bridge is distributed in the hope that it will be useful,
# but WITHOUT ANY WARRANTY; without even the implied warranty of
# MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
# GNU General Public License for more details.
#
# You should have received a copy of the GNU General Public License
# along with Proton Mail Bridge. If not, see <https://www.gnu.org/licenses/>.

cmake_minimum_required(VERSION 3.22)

#*****************************************************************************************************************************************************
# Deploy
#*****************************************************************************************************************************************************

install(SCRIPT ${deploy_script})

# QML
install(DIRECTORY "${QT_DIR}/qml/Qt"
        DESTINATION "${CMAKE_INSTALL_PREFIX}/bridge-gui.app/Contents/MacOS")
install(DIRECTORY "${QT_DIR}/qml/QtQml"
        DESTINATION "${CMAKE_INSTALL_PREFIX}/bridge-gui.app/Contents/MacOS")
install(DIRECTORY "${QT_DIR}/qml/QtQuick"
        DESTINATION "${CMAKE_INSTALL_PREFIX}/bridge-gui.app/Contents/MacOS")
# FRAMEWORKS
install(DIRECTORY "${QT_DIR}/lib/QtQmlWorkerScript.framework"
        DESTINATION "${CMAKE_INSTALL_PREFIX}/bridge-gui.app/Contents/Frameworks")
install(DIRECTORY "${QT_DIR}/lib/QtQuickControls2Impl.framework"
        DESTINATION "${CMAKE_INSTALL_PREFIX}/bridge-gui.app/Contents/Frameworks")
install(DIRECTORY "${QT_DIR}/lib/QtQuickLayouts.framework"
        DESTINATION "${CMAKE_INSTALL_PREFIX}/bridge-gui.app/Contents/Frameworks")
install(DIRECTORY "${QT_DIR}/lib/QtQuickDialogs2.framework"
        DESTINATION "${CMAKE_INSTALL_PREFIX}/bridge-gui.app/Contents/Frameworks")
install(DIRECTORY "${QT_DIR}/lib/QtQuickDialogs2QuickImpl.framework"
        DESTINATION "${CMAKE_INSTALL_PREFIX}/bridge-gui.app/Contents/Frameworks")
install(DIRECTORY "${QT_DIR}/lib/QtQuickDialogs2Utils.framework"
        DESTINATION "${CMAKE_INSTALL_PREFIX}/bridge-gui.app/Contents/Frameworks")
# PLUGINS
install(FILES "${QT_DIR}/plugins/imageformats/libqsvg.dylib"
        DESTINATION "${CMAKE_INSTALL_PREFIX}/bridge-gui.app/Contents/PlugIns/imageformats")

