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
include(../FindQt.cmake)
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

find_program(GRPC_CPP_PLUGIN grpc_cpp_plugin HINTS "${VCPKG_ROOT}/installed/arm64-osx/tools/grpc" "${VCPKG_ROOT}/installed/x64-osx/tools/grpc")
if (NOT GRPC_CPP_PLUGIN)
    message(FATAL_ERROR "grpc_cpp_plugin exe could not be found. Please add it to your path. it should be located in \${VCPKG_ROOT}/installed/arm64-osx/tools/grpc")
else ()
    message(STATUS "grpc_cpp_plugin found at ${GRPC_CPP_PLUGIN}")
endif ()


#*****************************************************************************************************************************************************
# Bridge gRPC service source code generation
#*****************************************************************************************************************************************************
set(PROTO_DIR "${CMAKE_CURRENT_SOURCE_DIR}/../../grpc")
set(PROTO_FILE "${PROTO_DIR}/bridge.proto")
set(GRPC_OUT_DIR "${CMAKE_CURRENT_SOURCE_DIR}/bridgepp/GRPC")
set(PROTO_CPP_FILE "${GRPC_OUT_DIR}/bridge.pb.cc")
set(PROTO_H_FILE "${GRPC_OUT_DIR}/bridge.pb.h")
set(GRPC_CPP_FILE "${GRPC_OUT_DIR}/bridge.grpc.pb.cc")
set(GRPC_H_FILE "${GRPC_OUT_DIR}/bridge.grpc.pb.h")

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
    COMMENT "Generating Bridge gPRC/Protobuf C++ code"
)


#*****************************************************************************************************************************************************
# Focus gRPC service source code generation
#*****************************************************************************************************************************************************
set(FOCUS_PROTO_DIR "${CMAKE_CURRENT_SOURCE_DIR}/../../../focus/proto")
set(FOCUS_PROTO_FILE "${FOCUS_PROTO_DIR}/focus.proto")
set(FOCUS_GRPC_OUT_DIR "${CMAKE_CURRENT_SOURCE_DIR}/bridgepp/FocusGRPC")
set(FOCUS_PROTO_CPP_FILE "${FOCUS_GRPC_OUT_DIR}/focus.pb.cc")
set(FOCUS_PROTO_H_FILE "${FOCUS_GRPC_OUT_DIR}/focus.pb.h")
set(FOCUS_GRPC_CPP_FILE "${FOCUS_GRPC_OUT_DIR}/focus.grpc.pb.cc")
set(FOCUS_GRPC_H_FILE "${FOCUS_GRPC_OUT_DIR}/focus.grpc.pb.h")

add_custom_command(
    OUTPUT
    ${FOCUS_PROTO_CPP_FILE}
    ${FOCUS_PROTO_H_FILE}
    ${FOCUS_GRPC_CPP_FILE}
    ${FOCUS_GRPC_H_FILE}
    COMMAND
    ${PROTOC_EXE}
    ARGS
    --proto_path=${FOCUS_PROTO_DIR}
    --plugin=protoc-gen-grpc="${GRPC_CPP_PLUGIN}"
    --cpp_out=${FOCUS_GRPC_OUT_DIR}
    --grpc_out=${FOCUS_GRPC_OUT_DIR}
    ${FOCUS_PROTO_FILE}
    DEPENDS
    ${FOCUS_PROTO_FILE}
    COMMENT "Generating focus gPRC/Protobuf Focus service C++ code"
)


#*****************************************************************************************************************************************************
# Source files and output
#*****************************************************************************************************************************************************


add_library(bridgepp
    bridgepp/BridgeUtils.cpp bridgepp/BridgeUtils.h
    bridgepp/Exception/Exception.h bridgepp/Exception/Exception.cpp
    bridgepp/GRPC/GRPCClient.cpp bridgepp/GRPC/GRPCClient.h
    bridgepp/GRPC/GRPCErrors.h bridgepp/GRPC/GRPCErrors.cpp
    bridgepp/GRPC/EventFactory.cpp bridgepp/GRPC/EventFactory.h
    bridgepp/GRPC/GRPCConfig.cpp bridgepp/GRPC/GRPCConfig.h
    bridgepp/GRPC/GRPCUtils.cpp bridgepp/GRPC/GRPCUtils.h
    ${PROTO_CPP_FILE} ${PROTO_H_FILE} ${GRPC_CPP_FILE} ${GRPC_H_FILE}
    ${FOCUS_PROTO_CPP_FILE} ${FOCUS_PROTO_H_FILE} ${FOCUS_GRPC_CPP_FILE} ${FOCUS_GRPC_H_FILE}
    bridgepp/FocusGRPC/FocusGRPCClient.cpp bridgepp/FocusGRPC/FocusGRPCClient.h
    bridgepp/Log/Log.h bridgepp/Log/Log.cpp
    bridgepp/Log/LogUtils.h bridgepp/Log/LogUtils.cpp
    bridgepp/ProcessMonitor.cpp bridgepp/ProcessMonitor.h
    bridgepp/User/User.cpp bridgepp/User/User.h
    bridgepp/Worker/Worker.h bridgepp/Worker/Overseer.h bridgepp/Worker/Overseer.cpp)

target_include_directories(bridgepp PUBLIC ${CMAKE_CURRENT_SOURCE_DIR})

target_link_libraries(bridgepp
    Qt6::Core
    protobuf::libprotobuf
    gRPC::grpc++
    )

target_precompile_headers(bridgepp PRIVATE Pch.h)

#*****************************************************************************************************************************************************
# GoogleTest
#*****************************************************************************************************************************************************

if (CMAKE_VERSION VERSION_GREATER_EQUAL "3.24.0")
    cmake_policy(SET CMP0135 NEW) # avoid warning DOWNLOAD_EXTRACT_TIMESTAMP
endif()

include(FetchContent)
FetchContent_Declare(
    googletest
    URL https://github.com/google/googletest/archive/b796f7d44681514f58a683a3a71ff17c94edb0c1.zip
)

# For Windows: Prevent overriding the parent project's compiler/linker settings
set(gtest_force_shared_crt ON CACHE BOOL "" FORCE)

FetchContent_MakeAvailable(googletest)

enable_testing()

#*****************************************************************************************************************************************************
# Tests
#*****************************************************************************************************************************************************
add_executable(bridgepp-test EXCLUDE_FROM_ALL
    Test/TestBridgeUtils.cpp
    Test/TestException.cpp
    Test/TestWorker.cpp Test/TestWorker.h)
add_dependencies(bridgepp-test bridgepp)
target_precompile_headers(bridgepp-test PRIVATE Pch.h)
target_link_libraries(bridgepp-test
    GTest::gtest_main
    bridgepp
    )

include(GoogleTest)
gtest_discover_tests(bridgepp-test)
