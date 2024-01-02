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
# along with Proton Mail Bridge. If not, see <https://www.gnu.org/licenses/>.

# The Qt libs are dynamically loaded with rules like: `@rpath/QtGui.framework/Versions/5/QtGui`
# @rpath instructs the dynamic linker to search a list of paths in order to locate the framework
# The rules can be listed using `otool -l "${path_to_binary}"`
# The building process of therecipe/qt or qmake leaves the rules with additional unwanted paths
#   + absolute path to build directory
#   + dummy replacement `/break_the_rpath`
# We need to manually remove those and add the path relative to executable: `@executable_path/../Frameworks`

path_to_binary=$1

if [ -z ${path_to_binary} ]; then
    echo "The first parameter should be path to binary"
    exit 2
fi

for path_to_remove in $(otool -l "${path_to_binary}" | egrep '/Users/|break_the_rpath' | awk '{print $2}');
do
    if [ ! -z "${path_to_remove}" ]; then
        printf "\e[0;32mRemove path to qt '${path_to_remove}' ...\033[0m\n\e[0;31m"
        install_name_tool -delete_rpath "${path_to_remove}" "${path_to_binary}" || exit 1
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

