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


## Make sure that mac exe will not contain broken library links
# * remove absolute paths for Qt libs
# * add relative part to app bundle Frameworks

path_to_binary=$1

if [ -z ${path_to_binary} ]; then
    echo "The first parameter should be path to binary"
    exit 2
fi

for remove_path_qt in $(otool -l "${path_to_binary}" | grep '/Users/' | awk '{print $2}');
do
    if [ ! -z "${remove_path_qt}" ]; then
        printf "\e[0;32mRemove path to qt ${remove_path_qt} ...\033[0m\n\e[0;31m"
        install_name_tool -delete_rpath "${remove_path_qt}" "${path_to_binary}" || exit 1
    fi
done
rpath_rule=$(otool -l "${path_to_binary}" | grep executable_path | awk '{print $2}')
printf "\e[0;32mRPATH rules '${rpath_rule}' \033[0m\n\e[0;31m"
if [ -z ${rpath_rule} ];
then
    echo "There should be at least executable path..."
    printf "\e[0;32mAdding path '@executable_path/../Frameworks'\033[0m\n\e[0;31m"
    install_name_tool -add_rpath "@executable_path/../Frameworks" "${path_to_binary}" || exit 1
fi

