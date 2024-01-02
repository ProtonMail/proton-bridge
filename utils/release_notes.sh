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


# Generate HTML release notes
# hosted at https://proton.me/download/{ie,bridge}/{stable,early}_releases.html
INFILE=$1
OUTFILE=${INFILE//.md/.html}

CHANNEL=beta
if [[ "$INFILE" =~ stable ]]; then
    CHANNEL=stable
fi

# Check dependencies
if ! which pandoc; then
  printf "PANDOC NOT FOUND!\nPlease install pandoc in order to build release notes.\n"
  exit 1
fi

# Check Pandoc version
PANDOC_VERSION=$(pandoc --version | grep --color=never -m 1 "pandoc"  | sed -nre 's/^[^0-9]*(([0-9]+\.)*[0-9]+).*/\1/p')
printf "PANDOC FOUND ! version : %s\n", "$PANDOC_VERSION"

# self-contained is deprecated since 2.19 in profit of --embed-resource option
DEPRECATING_VERSION="2.19.0"
# Build release notes
function ver { printf "%03d%03d%03d%03d" $(echo "$1" | tr '.' ' '); }

if [ $(ver $PANDOC_VERSION) -lt $(ver $DEPRECATING_VERSION) ]; then
    pandoc "$INFILE" -f markdown -t html -s -o "$OUTFILE" -c utils/release_notes.css --self-contained --section-divs --metadata title="Release notes - Proton Mail Bridge - $CHANNEL"
else
    pandoc "$INFILE" -f markdown -t html -s -o "$OUTFILE" -c utils/release_notes.css --embed-resource --standalone --section-divs --metadata title="Release notes - Proton Mail Bridge - $CHANNEL"
fi
