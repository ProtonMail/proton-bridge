##!/bin/bash

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
# along with Proton Mail Bridge.  If not, see <https://www.gnu.org/licenses/>.


REPLACE_FILES=$(find . -not -path "./vendor/*" \
    -not -path "./vendor-cache/*" \
    -not -path "./.cache/*" \
    -not -name "*mock*.go" \
    -regextype posix-egrep \
    -regex ".*\.go|.*\.qml|qmldir|.*\.qmlproject|.*\.txt|.*\.md|.*\.h|.*\.cpp|.*\.m|.*\.sh|.*\.py" \
    -exec grep -L "Copyright (c) 2024 Proton AG" {} \;)

for f in ${REPLACE_FILES}
do
    if [[ $f =~ rebranding.sh ]]; then continue; fi;
    echo "replacing $f"

    # This file is part of Proton Mail Bridge.
    # Proton Mail Bridge is free software: you can redistribute it and/or modify
    # Proton Mail Bridge is distributed in the hope that it will be useful,
    # along with Proton Mail Bridge.  If not, see <https://www.gnu.org/licenses/>.
    # Copyright (c) 2024 Proton AG
    for repl in  \
    "s/This file is part of ProtonMail Bridge./This file is part of Proton Mail Bridge./" \
    "s/ProtonMail Bridge is free software:/Proton Mail Bridge is free software:/" \
    "s/ProtonMail Bridge is distributed in the hope that it will be useful/Proton Mail Bridge is distributed in the hope that it will be useful/" \
    "s/along with ProtonMail Bridge. If not, see/along with Proton Mail Bridge. If not, see/" \
    "s/along with ProtonMail Bridge.  If not, see/along with Proton Mail Bridge. If not, see/" \
    "s/Copyright (c) 2022 Proton Technologies AG/Copyright (c) 2024 Proton AG/" \
    "s/Copyright (c) 2021 Proton Technologies AG/Copyright (c) 2024 Proton AG/" \
    "s/Copyright (c) 2020 Proton Technologies AG/Copyright (c) 2024 Proton AG/"
    do
        sed -i "$repl" "$f" || exit 3
    done
done


## Manual fixes
# ./CONTRIBUTING.md
# ./internal/frontend/qml/Proton/qmldir
# ./utils/githooks/pre-push
# ./BUILDS.md
# ./README.md
# ./Changelog.md
# ./doc/bridge.md

## Manual greps
# ack  ProtonMail | ack -v rebranding |  ack -v test | ack -v github.com/ProtonMail | ack ProtonMail
# ack Technologies

## GUI: use backend or global var
# internal/frontend/qml/Notifications/Notifications.qml
# internal/frontend/qml/StatusWindow.qml
# internal/frontend/qml/MainWindow.qml
# internal/frontend/qml/WelcomeGuide.qml

## CLI:
# internal/frontend/cli/utils.go
# internal/frontend/cli/frontend.go

## Config (no migration needed)
# internal/imap/server.go -- IMAP ID
# internal/smtp/user.go -- comment
# internal/config/tls/tls.go -- newly created cert only
# cmd/launcher/main.go (sentry)

## App rename and installation paths (might have huge impact on functionality)
# cmd/Desktop-Bridge/main.go -- base
#   - sentry
#   - notifications
#   - autostart
#   - cli.App name
#   - frontend programName
# internal/frontend/share/info.rc
#   - Test it works with icon (is used in release?)
#   - needs re-install
# dist/proton-bridge.desktop
#   - check it works with reinstall, needs reinstall
# internal/locations/locations.go
#   - check it works with reinstall, needs reinstall

## Keep unchanged
# internal/constants/constants.go:23 for locations provider

