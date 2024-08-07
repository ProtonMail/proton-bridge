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
set(BRIDGE_REPO_ROOT "${CMAKE_CURRENT_SOURCE_DIR}/../../../..")
include(../BridgeSetup.cmake)


#*****************************************************************************************************************************************************
# Project
#*****************************************************************************************************************************************************


project(bridge-gui LANGUAGES CXX)

if (NOT DEFINED BRIDGE_APP_FULL_NAME)
    message(FATAL_ERROR "BRIDGE_APP_FULL_NAME is not defined.")
else()
    message(STATUS "App name is ${BRIDGE_APP_FULL_NAME}")
endif()

if (NOT DEFINED BRIDGE_VENDOR)
    message(FATAL_ERROR "BRIDGE_VENDOR is not defined.")
else()
    message(STATUS "App vendor is ${BRIDGE_VENDOR}")
endif()

if (NOT DEFINED BRIDGE_APP_VERSION)
    message(FATAL_ERROR "BRIDGE_APP_VERSION is not defined.")
else()
    message(STATUS "Bridge version is ${BRIDGE_APP_VERSION}")
endif()

if (APPLE) # On macOS, we have some Objective-C++ code in DockIcon to deal with the dock icon.
    enable_language(OBJC OBJCXX)
endif()

set(CMAKE_CXX_STANDARD 20)
set(CMAKE_CXX_STANDARD_REQUIRED ON)

if (APPLE) # We need to link the Cocoa framework for the dock icon.
    find_library(COCOA_LIBRARY Cocoa REQUIRED)
endif()


#*****************************************************************************************************************************************************
# Qt
#*****************************************************************************************************************************************************
include(../FindQt.cmake)

# Use CMAKE_INSTALL_PREFIX that is also used internally by CMake
if (DEFINED ENV{BRIDGE_INSTALL_PATH})
    set(CMAKE_INSTALL_PREFIX "$ENV{BRIDGE_INSTALL_PATH}")
else(DEFINED ENV{BRIDGE_INSTALL_PATH})
    message(STATUS "Using Default install path (${CMAKE_INSTALL_PREFIX}), export BRIDGE_INSTALL_PATH to change it.")
endif(DEFINED ENV{BRIDGE_INSTALL_PATH})

if(NOT UNIX)
    # To change the value of QT_DEPLOY_BIN_DIR, ensure that the project sets CMAKE_INSTALL_BINDIR before the Core package is found.
    set(CMAKE_INSTALL_BINDIR ".")
endif(NOT UNIX)

find_package(Qt6 COMPONENTS Core Quick Qml QuickControls2 Widgets Svg Gui REQUIRED)
qt_standard_project_setup()
set(CMAKE_AUTORCC ON)
message(STATUS "Using Qt ${Qt6_VERSION}")

#*****************************************************************************************************************************************************
# Sentry Native
#*****************************************************************************************************************************************************
find_package(sentry CONFIG REQUIRED)

#*****************************************************************************************************************************************************
# Source files and output
#*****************************************************************************************************************************************************

configure_file(${CMAKE_CURRENT_SOURCE_DIR}/BuildConfig.h.in ${CMAKE_CURRENT_SOURCE_DIR}/BuildConfig.h)

if (NOT TARGET bridgepp)
    add_subdirectory(../bridgepp bridgepp)
endif()

if (APPLE)
    set(DOCK_ICON_SRC_FILE MacOS/DockIcon.mm)
else()
    set(DOCK_ICON_SRC_FILE MacOS/DockIcon.cpp)
endif()

if(UNIX)
    list(APPEND CMAKE_INSTALL_RPATH "$ORIGIN/lib" )
endif(UNIX)

add_executable(bridge-gui
    Resources.qrc
    AppController.cpp AppController.h
    BridgeApp.cpp BridgeApp.h
    BuildConfig.h
    CommandLine.cpp CommandLine.h
    EventStreamWorker.cpp EventStreamWorker.h
    LogUtils.cpp LogUtils.h
    main.cpp
    TrayIcon.cpp TrayIcon.h
    Pch.h
    QMLBackend.cpp QMLBackend.h
    UserList.cpp UserList.h
    SentryUtils.cpp SentryUtils.h
    Settings.cpp Settings.h
    ClipboardProxy.cpp ClipboardProxy.h
    ${DOCK_ICON_SRC_FILE} MacOS/DockIcon.h
    )

if (APPLE)
    target_sources(bridge-gui PRIVATE MacOS/SecondInstance.mm MacOS/SecondInstance.h)
endif(APPLE)

if (WIN32) # on Windows, we add a (non-Qt) resource file that contains the application icon and version information.
    string(TIMESTAMP BRIDGE_BUILD_YEAR "%Y")
    set(REGEX_NUMBER "[0123456789]") # CMake matches does not support \d.
    if (${BRIDGE_APP_VERSION} MATCHES "^(${REGEX_NUMBER}+)\\.(${REGEX_NUMBER}+)\\.(${REGEX_NUMBER}+)")
        set(BRIDGE_APP_VERSION_COMMA "${CMAKE_MATCH_1},${CMAKE_MATCH_2},${CMAKE_MATCH_3},0")
    else()
        message(FATAL_ERROR "Could not extract comma-separated version number from ${BRIDGE_APP_VERSION}")
    endif()
    configure_file("${CMAKE_CURRENT_SOURCE_DIR}/Resources.rc.in" "${CMAKE_CURRENT_SOURCE_DIR}/Resources.rc")
    target_sources(bridge-gui PRIVATE Resources.rc)
endif()

target_precompile_headers(bridge-gui PRIVATE Pch.h)
target_include_directories(bridge-gui PRIVATE "${CMAKE_CURRENT_SOURCE_DIR}" ${SENTRY_CONFIG_GENERATED_FILE_DIR})
target_link_libraries(bridge-gui
    Qt6::Widgets
    Qt6::Core
    Qt6::Quick
    Qt6::Qml
    Qt6::QuickControls2
    Qt6::Svg
    Qt6::Gui
    sentry::sentry
    bridgepp
    )

if (APPLE)
    target_link_libraries(bridge-gui ${COCOA_LIBRARY})
endif()

#*****************************************************************************************************************************************************
# Deploy
#*****************************************************************************************************************************************************
set( CMAKE_EXPORT_COMPILE_COMMANDS ON )

set_target_properties(bridge-gui PROPERTIES
        WIN32_EXECUTABLE TRUE
        MACOSX_BUNDLE TRUE)

install(TARGETS bridge-gui
        RUNTIME DESTINATION "${CMAKE_INSTALL_PREFIX}"
        BUNDLE DESTINATION  "${CMAKE_INSTALL_PREFIX}"
        LIBRARY DESTINATION "${CMAKE_INSTALL_PREFIX}"
        )

qt_generate_deploy_app_script(
        TARGET bridge-gui
        FILENAME_VARIABLE deploy_script
        NO_UNSUPPORTED_PLATFORM_ERROR)

if(UNIX AND NOT APPLE)
    set(DEPLOY_OS Linux)
elseif(APPLE)
    set(DEPLOY_OS Darwin)
else()
    set(DEPLOY_OS Windows)
endif()

include(Deploy${DEPLOY_OS}.cmake)
