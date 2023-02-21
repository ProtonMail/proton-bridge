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


project(bridge-gui-tester LANGUAGES CXX)

if (NOT DEFINED BRIDGE_APP_VERSION)
    message(FATAL_ERROR "BRIDGE_APP_VERSION is not defined.")
else()
    message(STATUS "Bridge version is ${BRIDGE_APP_VERSION}")
endif()

set(CMAKE_CXX_STANDARD 20)
set(CMAKE_CXX_STANDARD_REQUIRED ON)


#*****************************************************************************************************************************************************
# Qt
#*****************************************************************************************************************************************************
include(../FindQt.cmake)
find_package(Qt6 COMPONENTS Core Gui Widgets Qml REQUIRED)
qt_standard_project_setup()
message(STATUS "Using Qt ${Qt6_VERSION}")


#*****************************************************************************************************************************************************
# bridgelib
#*****************************************************************************************************************************************************
find_program(GO_BIN "go")
if (NOT GO_BIN)
    message(FATAL_ERROR "Could not location go compiler")
endif()
message(STATUS "go compiler is ${GO_BIN}")

if (APPLE) # set some env variable for go compiler on macOS. Note the CGO_ENABLED=1 is required when cross-compiling.
    if (CMAKE_OSX_ARCHITECTURES STREQUAL "arm64")
        set(GO_BIN "MACOSX_DEPLOYMENT_TARGET=11.0" "GOARCH=arm64" "CGO_CFLAGS=\"-mmacosx-version-min=11.0\"" CGO_ENABLED=1 ${GO_BIN})
    else ()
        set(GO_BIN "MACOSX_DEPLOYMENT_TARGET=10.15" "GOARCH=amd64" "CGO_CFLAGS=\"-mmacosx-version-min=10.15\"" CGO_ENABLED=1 ${GO_BIN})
    endif()
endif()

file(REAL_PATH "pkg/bridgelib" BRIDGELIB_DIR BASE_DIRECTORY  "${BRIDGE_REPO_ROOT}")
message(STATUS "bridgelib folder is ${BRIDGELIB_DIR}")

set(BRIDGELIB_OUTPUT_DIR ${CMAKE_CURRENT_BINARY_DIR})
set(BRIDGELIB_BASE_NAME "bridgelib")
if (UNIX AND NOT APPLE)
    set(BRIDGELIB_LIB_FILE "${BRIDGELIB_BASE_NAME}.so")
endif()
if (APPLE)
    set(BRIDGELIB_LIB_FILE "${BRIDGELIB_BASE_NAME}.dylib")
endif()
if (WIN32)
    set(BRIDGELIB_LIB_FILE "${BRIDGELIB_BASE_NAME}.dll")
endif()

set(BRIDGELIB_OUTPUT_PATH "${BRIDGELIB_OUTPUT_DIR}/${BRIDGELIB_LIB_FILE}")

add_custom_target(
    bridgelib
    COMMAND ${GO_BIN} build -o ${BRIDGELIB_OUTPUT_PATH} --buildmode c-shared
    WORKING_DIRECTORY ${BRIDGELIB_DIR}
    COMMENT "Compile bridgelib library"
)


#*****************************************************************************************************************************************************
# Source files and output
#*****************************************************************************************************************************************************

if (NOT TARGET bridgepp)
add_subdirectory(../bridgepp bridgepp)
endif()

add_executable(bridge-gui-tester
    AppController.cpp AppController.h
    Cert.cpp Cert.h
    main.cpp
    MainWindow.cpp MainWindow.h
    GRPCMetaDataProcessor.cpp GRPCMetaDataProcessor.h
    GRPCQtProxy.cpp GRPCQtProxy.h
    GRPCService.cpp GRPCService.h
    GRPCServerWorker.cpp GRPCServerWorker.h
    Tabs/SettingsTab.cpp Tabs/SettingsTab.h
    Tabs/UsersTab.cpp Tabs/UsersTab.h
    UserDialog.cpp UserDialog.h
    UserTable.cpp UserTable.h
    )

add_dependencies(bridge-gui-tester bridgelib)

target_precompile_headers(bridge-gui-tester PRIVATE Pch.h)
target_include_directories(bridge-gui-tester PRIVATE ${CMAKE_CURRENT_SOURCE_DIR})
target_compile_definitions(bridge-gui-tester PRIVATE BRIDGE_APP_VERSION=\"${BRIDGE_APP_VERSION}\")


target_link_libraries(bridge-gui-tester
    Qt6::Core
    Qt6::Gui
    Qt6::Widgets
    Qt6::Qml
    bridgepp
    )
