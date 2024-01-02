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

if [ $# -ne 1 ]; then
    echo "First argument must be path to file"
    exit 2
fi

CHANGELOG_FILE=$1

ERROR_COUNT_FILE="`mktemp`"
echo "0">$ERROR_COUNT_FILE

##########################
# -- Helper functions -- #
##########################

# err print out a given error ($2) and line where it happens ($1),
# also it increases count of errors.
err () {
    echo "CHANGELOG-LINTER: $2 on the following line:"
    echo "$1"

    echo $((`cat $ERROR_COUNT_FILE` + 1)) > $ERROR_COUNT_FILE
}

# trim removes extra whitespaces.
trim () {
    echo "$1"|sed 's/[ \t]*$//g;s/^[ \t]//g'
}

# paragraph_continues checks that given line ends with sentence ending symbol like [.!?:].
is_ended() {
    ENDING="${1: -1}"
    [ "$ENDING" == "." ] || [ "$ENDING" == "!" ] || [ "$ENDING" == "?" ] || [ "$ENDING" == ":" ]
}

# paragraph_continues checks that given line starts with a small letter.
# (so it's not a start of a new sentence or block.)
paragraph_continues() {
    LINE_START="${1:0:1}"
    echo "$LINE_START" | grep -q "[a-z]"
}

##########################
# -- Linter functions -- #
##########################

# check_lists checks a format of lists.
# - Starting with " * " (minus lists is not allowed).
# - Containing whole sentence with first capital letter and dot at the end.
check_lists () {
    if [ "${1:0:2}" == "- " ]; then
        err "$1" '"-" lists is not allowed, use "*" instead'
    fi

    if [ "${1:0:2}" != "* " ]; then # It's not a list. Skipping...
        return
    fi

    if ! is_ended "$1" && ! paragraph_continues "$2"; then
        err "$1" "List should contain a full sentence ending with one of the [!?.:] symbols"
    fi

    if ! (echo "${1:0:3}" | grep -q "[A-Z\#\`\"]"); then
        err "$1" "List should contain a full sentence starting with a capital letter"
    fi
}

# check_change_types checks a format change type headings.
# - See https://keepachangelog.com/en/1.0.0/ for allowed formats.
check_change_types () {
    if [ "${1:0:4}" != "### " ]; then # It's not a type heading. Skipping...
        return
    fi

    case "$1" in
        "### Added"|"### Changed"|"### Deprecated"|"### Removed"|"### Fixed"|"### Security") ;; # Standard keepachangelog.com compliant types.
        "### Release notes"|"### Fixed bugs") ;;                                                # Bridge additional in app release notes types.
        "### Guiding Principles"|"### Types of changes") ;;                                     # Ignoring guide at the end of the changelog.
        *) err "$1" "Change type must be one of the Added, Changed, Deprecated, Removed, Fixed, Hoftix"
    esac 
}


# check_application_name checks if a application name is defined on each record.
check_application_name () {
    if [ "${1:0:4}" != "## [" ]; then # It's not a version heading. Skipping...
        return
    fi

    case "`echo $1|cut -d"[" -f2|cut -d" " -f1`" in
        "IE"|"Bridge") ;;
        *) err "$1" "Either \"IE\" or \"Bridge\" application name should be inside the version header"
    esac 
}

# ignored_line determines lines which will not be linted.
ignored_line () {
    [ "$1" == "" ] # Ignoring empty lines.
}


######################
# -- Main routine -- #
######################

LINE_BEFORE="" # Storing the line before for some multiline operations.
cat $CHANGELOG_FILE|while read L; do
    LINE=`trim "$L"`

    if ignored_line "$LINE"; then
        continue 
    fi


    check_lists "$LINE_BEFORE" "$LINE"
    check_change_types "$LINE"
    check_application_name "$LINE"


    LINE_BEFORE="$LINE"
done

ERROR_COUNT=`cat $ERROR_COUNT_FILE`
rm $ERROR_COUNT_FILE

echo "CHANGELOG-LINTER found $ERROR_COUNT problems."|sed "s/found 0 problems/passed successfully ;)/"
exit $ERROR_COUNT
