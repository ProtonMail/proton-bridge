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

set(CMAKE_INSTALL_RPATH "${CMAKE_INSTALL_BINDIR}" "${CMAKE_INSTALL_LIBDIR}" "." "../lib")
install(DIRECTORY "$ENV{QT6DIR}/qml" "$ENV{QT6DIR}/plugins"
        DESTINATION "${CMAKE_INSTALL_PREFIX}")

macro( AppendLib LIB_NAME HINT_PATH)
    string(TOUPPER ${LIB_NAME} UP_NAME)

    find_library(PATH_${UP_NAME} ${LIB_NAME} HINTS "${HINT_PATH}")

    if( ${PATH_${UP_NAME}} STREQUAL "PATH_${UP_NAME}-NOTFOUND")
        message(SEND_ERROR "${LIB_NAME} was not found in ${HINT_PATH}")
    else()
        get_filename_component(REAL_PATH_${UP_NAME} ${PATH_${UP_NAME}} REALPATH)
        list(APPEND DEPLOY_LIBS ${PATH_${UP_NAME}})
        list(APPEND DEPLOY_LIBS ${REAL_PATH_${UP_NAME}})
    endif()
endmacro()

macro( AppendQt6Lib LIB_NAME)
    AppendLib("${LIB_NAME}" "$ENV{QT6DIR}/lib/")
endmacro()

#Qt6
AppendQt6Lib("libQt6QuickControls2.so.6")
AppendQt6Lib("libQt6Quick.so.6")
AppendQt6Lib("libQt6QmlModels.so.6")
AppendQt6Lib("libQt6Qml.so.6")
AppendQt6Lib("libQt6Network.so.6")
AppendQt6Lib("libQt6OpenGL.so.6")
AppendQt6Lib("libQt6Gui.so.6")
AppendQt6Lib("libQt6Core.so.6")
AppendQt6Lib("libQt6QuickTemplates2.so.6")
AppendQt6Lib("libQt6DBus.so.6")
AppendQt6Lib("libicui18n.so.56")
AppendQt6Lib("libicuuc.so.56")
AppendQt6Lib("libicudata.so.56")
AppendQt6Lib("libQt6XcbQpa.so.6")
# QML dependencies
AppendQt6Lib("libQt6QmlWorkerScript.so.6")
AppendQt6Lib("libQt6Widgets.so.6")
AppendQt6Lib("libQt6QuickControls2Impl.so.6")
AppendQt6Lib("libQt6QuickLayouts.so.6")
AppendQt6Lib("libQt6QuickDialogs2.so.6")
AppendQt6Lib("libQt6QuickDialogs2QuickImpl.so.6")
AppendQt6Lib("libQt6QuickDialogs2Utils.so.6")
AppendQt6Lib("libQt6Svg.so.6")

install(FILES ${DEPLOY_LIBS} DESTINATION "${CMAKE_INSTALL_PREFIX}/lib")
