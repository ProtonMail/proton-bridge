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

$scriptpath = $MyInvocation.MyCommand.Path
$scriptDir = Split-Path $scriptpath
$bridgeRepoRootDir = Join-Path $scriptDir "../../../.." -Resolve
Write-host "Bridge-gui directory is $scriptDir"
Write-host "Bridge repos root dir $bridgeRepoRootDir"
Push-Location $scriptDir

$ErrorActionPreference = "Stop"

$cmakeExe=$(Get-Command cmake).source
if ($null -eq $cmakeExe)
{
    $cmakeExe = "C:\Program Files\Microsoft Visual Studio\2022\Community\Common7\IDE\CommonExtensions\Microsoft\CMake\CMake\bin\cmake.exe" # Hardcoded for now.
}
Write-host "CMake found here : $cmakeExe"
$cmake_version = . $cmakeExe --version
Write-host "CMake version : $cmake_version"

$bridgeVersion = ($env:BRIDGE_APP_VERSION)
if ($null -eq $bridgeVersion)
{
    $bridgeVersion = . (Join-Path $bridgeRepoRootDir "utils/bridge_app_version.ps1")
}

$bridgeFullName = ($env:BRIDGE_APP_FULL_NAME)
if ($null -eq $bridgeFullName)
{
    $bridgeFullName = "Proton Mail Bridge"
}

$bridgeVendor = ($env:BRIDGE_VENDOR)
if ($null -eq $bridgeVendor)
{
    $bridgeVendor = "Proton AG"
}

$buildConfig = ($env:BRIDGE_GUI_BUILD_CONFIG)
if ($null -eq $buildConfig)
{
    $buildConfig =  "Debug"
}

$buildDir=(Join-Path $scriptDir "cmake-build-$buildConfig".ToLower())
$vcpkgRoot = (Join-Path $bridgeRepoRootDir "extern/vcpkg" -Resolve)
$vcpkgExe = (Join-Path $vcpkgRoot "vcpkg.exe")
$vcpkgBootstrap = (Join-Path $vcpkgRoot "bootstrap-vcpkg.bat")

function check_exit() {
    if ($? -ne $True)
    {
        Write-Host "Process failed: $args[0] : $?"
        Remove-Item "$buildDir" -Recurse -ErrorAction Ignore
        exit 1
    }
}

Write-host "Running build for version $bridgeVersion - $buildConfig in $buildDir"

$REVISION_HASH = git rev-parse --short=10 HEAD
$bridgeTag = ($env:BRIDGE_TAG)
$bridgeDsnSentry = ($env:BRIDGE_DSN_SENTRY)
$bridgeBuidTime = ($env:BRIDGE_BUILD_TIME)

$bridgeBuildEnv = ($env:BRIDGE_BUILD_ENV)
if ($null -eq $bridgeBuildEnv)
{
    $bridgeBuildEnv =  "dev"
}

git submodule update --init --recursive $vcpkgRoot
. $vcpkgBootstrap -disableMetrics
. $vcpkgExe install sentry-native:x64-windows grpc:x64-windows --clean-after-build
. $vcpkgExe upgrade --no-dry-run
. $cmakeExe -G "Visual Studio 17 2022" -DCMAKE_BUILD_TYPE="$buildConfig" `
                                       -DBRIDGE_APP_FULL_NAME="$bridgeFullName" `
                                       -DBRIDGE_VENDOR="$bridgeVendor" `
                                       -DBRIDGE_REVISION="$REVISION_HASH" `
                                       -DBRIDGE_TAG="$bridgeTag" `
                                       -DBRIDGE_APP_VERSION="$bridgeVersion" `
                                       -DBRIDGE_BUILD_TIME="$bridgeBuidTime" `
                                       -DBRIDGE_DSN_SENTRY="$bridgeDsnSentry" `
                                       -DBRIDGE_BUILD_ENV="$bridgeBuildEnv" `
                                       -S . -B $buildDir

check_exit "CMake failed"
. $cmakeExe --build $buildDir --config "$buildConfig"
check_exit "Build failed"

if  ($($args.count) -gt 0 )
{
    if ($args[0] = "install")
    {
        . $cmakeExe --install $buildDir
        check_exit "Install failed"
    }
}

Pop-Location
