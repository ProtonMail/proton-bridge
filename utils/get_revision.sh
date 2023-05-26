#!/usr/bin/env bash

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


if ! which git &> /dev/null; then
    echo "NOGIT"
    exit
fi

if ! git status &> /dev/null; then
    echo "NOREPO"
    exit
fi

if [ "$1" == "tag" ]; then
    if ! git describe --tags; then
        echo "NOTAG"
    fi
    exit
fi

if ! git rev-parse --short=10 HEAD; then
    echo "NOREV"
    exit
fi

