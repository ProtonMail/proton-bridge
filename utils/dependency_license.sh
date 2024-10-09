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

set -eo pipefail

src=go.mod
tgt=COPYING_NOTES.md

STARTAUTOGEN="<!-- START AUTOGEN -->"
ENDAUTOGEN="<!-- END AUTOGEN -->"
RE_STARTAUTOGEN="^${STARTAUTOGEN}$"
RE_ENDAUTOGEN="^${ENDAUTOGEN}$"
tmpDepLicenses=""

error(){
    echo "Error: $*"
    exit 1
}

generate_dep_licenses(){
    [ -r $src ] || error "Cannot read file '$src'"


    tmpDepLicenses="$(mktemp)"

    # Collect all go.mod lines beginig with tab:
    # * which no replace
    # * which have replace
    grep -E $'^\t[^=>]*$'    $src  | sed -r 's/\t([^ ]*) v.*/\1/g'         > "$tmpDepLicenses"
    grep -E $'^\t.*=>.*v.*$' $src  | sed -r 's/^.*=> ([^ ]*)( v.*)?/\1/g' >> "$tmpDepLicenses"

    # Replace each line with formatted link
    sed  -i -r '/^github.com\/therecipe\/qt\/internal\/binding\/files\/docs\//d;' "$tmpDepLicenses"
    sed -i -r 's|^(.*)/([[:alnum:]-]+)/(v[[:digit:]]+)$|* [\2](https://\1/\2/\3)|g' "$tmpDepLicenses"
    sed -i -r 's|^(.*)/([[:alnum:]-]+)$|* [\2](https://\1/\2)|g' "$tmpDepLicenses"
    sed -i -r 's|^(.*)/([[:alnum:]-]+).(v[[:digit:]]+)$|* [\2](https://\1/\2.\3)|g' "$tmpDepLicenses"


    ## add license file to github links, and others
    sed -i -r '/github.com/s|^(.*(https://[^)]+).*)$|\1 available under [license](\2/blob/master/LICENSE) |g' "$tmpDepLicenses"
    sed -i -r '/gitlab.com/s|^(.*(https://[^)]+).*)$|\1 available under [license](\2/blob/master/LICENSE) |g' "$tmpDepLicenses"
    sed -i -r '/go.etcd.io\/bbolt/s|^(.*)$|\1 available under [license](https://github.com/etcd-io/bbolt/blob/master/LICENSE) |g' "$tmpDepLicenses"
    sed -i -r '/howett.net\/plist/s|^(.*)$|\1 available under [license](https://github.com/DHowett/go-plist/blob/main/LICENSE) |g' "$tmpDepLicenses"
    sed -i -r '/golang.org\/x/s|^(.*golang.org/x/([^)]+).*)$|\1 available under [license](https://cs.opensource.google/go/x/\2/+/master:LICENSE) |g' "$tmpDepLicenses"
    sed -i -r '/google.golang.org\/grpc/s|^(.*)$|\1 available under [license](https://github.com/grpc/grpc-go/blob/master/LICENSE) |g' "$tmpDepLicenses"
    sed -i -r '/google.golang.org\/protobuf/s|^(.*)$|\1 available under [license](https://github.com/protocolbuffers/protobuf/blob/main/LICENSE) |g' "$tmpDepLicenses"
    sed -i -r '/go.uber.org\/goleak/s|^(.*)$|\1 available under [license](https://pkg.go.dev/go.uber.org/goleak?tab=licenses) |g' "$tmpDepLicenses"
    sed -i -r '/ariga.io\/atlas/s|^(.*)$|\1 available under [license](https://github.com/ariga/atlas/blob/master/LICENSE) |g' "$tmpDepLicenses"
    sed -i -r '/entgo.io\/ent/s|^(.*)$|\1 available under [license](https://pkg.go.dev/entgo.io/ent?tab=licenses) |g' "$tmpDepLicenses"
    sed -i -r '/google.golang.org\/genproto/s|^(.*)$|\1 available under [license](https://pkg.go.dev/google.golang.org/genproto?tab=licenses) |g' "$tmpDepLicenses"
    sed -i -r '/gopkg.in\/yaml\.v3/s|^(.*)$|\1 available under [license](https://github.com/go-yaml/yaml/blob/v3.0.1/LICENSE) |g' "$tmpDepLicenses"
    sed -i -r '/google.golang.org\/appengine/s|^(.*)$|\1 available under [license](https://pkg.go.dev/google.golang.org/appengine?tab=licenses) |g' "$tmpDepLicenses"
    sed -i -r '/go.opencensus.io/s|^(.*)$|* [go.opencensus.io](https://pkg.go.dev/go.opencensus.io?tab=licenses) available under [license](https://pkg.go.dev/go.opencensus.io?tab=licenses) |g' "$tmpDepLicenses"
    sed -i -r '/google.golang.org\/api/s|^(.*)$|\1 available under [license](https://pkg.go.dev/google.golang.org/api?tab=licenses) |g' "$tmpDepLicenses"
    sed -i -r '/cloud.google.com\/go\/compute\/metadata/s|^(.*)$|\1 available under [license](https://pkg.go.dev/cloud.google.com/go/compute/metadata?tab=licenses) |g' "$tmpDepLicenses"
    sed -i -r '/cloud.google.com\/go\/compute\/metadata/!{/cloud.google.com\/go\/compute/s|^(.*)$|\1 available under [license](https://pkg.go.dev/cloud.google.com/go/compute?tab=licenses) |g}' "$tmpDepLicenses"
    sed -i -r '/gopkg.in\/yaml\.v3/s|^(.*)$|\1 available under [license](https://github.com/go-yaml/yaml/blob/v3.0.1/LICENSE) |g' "$tmpDepLicenses"
}


check_dependecies(){
    generate_dep_licenses

    tmpHaveLicenses=$(mktemp)
    sed "/${RE_STARTAUTOGEN}/,/${RE_ENDAUTOGEN}/!d;//d" $tgt > "$tmpHaveLicenses"

    #echo "have"
    #cat "$tmpHaveLicenses"

    #echo "want"
    #cat "$tmpDepLicenses"

    diffOK=0
    if ! diff "$tmpHaveLicenses" "$tmpDepLicenses"; then diffOK=1; fi

    rm "$tmpDepLicenses" || echo "Failed to clean tmp file"
    rm "$tmpHaveLicenses" || echo "Failed to clean tmp file"

    [ $diffOK -eq 0 ] || error "Dependency licenses are not up-to-date"
    exit 0
}

update_dependecies(){
    generate_dep_licenses

    sed -i -e "/${RE_STARTAUTOGEN}/,/${RE_ENDAUTOGEN}/!b" \
        -e "/${RE_ENDAUTOGEN}/i ${STARTAUTOGEN}" \
        -e "/${RE_ENDAUTOGEN}/r $tmpDepLicenses" \
        -e "/${RE_ENDAUTOGEN}/a ${ENDAUTOGEN}" \
        -e "d" \
        $tgt


    rm "$tmpDepLicenses" || echo "Failed to clean tmp file"

    exit 0
}

case $1 in
    "check") check_dependecies;;
    "update") update_dependecies;;
    *) error "One of actions needed: check update" ;;
esac

