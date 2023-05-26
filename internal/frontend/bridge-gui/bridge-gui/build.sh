#!/bin/bash
# Copyright (c) 2023 Proton AG
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

if [[ "$OSTYPE" == "msys" ]] || [[ "$OSTYPE" == "win32" ]] ; then
    Powershell.exe -File build.ps1 "$@"
    exit $?
fi

set -x

realpath() {
    START_DIR=$PWD
    BASENAME="$(basename "$1")"
    cd "$(dirname "$1")" || exit
    LNK="$(readlink "$BASENAME")"
    while [ "$LNK" ]; do
        BASENAME="$(basename "$LNK")"
        cd "$(dirname "$LNK")" || exit
        LNK="$(readlink "$BASENAME")"
    done
    REALPATH="$PWD/$BASENAME"
    cd "$START_DIR" || exit
    echo "$REALPATH"
}

check_exit() {
    # shellcheck disable=SC2181
    if [ $? -ne 0 ]; then
        echo "Process failed: $1"
        rm -r "$BUILD_DIR"
        exit 1
    fi
}

BRIDGE_REPO_ROOT=$(realpath "../../../..")
BRIDGE_INSTALL_PATH=${BRIDGE_INSTALL_PATH:-deploy}
BRIDGE_APP_VERSION=${BRIDGE_APP_VERSION:-$("${BRIDGE_REPO_ROOT}/utils/bridge_app_version.sh")}
BRIDGE_APP_FULL_NAME=${BRIDGE_APP_FULL_NAME:-"Proton Mail Bridge"}
BRIDGE_VENDOR=${BRIDGE_VENDOR:-"Proton AG"}
BUILD_CONFIG=${BRIDGE_GUI_BUILD_CONFIG:-Debug}
BUILD_DIR=$(echo "./cmake-build-${BUILD_CONFIG}" | tr '[:upper:]' '[:lower:]')
VCPKG_ROOT="${BRIDGE_REPO_ROOT}/extern/vcpkg"
BRIDGE_REVISION=$(git rev-parse --short=10 HEAD)
BRIDGE_TAG=${BRIDGE_TAG:-"NOTAG"}
BRIDGE_DSN_SENTRY=${BRIDGE_DSN_SENTRY}
BRIDGE_BUILD_TIME=${BRIDGE_BUILD_TIME}
BRIDGE_BUILD_ENV= ${BRIDGE_BUILD_ENV:-"dev"}
git submodule update --init --recursive ${VCPKG_ROOT}
check_exit "Failed to initialize vcpkg as a submodule."

echo submodule udpated

VCPKG_EXE="${VCPKG_ROOT}/vcpkg"
VCPKG_BOOTSTRAP="${VCPKG_ROOT}/bootstrap-vcpkg.sh"


${VCPKG_BOOTSTRAP} -disableMetrics
check_exit "Failed to bootstrap vcpkg."

if [[ "$OSTYPE" == "darwin"* ]]; then
    ${VCPKG_EXE} install sentry-native:arm64-osx-min-11-0 grpc:arm64-osx-min-11-0 --overlay-triplets=vcpkg/triplets --clean-after-build
    check_exit "Failed installing gRPC for macOS / Apple Silicon"
    ${VCPKG_EXE} install sentry-native:x64-osx-min-10-15 grpc:x64-osx-min-10-15 --overlay-triplets=vcpkg/triplets --clean-after-build
    check_exit "Failed installing gRPC for macOS / Intel x64"
elif [[ "$OSTYPE" == "linux"* ]]; then
    ${VCPKG_EXE} install sentry-native:x64-linux grpc:x64-linux --clean-after-build
    check_exit "Failed installing gRPC for Linux / Intel x64"
else
    echo "For Windows, use the build.ps1 Powershell script."
    exit 1
fi

${VCPKG_EXE} upgrade --no-dry-run

if [[ "$OSTYPE" == "darwin"* ]]; then
  BRIDGE_CMAKE_MACOS_OPTS="-DCMAKE_OSX_ARCHITECTURES=${BRIDGE_MACOS_ARCH:-$(uname -m)}"
else
  BRIDGE_CMAKE_MACOS_OPTS=""
fi

cmake  \
    -DCMAKE_BUILD_TYPE="${BUILD_CONFIG}" \
    -DBRIDGE_APP_FULL_NAME="${BRIDGE_APP_FULL_NAME}" \
    -DBRIDGE_VENDOR="${BRIDGE_VENDOR}" \
    -DBRIDGE_REVISION="${BRIDGE_REVISION}" \
    -DBRIDGE_TAG="${BRIDGE_TAG}" \
    -DBRIDGE_DSN_SENTRY="${BRIDGE_DSN_SENTRY}" \
    -DBRIDGE_BRIDGE_TIME="${BRIDGE_BRIDGE_TIME}" \
    -DBRIDGE_BUILD_ENV="${BRIDGE_BUILD_ENV}" \
    -DBRIDGE_APP_VERSION="${BRIDGE_APP_VERSION}" "${BRIDGE_CMAKE_MACOS_OPTS}" \
    -G Ninja \
    -S . \
    -B "${BUILD_DIR}"
check_exit "CMake failed"

cmake --build "${BUILD_DIR}" -v
check_exit "build failed"

if [ "$1" == "install" ]; then
    cmake --install "${BUILD_DIR}" -v
    check_exit "install failed"
fi
