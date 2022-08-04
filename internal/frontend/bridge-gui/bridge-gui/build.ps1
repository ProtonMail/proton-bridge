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

#!/bin/bash

$ErrorActionPreference = "Stop"

$cmakeExe = "C:\Program Files\Microsoft Visual Studio\2022\Community\Common7\IDE\CommonExtensions\Microsoft\CMake\CMake\bin\cmake.exe" # Hardcoded for now.
$bridgeVersion = ($env:BRIDGE_APP_VERSION ??= "2.2.1+") # TODO get the version number from a unified location.
$buildConfig = ($env:BRIDGE_GUI_BUILD_CONFIG ??= "Debug")
$buildDir=(Join-Path $PSScriptRoot "cmake-build-$buildConfig-visual-studio".ToLower())
$vcpkgRoot = (Join-Path $PSScriptRoot "../../../../extern/vcpkg" -Resolve)
$vcpkgExe = (Join-Path $vcpkgRoot "vcpkg.exe")
$vcpkgBootstrap = (Join-Path $vcpkgRoot "bootstrap-vcpkg.bat")

git submodule update --init --recursive $vcpkgRoot
. $vcpkgBootstrap -disableMetrics
. $vcpkgExe install grpc:x64-windows
. $vcpkgExe upgrade --no-dry-run 
. $cmakeExe -G "Visual Studio 17 2022" -DCMAKE_BUILD_TYPE="$buildConfig" -DBRIDGE_APP_VERSION="$bridgeVersion" -S . -B $buildDir
. $cmakeExe --build $buildDir
