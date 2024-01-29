# Copyright (c) 2024 Proton AG
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

# Remove-Bridge is a script for quitting, uninstalling, and deleted all Bridge related files on Windows

# Define variables with path to Bridge files (vault, cache, startup entry etc)
$RoamProtonmail = "$env:APPDATA\protonmail"
$RoamProtonAG = "$env:APPDATA\Proton AG"
$LocalProtonmail = "$env:LOCALAPPDATA\protonmail"
$LocalProtonAG = "$env:LOCALAPPDATA\Proton AG"
$StartUpProtonBridge = "$env:APPDATA\Microsoft\Windows\Start Menu\Programs\Startup\Proton Mail Bridge.lnk"

function Uninstall-PMBridge {
    # Uninstalling REBRANDED version of Bridge
    # Find the UninstallSTring in the registry (64bit & 32bit)
    # Use the UninstallString with `msiexec.exe` to uninstall Bridge

    $registry64 = Get-ChildItem "HKLM:\SOFTWARE\Microsoft\Windows\CurrentVersion\Uninstall" | ForEach-Object { Get-ItemProperty $_.PSPath } | Where-Object { $_ -match "Proton Mail Bridge" } | Select-Object UninstallString

    if ($registry64) {
        $registry64 = $registry64 | Select-Object -Last 1
        $registry64 = $registry64.UninstallString -Replace "msiexec.exe","" -Replace "/I","" -Replace "/X",""
        $registry64 = $registry64.Trim()
        Start-Process "msiexec.exe" -arg "/X $registry64 /passive" -Wait
    }

    $registry32 = Get-ChildItem "HKLM:\SOFTWARE\Wow6432Node\Microsoft\Windows\CurrentVersion\Uninstall" | ForEach-Object { Get-ItemProperty $_.PSPath } | Where-Object { $_ -match "Proton Mail Bridge" } | Select-Object UninstallString

    if ($registry32) {
        $registry32 = $registry32 | Select-Object -Last 1
        $registry32 = $registry32.UninstallString -Replace "msiexec.exe","" -Replace "/I","" -Replace "/X",""
        $registry32 = $registry32.Trim()
        Start-Process "msiexec.exe" -arg "/X $registry32 /passive" -Wait
    }


    # Uninstalling PRE-REBRANDED version of Bridge
    # Find the UninstallSTring in the registry (64bit & 32bit)
    # Use the UninstallString with `msiexec.exe` to uninstall Bridge

    $preRebrandRegistry64 = Get-ChildItem "HKLM:\SOFTWARE\Microsoft\Windows\CurrentVersion\Uninstall" | ForEach-Object { Get-ItemProperty $_.PSPath } | Where-Object { $_ -match "ProtonMail Bridge" } | Select-Object UninstallString

    if ($preRebrandRegistry64) {
        $preRebrandRegistry64 = $preRebrandRegistry64 | Select-Object -Last 1
        $preRebrandRegistry64 = $preRebrandRegistry64.UninstallString -Replace "msiexec.exe","" -Replace "/I","" -Replace "/X",""
        $preRebrandRegistry64 = $preRebrandRegistry64.Trim()
        Start-Process "msiexec.exe" -arg "/X $preRebrandRegistry64 /passive" -Wait
    }

    $preRebrandRegistry32 = Get-ChildItem "HKLM:\SOFTWARE\Wow6432Node\Microsoft\Windows\CurrentVersion\Uninstall" | ForEach-Object { Get-ItemProperty $_.PSPath } | Where-Object { $_ -match "ProtonMail Bridge" } | Select-Object UninstallString

    if ($preRebrandRegistry32) {
        $preRebrandRegistry32 = $preRebrandRegistry32 | Select-Object -Last 1
        $preRebrandRegistry32 = $preRebrandRegistry32.UninstallString -Replace "msiexec.exe","" -Replace "/I","" -Replace "/X",""
        $preRebrandRegistry32 = $preRebrandRegistry32.Trim()
        Start-Process "msiexec.exe" -arg "/X $preRebrandRegistry32 /passive" -Wait
    }
    
}


function Stop-PMBridge {
    # Stop the `bridge` process to completely quit Bridge

    $bridge = Get-Process "bridge" -ErrorAction SilentlyContinue

    if ($bridge){

        $bridge | Stop-Process -Force

    }

}


function Remove-PMBridgeResources {
    # Delete all the Bridge resource folders
    # They should be deleted with uninstalling Bridge
    # But to just make sure do this again

    Remove-Item $RoamProtonmail -Force -Recurse -ErrorAction SilentlyContinue
    Remove-Item $RoamProtonAG -Force -Recurse -ErrorAction SilentlyContinue
    Remove-Item $LocalProtonmail -Force -Recurse -ErrorAction SilentlyContinue
    Remove-Item $LocalProtonAG -Force -Recurse -ErrorAction SilentlyContinue
    Remove-Item $StartUpProtonBridge -Force -Recurse -ErrorAction SilentlyContinue

}


function Find-PMBridgeResources {
    # Search and check if the Bridge resource folders
    # Are deleted
    # Write to Output the result

    $FolderExists = $false

    if ( Test-Path -Path $RoamProtonmail ){
        Write-Host "`r`n'$RoamProtonmail' is not deleted!" -ForegroundColor Red
        $FolderExists = $true
    }

    if ( Test-Path -Path $RoamProtonAG ) {
        Write-Host "`r`n'$RoamProtonAG' is not deleted!" -ForegroundColor Red
        $FolderExists = $true
    }

    if ( Test-Path -Path $LocalProtonmail ) {
        Write-Host "`r`n'$LocalProtonmail' is not deleted!" -ForegroundColor Red
        $FolderExists = $true
    }

    if ( Test-Path -Path $LocalProtonAG ) {
        Write-Host "`r`n'$LocalProtonAG' is not deleted!" -ForegroundColor Red
        $FolderExists = $true
    }

    if ( Test-Path -Path $StartUpProtonBridge ) {
        Write-Host "`r`n'$StartUpProtonBridge' is not deleted!" -ForegroundColor Red
        $FolderExists = $true
    }

    if ( $FolderExists ) {
        Write-Host "`r`nSome directories were not deleted properly!`r`n" -ForegroundColor Red
    }

    else {
        Write-Host "`r`nAll Bridge resource folders deleted!`r`n" -ForegroundColor Green
    }

}


function Remove-PMBridgeCredentials {
    # Delete the entries in the credential manager

    $CredentialsData = @((cmdkey /listall | Where-Object{$_ -like "*LegacyGeneric:target=protonmail*"}).replace("Target: ",""))

    for($i =0; $i -le ($CredentialsData.Count -1); $i++){
        [string]$DeleteData = $CredentialsData[$i].trim()
        cmdkey /delete:$DeleteData
    }

}


function Invoke-BridgeFunctions {
    Stop-PMBridge
    Uninstall-PMBridge
    Remove-PMBridgeResources
    Find-PMBridgeResources
    Remove-PMBridgeCredentials
}


Invoke-BridgeFunctions
