#!/bin/bash

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
