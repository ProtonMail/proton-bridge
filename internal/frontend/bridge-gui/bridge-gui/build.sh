#!/bin/bash
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

if [[ "$OSTYPE" == "msys" ]] || [[ "$OSTYPE" == "win32" ]] ; then
    Powershell.exe -File build.ps1 "$@"
    exit $?
fi


BRIDGE_REPO_ROOT="../../../.."
BRIDGE_INSTALL_PATH=${BRIDGE_INSTALL_PATH:-deploy}
BRIDGE_APP_VERSION=${BRIDGE_APP_VERSION:-$("${BRIDGE_REPO_ROOT}/utils/bridge_app_version.sh")}
BUILD_CONFIG=${BRIDGE_GUI_BUILD_CONFIG:-Debug}
BUILD_DIR=$(echo "./cmake-build-${BUILD_CONFIG}" | tr '[:upper:]' '[:lower:]')

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

VCPKG_ROOT="${BRIDGE_REPO_ROOT}/extern/vcpkg"

git submodule update --init --recursive ${VCPKG_ROOT}
check_exit "Failed to initialize vcpkg as a submodule."

echo submodule udpated

VCPKG_ROOT=$(realpath "$VCPKG_ROOT")
VCPKG_EXE="${VCPKG_ROOT}/vcpkg"
VCPKG_BOOTSTRAP="${VCPKG_ROOT}/bootstrap-vcpkg.sh"


${VCPKG_BOOTSTRAP} -disableMetrics
check_exit "Failed to bootstrap vcpkg."

if [[ "$OSTYPE" == "darwin"* ]]; then
    if [[ "$(uname -m)" == "arm64" ]]; then
        ${VCPKG_EXE} install grpc:arm64-osx --clean-after-build
        check_exit "Failed installing gRPC for macOS / Apple Silicon"
    fi
    ${VCPKG_EXE} install grpc:x64-osx --clean-after-build
    check_exit "Failed installing gRPC for macOS / Intel x64"
elif [[ "$OSTYPE" == "linux"* ]]; then
    ${VCPKG_EXE} install grpc:x64-linux --clean-after-build
    check_exit "Failed installing gRPC for Linux / Intel x64"
else
    echo "For Windows, use the build.ps1 Powershell script."
    exit 1
fi

${VCPKG_EXE} upgrade --no-dry-run

cmake  \
    -DCMAKE_BUILD_TYPE="${BUILD_CONFIG}" \
    -DBRIDGE_APP_VERSION="${BRIDGE_APP_VERSION}" \
    -G Ninja \
    -S . \
    -B "${BUILD_DIR}"
check_exit "CMake failed"

cmake --build "${BUILD_DIR}"
check_exit "build failed"

if [ "$1" == "install" ]; then
    cmake --install "${BUILD_DIR}"
    check_exit "install failed"
fi
