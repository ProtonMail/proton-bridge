// Copyright (c) 2021 Proton Technologies AG
//
// This file is part of ProtonMail Bridge.
//
// ProtonMail Bridge is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// ProtonMail Bridge is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with ProtonMail Bridge.  If not, see <https://www.gnu.org/licenses/>.

#!/bin/bash

# Generate HTML release notes
# hosted at https://protonmail.com/download/{ie,bridge}/release_notes.html

# Load props
APP_NAME=$1
if [ "$APP_NAME" == "" ]; then 
    APP_NAME="Bridge"
fi

APP_TYPE=$(echo "$APP_NAME"|tr [A-Z] [a-z])
INFILE="release-notes/${APP_TYPE}.md"
OUTFILE="release-notes/${APP_TYPE}.html"

# Check dependencies
if ! which -s pandoc; then 
  echo "PANDOC NOT FOUND!\nPlease install pandoc in order to build release notes."
  exit 1
fi

# Build release notes
pandoc $INFILE -f markdown -t html -s -o $OUTFILE -c utils/release_notes.css --self-contained --section-divs --metadata title="Release notes - ProtonMail $APP_NAME"
