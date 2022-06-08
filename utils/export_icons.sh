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
# along with Proton Mail Bridge.  If not, see <https://www.gnu.org/licenses/>.



RAW_PATHS=./dist/raw

create_bitmaps(){
    rm -f ${RAW_PATHS}/*_icon_*.png

    export_png ${RAW_PATHS}/mac_icon_512x512.svg 384 1024
    export_png ${RAW_PATHS}/mac_icon_512x512.svg 192 512
    export_png ${RAW_PATHS}/mac_icon_512x512.svg 96  256
    export_png ${RAW_PATHS}/mac_icon_256x256.svg 96  128
    export_png ${RAW_PATHS}/mac_icon_32x32.svg   192 32
    export_png ${RAW_PATHS}/mac_icon_32x32.svg   96  16

    export_png ${RAW_PATHS}/win+lin_icon_256x256.svg 192 256
}


# Inkscape (more precisely cairo) doesn't support customization of rendering
# and direct output is too sharp. Therefore, we double DPI for inkscape export
# and then scale down to correct dimension.
export_png(){
    inSVG=$1
    dpi=$2
    size=$3

    dimensions=${size}x${size}
    outPNG=$(echo "$inSVG" | sed 's/\(.*_icon\)_.*/\1/')_${dimensions}.png

    echo "$inSVG -> $outPNG $dpi $dimensions"
    inkscape "$inSVG" --export-filename=tmp.png --export-dpi "$dpi"
    file tmp.png
    convert tmp.png -resize "$dimensions" "$outPNG"
    file "$outPNG"
    rm tmp.png
}

create_mac_icon(){
    out=./dist/Bridge.icns
    rm -f ${out}
    png2icns ${out} \
        ${RAW_PATHS}/mac_icon_{1024x1024,512x512,256x256,128x128,32x32,16x16}.png
}

create_windows_icon(){
    out=./dist/bridge.ico
    rm -f ${out}
    convert \
        ${RAW_PATHS}/win+lin_icon_256x256.png \
        -define icon:auto-resize=256,128,64,48,32,16 ${out}
}

create_linux_icon(){
    out=./dist/bridge.svg
    rm -f ${out}
    cp ${RAW_PATHS}/win+lin_icon_256x256.svg ${out}
}

create_bitmaps
create_mac_icon
create_windows_icon
create_linux_icon
