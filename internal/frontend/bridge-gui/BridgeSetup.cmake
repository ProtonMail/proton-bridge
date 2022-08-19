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


include_guard()


if (NOT DEFINED BRIDGE_REPO_ROOT)
    message(FATAL_ERROR "BRIDGE_REPO_ROOT is not defined.")
endif()
message(STATUS "BRIDGE_REPO_ROOT is ${BRIDGE_REPO_ROOT}")

#*****************************************************************************************************************************************************
# Bridge version
#*****************************************************************************************************************************************************
if (NOT DEFINED BRIDGE_APP_VERSION)
    if (WIN32)
        find_program(POWERSHELL_EXE powershell.exe)
        if (NOT POWERSHELL_EXE)
            message(FATAL_ERROR "PowerShell could not be found.")
        endif()
        execute_process(COMMAND "${POWERSHELL_EXE}" -ExecutionPolicy Bypass "${BRIDGE_REPO_ROOT}/utils/bridge_app_version.ps1"
            OUTPUT_VARIABLE BRIDGE_APP_VERSION OUTPUT_STRIP_TRAILING_WHITESPACE COMMAND_ERROR_IS_FATAL ANY)
    else()
        execute_process(COMMAND "${BRIDGE_REPO_ROOT}/utils/bridge_app_version.sh"
            OUTPUT_VARIABLE BRIDGE_APP_VERSION OUTPUT_STRIP_TRAILING_WHITESPACE COMMAND_ERROR_IS_FATAL ANY)
    endif()
    if (NOT BRIDGE_APP_VERSION)
        message(FATAL_ERROR "Could not determine BRIDGE_APP_VERSION.")
    endif()
endif()


#****************************************************************************************************************************************************
# vcpkg, toolchain, and architecture
#****************************************************************************************************************************************************
# We rely on vcpkg for to get gRPC / Protobuf
# run build.sh / build.ps1 to get gRPC / Protobuf and dependencies installed.

set(VCPKG_ROOT "${BRIDGE_REPO_ROOT}/extern/vcpkg")
message(STATUS "VCPKG_ROOT is ${VCPKG_ROOT}")
if (WIN32)
    find_program(VCPKG_EXE "${VCPKG_ROOT}/vcpkg.exe")
else()
    find_program(VCPKG_EXE "${VCPKG_ROOT}/vcpkg")
endif()
if (NOT VCPKG_EXE)
    message(FATAL_ERROR "vcpkg is not installed. Run build.sh (macOS/Linux) or build.ps1 (Windows) first.")
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
    message(STATUS "Building for Intel x64 Windows computers")
    set(VCPKG_TARGET_TRIPLET x64-windows)
endif()

set(CMAKE_TOOLCHAIN_FILE "${VCPKG_ROOT}/scripts/buildsystems/vcpkg.cmake" CACHE STRING "toolchain")