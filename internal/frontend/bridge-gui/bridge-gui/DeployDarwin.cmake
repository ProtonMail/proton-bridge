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
install(DIRECTORY "${QT_DIR}/qml/Qt/labs/platform"
        DESTINATION "${CMAKE_INSTALL_PREFIX}/bridge-gui.app/Contents/MacOS/Qt/labs")
install(DIRECTORY "${QT_DIR}/qml/QtQml"
        DESTINATION "${CMAKE_INSTALL_PREFIX}/bridge-gui.app/Contents/MacOS")
install(DIRECTORY "${QT_DIR}/qml/QtQuick"
        DESTINATION "${CMAKE_INSTALL_PREFIX}/bridge-gui.app/Contents/MacOS"
        PATTERN "VirtualKeyboard" EXCLUDE
        PATTERN "Effects" EXCLUDE
        PATTERN "LocalStorage" EXCLUDE
        PATTERN "NativeStyle" EXCLUDE
        PATTERN "Particles" EXCLUDE
        PATTERN "Scene2D" EXCLUDE
        PATTERN "Scene3D" EXCLUDE
        PATTERN "Shapes" EXCLUDE
        PATTERN "Timeline" EXCLUDE
        PATTERN "VectorImage" EXCLUDE

        PATTERN "Controls/FluentWinUI3" EXCLUDE
        PATTERN "Controls/designer" EXCLUDE
        PATTERN "Controls/Fusion" EXCLUDE
        PATTERN "Controls/Imagine" EXCLUDE
        PATTERN "Controls/Material" EXCLUDE
        PATTERN "Controls/Universal" EXCLUDE
        PATTERN "Controls/iOS" EXCLUDE
        PATTERN "Controls/macOS" EXCLUDE)

# FRAMEWORKS
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
# ADDITIONAL FRAMEWORKS FOR Qt 6.8
install(DIRECTORY "${QT_DIR}/lib/QtQuickControls2Basic.framework"
        DESTINATION "${CMAKE_INSTALL_PREFIX}/bridge-gui.app/Contents/Frameworks")
install(DIRECTORY "${QT_DIR}/lib/QtLabsPlatform.framework"
        DESTINATION "${CMAKE_INSTALL_PREFIX}/bridge-gui.app/Contents/Frameworks")
install(DIRECTORY "${QT_DIR}/lib/QtQuickControls2BasicStyleImpl.framework"
        DESTINATION "${CMAKE_INSTALL_PREFIX}/bridge-gui.app/Contents/Frameworks")

# PLUGINS
install(FILES "${QT_DIR}/plugins/imageformats/libqsvg.dylib"
        DESTINATION "${CMAKE_INSTALL_PREFIX}/bridge-gui.app/Contents/PlugIns/imageformats")

# crash handler utils
## Build
add_custom_command(OUTPUT "${CMAKE_CURRENT_BINARY_DIR}/gen_crashpad/crashpad_handler"
    COMMAND lipo
    ARGS -create -output "${CMAKE_CURRENT_BINARY_DIR}/gen_crashpad/crashpad_handler" "${VCPKG_INSTALLED_DIR}/arm64-osx-min-11-0/tools/sentry-native/crashpad_handler" "${VCPKG_INSTALLED_DIR}/x64-osx-min-10-15/tools/sentry-native/crashpad_handler"
    COMMENT Unifying crashpad_handler
)
add_custom_target(unify_crashpadHandler ALL DEPENDS "${CMAKE_CURRENT_BINARY_DIR}/gen_crashpad/crashpad_handler")
add_dependencies(bridge-gui unify_crashpadHandler)
## Install
install(PROGRAMS "${CMAKE_CURRENT_BINARY_DIR}/gen_crashpad/crashpad_handler"
DESTINATION "${CMAKE_INSTALL_PREFIX}/bridge-gui.app/Contents/MacOS/")

