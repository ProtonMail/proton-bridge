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

#We rely on vcpkg for to get gRPC+Protobuf
if (NOT DEFINED ENV{VCPKG_ROOT})
    message(FATAL_ERROR "vcpkg is required. Install vcpkg and define VCPKG_ROOT to point the the vcpkg installation folder. (e.g. ~/vcpkg/")
endif()


# For now we support only a single architecture for macOS (ARM64 or x86_64). We need to investigate how to build universal binaries with vcpkg.
if (APPLE)
    if (NOT DEFINED CMAKE_OSX_ARCHITECTURES)
        execute_process(COMMAND "uname" "-m" OUTPUT_VARIABLE UNAME_RESULT OUTPUT_STRIP_TRAILING_WHITESPACE)
        set(CMAKE_OSX_ARCHITECTURES ${UNAME_RESULT} CACHE STRING "osx_architectures")
    endif()

    if (CMAKE_OSX_ARCHITECTURES STREQUAL "arm64")
        message(STATUS "Building for Apple Silicon Mac computers")
        set(VCPKG_TARGET_TRIPLET arm64-osx)
    elseif (CMAKE_OSX_ARCHITECTURES STREQUAL "x86_64")
        message(STATUS "Building for Intel based Mac computers")
        set(VCPKG_TARGET_TRIPLET x64-osx)
    else ()
        message(FATAL_ERROR "Unknown value for CMAKE_OSX_ARCHITECTURE. Please use one of \"arm64\" and \"x86_64\". Multiple architectures are not supported.")
    endif ()
endif()

if  (WIN32)
    set(VCPKG_TARGET_TRIPLET x64-mingw-static)
endif()
set(CMAKE_TOOLCHAIN_FILE "$ENV{VCPKG_ROOT}/scripts/buildsystems/vcpkg.cmake" CACHE STRING "toolchain")


project(bridge-gui LANGUAGES CXX)
if (APPLE) # On macOS, we have some Objective-C++ code in DockIcon to deal with the dock icon.
    enable_language(OBJC OBJCXX)
endif()

if (NOT DEFINED ENV{QT6DIR})
    message(FATAL_ERROR "QT6DIR needs to be defined and point to the root of your Qt 6 folder (e.g. /Users/MyName/Qt/6.3.1/clang_64).")
endif()

set(CMAKE_PREFIX_PATH $ENV{QT6DIR} ${CMAKE_PREFIX_PATH})
set(CMAKE_CXX_STANDARD 20)
set(CMAKE_CXX_STANDARD_REQUIRED ON)
set(CMAKE_AUTOMOC ON)
set(CMAKE_AUTORCC ON)
set(CMAKE_AUTOUIC ON)
find_package(Protobuf CONFIG REQUIRED)
message(STATUS "Using protobuf ${Protobuf_VERSION}")


find_package(gRPC CONFIG REQUIRED)
message(STATUS "Using gRPC ${gRPC_VERSION}")


if (APPLE) # We need to link the Cocoa framework for the dock icon.
    find_library(COCOA_LIBRARY Cocoa REQ UIRED)
endif()


find_package(Qt6 COMPONENTS
    Core
    Quick
    Qml
    QuickControls2
    REQUIRED)
message(STATUS "Using Qt ${Qt5_VERSION}")

find_program(PROTOC_EXE protoc REQUIRED)
message(STATUS "protoc found ${PROTOC_EXE}")

message(STATUS "grpc_cpp_plugin ${grpc_cpp_plugin}")

find_program(GRPC_CPP_PLUGIN grpc_cpp_plugin)
if(GRPC_CPP_PLUGIN STREQUAL GRPC_CPP_PLUGIN-NOTFOUND)
message(FATAL_ERROR "grpc_cpp_plugin exe could not be found. Please add it to your path. it should be located in \${VCPKG_ROOT}/installed/arm64-osx/tools/grpc")
else()
    message(STATUS "grpc_cpp_plugin found at ${GRPC_CPP_PLUGIN}")
endif()

set(PROTO_DIR  "${CMAKE_CURRENT_SOURCE_DIR}/../grpc")
set(PROTO_FILE "${PROTO_DIR}/bridge.proto")
set(GRPC_OUT_DIR "${CMAKE_CURRENT_SOURCE_DIR}/GRPC")
set(PROTO_CPP_FILE "${GRPC_OUT_DIR}/bridge.pb.cc")
set(PROTO_H_FILE "${GRPC_OUT_DIR}/bridge.pb.h")
set(GRPC_CPP_FILE "${GRPC_OUT_DIR}/bridge.grpc.pb.cc")
set(GRPC_H_FILE "${GRPC_OUT_DIR}/bridge.grpc.pb.h")
if (APPLE)
    set(DOCK_ICON_SRC_FILE DockIcon/DockIcon.mm)
else()
    set(DOCK_ICON_SRC_FILE DockIcon/DockIcon.cpp)
endif()

add_custom_command(
    OUTPUT
    ${PROTO_CPP_FILE}
    ${PROTO_H_FILE}
    ${GRPC_CPP_FILE}
    ${GRPC_H_FILE}
    COMMAND
    ${PROTOC_EXE}
    ARGS
    --proto_path=${PROTO_DIR}
    --plugin=protoc-gen-grpc="${GRPC_CPP_PLUGIN}"
    --cpp_out=${GRPC_OUT_DIR}
    --grpc_out=${GRPC_OUT_DIR}
    ${PROTO_FILE}
    DEPENDS
    ${PROTO_FILE}
    COMMENT "Generating gPRC/Protobuf C++ code"
)

add_executable(bridge-gui
    Resources.qrc
    ${PROTO_CPP_FILE} ${PROTO_H_FILE} ${GRPC_CPP_FILE} ${GRPC_H_FILE}
    AppController.cpp AppController.h
    BridgeMonitor.cpp BridgeMonitor.h
    EventStreamWorker.cpp EventStreamWorker.h
    Exception.cpp Exception.h
    Log.cpp Log.h
    main.cpp
    Pch.h
    QMLBackend.cpp QMLBackend.h
    ${DOCK_ICON_SRC_FILE} DockIcon/DockIcon.h
    GRPC/GRPCClient.cpp GRPC/GRPCClient.h
    GRPC/GRPCUtils.cpp GRPC/GRPCUtils.h
    User/User.cpp User/User.h User/UserList.cpp User/UserList.h
    Worker/Overseer.cpp Worker/Overseer.h
    Worker/Worker.h)

include_directories(${CMAKE_CURRENT_SOURCE_DIR})

target_precompile_headers(bridge-gui PRIVATE Pch.h)

target_link_libraries(bridge-gui
    Qt6::Core
    Qt6::Quick
    Qt6::Qml
    Qt6::QuickControls2
    protobuf::libprotobuf
    gRPC::grpc++
    )

if (APPLE)
target_link_libraries(bridge-gui ${COCOA_LIBRARY})
endif()
