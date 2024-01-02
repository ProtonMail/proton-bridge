#!/bin/bash

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

case $1 in
    check|add|change-year)
        ;;
    *)
        echo "Please select one of the actions:"
        echo "    check"
        echo "    add"
        echo "    change-year"
        exit 1
esac
[[ "$1" == "check" ]] && [[ -n "${MISSING_FILES}" ]] && exit 1

YEAR=`date +%Y`
GREP_COPYRIGHT="Copyright (c) ${YEAR} Proton AG"

MISSING_FILES=$(find . \
    -not -path "./extern/*" \
    -not -path "./*deploy/*" \
    -not -path "./vendor/*" \
    -not -path "./vendor-cache/*" \
    -not -path "./.cache/*" \
    -not -name "*mock*.go" \
    -not -path "*/cmake-build-*/*" \
    -not -name "*.pb.go" \
    -not -name "*.pb.h" \
    -not -name "*.pb.cc" \
    -not -name "*_moc.h" \
    -regextype posix-egrep -regex ".*\.go|.*\.qml|.*\.sh|.*\.py|.*\.cpp|.*\.cc|.*\.h|.*\.hpp|.*\.m|.*\.mm|.*\.h\.in" \
    -exec grep -L "$GREP_COPYRIGHT" {} \;)

MANUAL_CHECK=$(grep -L "$GREP_COPYRIGHT" \
    ./utils/license_header.txt \
    README.md)

if [ -n "${MANUAL_CHECK}" ]; then
    MISSING_FILES="$MISSING_FILES $MANUAL_CHECK"
fi;

for f in ${MISSING_FILES}
do
    echo -n "MISSING LICENSE or WRONG YEAR in $f"
    if [[ $1 == "add" ]]
    then
        cat ./utils/license_header.txt "$f" > tmp
        mv tmp "$f"
        echo -n "... license added"
    fi
    if [[ $1 == "change-year" ]]
    then
        sed -i "s/Copyright (c) [0-9]\\{4\\} Proton AG/Copyright (c) ${YEAR} Proton AG/" "$f" || exit 3
        echo -n "... replaced copyright year"
    fi
    echo
done

[[ "$1" == "check" ]] && [[ -n "${MISSING_FILES}" ]] && exit 1
exit 0
