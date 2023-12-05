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
# along with Proton Mail Bridge.  If not, see <https://www.gnu.org/licenses/>.


set -eo pipefail

main(){
    go install golang.org/x/vuln/cmd/govulncheck@latest
    make gofiles
    govulncheck -json ./... > vulns.json

    jq -r '.osv.id | select( . != null )' < vulns.json > vulns_osv_ids.txt

    ignore GO-2023-2102 "GODT-3160 update go to 1.21.4"
    ignore GO-2023-2043 "GODT-3160 update go to 1.21.4"
    ignore GO-2023-2041 "GODT-3160 update go to 1.21.4"
    ignore GO-2023-1878 "GODT-3160 update go to 1.21.4"
    ignore GO-2023-1987 "GODT-3160 update go to 1.21.4"
    ignore GO-2023-1840 "GODT-3160 update go to 1.21.4"
    ignore GO-2023-2185 "GODT-3160 update go to 1.21.4"
    ignore GO-2023-2186 "GODT-3160 update go to 1.21.4"
    ignore GO-2023-2328 "GODT-3124 RESTY race condition"

    has_vulns

    echo
    echo "No new vulnerabilities found."
}

ignore(){
    echo "ignoring $1 fix: $2"
    cp vulns_osv_ids.txt tmp
    grep -v "$1" < tmp > vulns_osv_ids.txt || true
    rm tmp
}

has_vulns(){
    has=false
    while read -r osv; do
        jq \
            --arg osvid "$osv" \
            '.osv | select ( .id == $osvid) | {"id":.id, "ranges": .affected[0].ranges,  "import": .affected[0].ecosystem_specific.imports[0].path}' \
            < vulns.json
        has=true
    done < vulns_osv_ids.txt

    if [ "$has" == true ]; then
        echo
        echo "Vulnerability found"
        return 1
    fi
}

main
