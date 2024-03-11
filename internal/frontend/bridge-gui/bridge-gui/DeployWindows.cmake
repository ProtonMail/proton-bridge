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

macro( AppendLib LIB_NAME HINT_PATH)
    string(TOUPPER ${LIB_NAME} UP_NAME)

    find_file(PATH_${UP_NAME} ${LIB_NAME} HINTS "${HINT_PATH}")

    if( ${PATH_${UP_NAME}} STREQUAL "PATH_${UP_NAME}-NOTFOUND")
        message(SEND_ERROR "${LIB_NAME} was not found in ${HINT_PATH}")
    else()
        list(APPEND DEPLOY_LIBS "${PATH_${UP_NAME}}")
    endif()
endmacro()

macro( AppendVCPKGLib LIB_NAME)
    AppendLib("${LIB_NAME}" "${VCPKG_ROOT}/installed/x64-windows/bin")
endmacro()

cmake_path(CONVERT "${QT_DIR}/bin" TO_CMAKE_PATH_LIST QT_DIR_LIB)
macro( AppendQt6Lib LIB_NAME)
    AppendLib("${LIB_NAME}" "${QT_DIR_LIB}")
endmacro()

# Force plugins to be installed near the exe.
install(SCRIPT ${deploy_script})

# Vcpkg DLLs
AppendVCPKGLib("abseil_dll.dll")
AppendVCPKGLib("cares.dll")
AppendVCPKGLib("libcrypto-3-x64.dll")
AppendVCPKGLib("libprotobuf.dll")
AppendVCPKGLib("libssl-3-x64.dll")
AppendVCPKGLib("re2.dll")
AppendVCPKGLib("sentry.dll")
AppendVCPKGLib("zlib1.dll")
# QML DLLs
AppendQt6Lib("Qt6QmlWorkerScript.dll")
AppendQt6Lib("Qt6Widgets.dll")
AppendQt6Lib("Qt6QuickControls2Impl.dll")
AppendQt6Lib("Qt6QuickLayouts.dll")
AppendQt6Lib("Qt6QuickDialogs2.dll")
AppendQt6Lib("Qt6QuickDialogs2QuickImpl.dll")
AppendQt6Lib("Qt6QuickDialogs2Utils.dll")

install(FILES ${DEPLOY_LIBS} DESTINATION "${CMAKE_INSTALL_PREFIX}")

# QML PlugIns
install(DIRECTORY ${QT_DIR}/qml/Qt/labs/platform DESTINATION "${CMAKE_INSTALL_PREFIX}/Qt/labs/")
install(DIRECTORY ${QT_DIR}/qml/QtQml DESTINATION "${CMAKE_INSTALL_PREFIX}")
install(DIRECTORY ${QT_DIR}/qml/QtQuick DESTINATION "${CMAKE_INSTALL_PREFIX}")

# crash handler utils
install(PROGRAMS "${VCPKG_INSTALLED_DIR}/${VCPKG_TARGET_TRIPLET}/tools/sentry-native/crashpad_handler.exe" DESTINATION "${CMAKE_INSTALL_PREFIX}")

# Runtime system libs
set(CMAKE_INSTALL_SYSTEM_RUNTIME_LIBS_SKIP TRUE)
include(InstallRequiredSystemLibraries)
install( PROGRAMS ${CMAKE_INSTALL_SYSTEM_RUNTIME_LIBS} DESTINATION ${CMAKE_INSTALL_PREFIX})


