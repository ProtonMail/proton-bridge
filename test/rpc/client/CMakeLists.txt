cmake_minimum_required(VERSION 3.22)

set(VCPKG_TARGET_TRIPLET arm64-osx)
set(CMAKE_TOOLCHAIN_FILE "$ENV{VCPKG_ROOT}/scripts/buildsystems/vcpkg.cmake")

project(BridgeRpcClientTest VERSION 0.1 LANGUAGES CXX)

if (NOT DEFINED ENV{QTDIR})
   message(FATAL_ERROR "QTDIR needs to be defined and point to the root of your Qt folder. (e.g. /Users/MyName/Qt/6.3/clang_64).")
endif()

set(CMAKE_CXX_STANDARD 17)
set(CMAKE_AUTOMOC ON)
set(CMAKE_AUTORCC ON)
set(CMAKE_AUTOMATIC ON)

find_package(Protobuf CONFIG REQUIRED)
message(STATUS "Using protobuf ${Protobuf_VERSION}")

find_package(gRPC CONFIG REQUIRED)
message(STATUS "Using gRPC ${gRPC_VERSION}")

find_package(Qt6 COMPONENTS Core Test REQUIRED)

find_program(PROTOC_EXE protoc)
find_program(GRPC_CPP_PLUGIN grpc_cpp_plugin)
set(PROTO_DIR "${CMAKE_CURRENT_SOURCE_DIR}/../../../internal/grpc")
set(PROTO_FILE "${PROTO_DIR}/bridge.proto")
set(GRPC_OUT_DIR "${CMAKE_CURRENT_SOURCE_DIR}/grpc")
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
   COMMENT "Generating gPRC/Protobuf C++ code"
)

add_executable(BridgeRpcClientTest
   BridgeRpcClientTest.cpp
   RpcClient.cpp RpcClient.h
   Exception.cpp Exception.h
   ${PROTO_CPP_FILE}
   ${PROTO_H_FILE}
   ${GRPC_CPP_FILE}
   ${GRPC_H_FILE}
   )

target_link_libraries(BridgeRpcClientTest
   Qt::Core Qt::Test
   protobuf::libprotobuf
   gRPC::grpc++ gRPC::grpc++_reflection
   )

