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
    Tabs/EventsTab.cpp Tabs/EventsTab.h
    Tabs/KnowledgeBaseTab.cpp Tabs/KnowledgeBaseTab.h
    Tabs/SettingsTab.cpp Tabs/SettingsTab.h
    Tabs/UsersTab.cpp Tabs/UsersTab.h
    UserDialog.cpp UserDialog.h
    UserTable.cpp UserTable.h
)

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
