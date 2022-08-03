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


include_guard()
set(VCPKG_ROOT "${CMAKE_CURRENT_SOURCE_DIR}/../../../extern/vcpkg")
include(bridge_setup.cmake)


#****************************************************************************************************************************************************
# Base project setup
#****************************************************************************************************************************************************
project(bridgepp LANGUAGES CXX)

set(CMAKE_CXX_STANDARD 20)
set(CMAKE_CXX_STANDARD_REQUIRED ON)

if (NOT DEFINED BRIDGE_APP_VERSION)
    message(FATAL_ERROR "BRIDGE_APP_VERSION is not defined.")
else()
    message(STATUS "Bridge version is ${BRIDGE_APP_VERSION}")
endif()

#****************************************************************************************************************************************************
# Qt
#****************************************************************************************************************************************************
if (NOT DEFINED ENV{QT6DIR})
    message(FATAL_ERROR "QT6DIR needs to be defined and point to the root of your Qt 6 folder (e.g. /Users/MyName/Qt/6.3.1/clang_64).")
endif ()
set(CMAKE_PREFIX_PATH $ENV{QT6DIR} ${CMAKE_PREFIX_PATH})
find_package(Qt6 COMPONENTS Core REQUIRED)
qt_standard_project_setup()

#****************************************************************************************************************************************************
# gRPC / Protobuf
#****************************************************************************************************************************************************
find_package(Protobuf CONFIG REQUIRED)
message(STATUS "Using protobuf ${Protobuf_VERSION}")

find_package(gRPC CONFIG REQUIRED)
message(STATUS "Using gRPC ${gRPC_VERSION}")

find_program(PROTOC_EXE protoc REQUIRED)
message(STATUS "protoc found ${PROTOC_EXE}")

message(STATUS "grpc_cpp_plugin ${grpc_cpp_plugin}")

find_program(GRPC_CPP_PLUGIN grpc_cpp_plugin)
if (GRPC_CPP_PLUGIN STREQUAL GRPC_CPP_PLUGIN-NOTFOUND)
    message(FATAL_ERROR "grpc_cpp_plugin exe could not be found. Please add it to your path. it should be located in \${VCPKG_ROOT}/installed/arm64-osx/tools/grpc")
else ()
    message(STATUS "grpc_cpp_plugin found at ${GRPC_CPP_PLUGIN}")
endif ()

set(PROTO_DIR "${CMAKE_CURRENT_SOURCE_DIR}/../grpc")
set(PROTO_FILE "${PROTO_DIR}/bridge.proto")
set(GRPC_OUT_DIR "bridgepp/GRPC")
set(PROTO_CPP_FILE "${GRPC_OUT_DIR}/bridge.pb.cc")
set(PROTO_H_FILE "${GRPC_OUT_DIR}/bridge.pb.h")
set(GRPC_CPP_FILE "${GRPC_OUT_DIR}/bridge.grpc.pb.cc")
set(GRPC_H_FILE "${GRPC_OUT_DIR}/bridge.grpc.pb.h")


#*****************************************************************************************************************************************************
# Source files and output
#*****************************************************************************************************************************************************

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


add_library(bridgepp
    bridgepp/BridgeUtils.cpp bridgepp/BridgeUtils.h
    bridgepp/Exception/Exception.h bridgepp/Exception/Exception.cpp
    bridgepp/GRPC/GRPCClient.cpp bridgepp/GRPC/GRPCClient.h
    bridgepp/GRPC/GRPCUtils.cpp bridgepp/GRPC/GRPCUtils.h
    ${PROTO_CPP_FILE} ${PROTO_H_FILE} ${GRPC_CPP_FILE} ${GRPC_H_FILE}
    bridgepp/Log/Log.h bridgepp/Log/Log.cpp
    bridgepp/User/User.cpp bridgepp/User/User.h
    bridgepp/Worker/Worker.h bridgepp/Worker/Overseer.h bridgepp/Worker/Overseer.cpp
    )

target_include_directories(bridgepp PUBLIC ${CMAKE_CURRENT_SOURCE_DIR})

target_link_libraries(bridgepp
    Qt6::Core
    protobuf::libprotobuf
    gRPC::grpc++
    )

target_precompile_headers(bridgepp PRIVATE Pch.h)
