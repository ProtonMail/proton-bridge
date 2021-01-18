#!/bin/bash

# Copyright (c) 2021 Proton Technologies AG
#
# This file is part of ProtonMail Bridge.
#
# ProtonMail Bridge is free software: you can redistribute it and/or modify
# it under the terms of the GNU General Public License as published by
# the Free Software Foundation, either version 3 of the License, or
# (at your option) any later version.
#
# ProtonMail Bridge is distributed in the hope that it will be useful,
# but WITHOUT ANY WARRANTY; without even the implied warranty of
# MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
# GNU General Public License for more details.
#
# You should have received a copy of the GNU General Public License
# along with ProtonMail Bridge.  If not, see <https://www.gnu.org/licenses/>.


# create QML JSON object from list of golang constants
# run this script and output line stored in `out.qml` insert to `Gui.qml`

list="
qtfrontend.PathOK
qtfrontend.PathEmptyPath
qtfrontend.PathWrongPath
qtfrontend.PathNotADir
qtfrontend.PathWrongPermissions
qtfrontend.PathDirEmpty

errors.ErrUnknownError
errors.ErrEventAPILogout
errors.ErrUpgradeAPI
errors.ErrUpgradeJSON
errors.ErrUserAuth
errors.ErrQApplication
errors.ErrEmailExportFailed
errors.ErrEmailExportMissing
errors.ErrNothingToImport
errors.ErrEmailImportFailed
errors.ErrDraftImportFailed
errors.ErrDraftLabelFailed
errors.ErrEncryptMessageAttachment
errors.ErrEncryptMessage
errors.ErrNoInternetWhileImport
errors.ErrUnlockUser
errors.ErrSourceMessageNotSelected

source.ErrCannotParseMail
source.ErrWrongLoginOrPassword
source.ErrWrongServerPathOrPort
source.ErrWrongAuthMethod
source.ErrIMAPFetchFailed

qtfrontend.ErrLocalSourceLoadFailed
qtfrontend.ErrPMLoadFailed
qtfrontend.ErrRemoteSourceLoadFailed
qtfrontend.ErrLoadAccountList
qtfrontend.ErrExit
qtfrontend.ErrRetry
qtfrontend.ErrAsk
qtfrontend.ErrImportFailed
qtfrontend.ErrCreateLabelFailed
qtfrontend.ErrCreateFolderFailed
qtfrontend.ErrUpdateLabelFailed
qtfrontend.ErrUpdateFolderFailed
qtfrontend.ErrFillFolderName
qtfrontend.ErrSelectFolderColor
qtfrontend.ErrNoInternet

qtfrontend.FolderTypeSystem
qtfrontend.FolderTypeLabel
qtfrontend.FolderTypeFolder
qtfrontend.FolderTypeExternal

backend.ProgressInit
backend.ProgressLooping
backend.ErrPMAPIMessageTooLarge

qtfrontend.StatusNoInternet
qtfrontend.StatusCheckingInternet
qtfrontend.StatusNewVersionAvailable
qtfrontend.StatusUpToDate
qtfrontend.StatusForceUpgrade
"

first=true


if true; then
    echo '// +build ignore'
    echo ''
    echo 'package main'
    echo ''
    echo 'import ('
    echo '    "github.com/ProtonMail/Import-Export/backend"'
    echo '    "github.com/ProtonMail/Import-Export/backend/source"'
    echo '    "github.com/ProtonMail/Import-Export/backend/errors"'
    echo '    "github.com/ProtonMail/Import-Export/frontend"'
    echo '    "fmt"'
    echo ')'
    echo ''
    echo 'func main(){'
    echo '    checkValues := map[int]string{}'
    echo '    checkDuplicates := map[string]bool{}'
    echo '    fmt.Print("{")'
    for c in $list
    do
        if ! $first; then
            echo 'fmt.Print(",")'
        fi

        if [[ $c =~ .*Err ]]; then
            ## Add check that all Err have different value
            echo 'if enumName,ok := checkValues[int('$c')]; ok {'
            echo '  panic("Enum '$c' and "+enumName+" has same value")'
            echo '}'
            echo 'checkValues[int('$c')]="'$c'"'
        fi

        cname=`echo $c | cut -d. -f2`
        lowCase=${cname,}

        ## Add check that all qml enums have different value
        echo 'if checkDuplicates["'$lowCase'"]{'
        echo '  panic("Enum with same lowcase name as '$c' has already been registered")'
        echo '}'
        echo 'checkDuplicates["'$lowCase'"]=true'

        ## add value in lowercase
        echo 'fmt.Printf("\"'$lowCase'\":%#v",'$c')'

        first=false
    done
    echo '    fmt.Print("}")'
    echo '}'
fi > main.go


if true; then
echo -n "property var enums : JSON.parse('"
go run main.go || exit 5
echo -n "')"
fi > out.qml

rm main.go
sed -i "s/property var enums : JSON.parse.*$/`cat out.qml`/" ./qml/Gui.qml
rm out.qml

