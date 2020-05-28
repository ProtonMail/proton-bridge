#!/bin/bash

# Copyright (c) 2020 Proton Technologies AG
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

# create bitmaps
for shape in rounded rectangle
do
    for usage in systray syswarn app
    do
        group=$shape-$usage
        inkscape --without-gui  --export-id=$group --export-png=$group.png all_icons.svg
    done
done

# mac icon
png2icns Bridge.icns rounded-app.png

# windows icon
convert rectangle-app.png -define icon:auto-resize=256,128,64,48,32,16 logo.ico
