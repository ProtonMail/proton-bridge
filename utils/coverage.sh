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


# This script calculates coverage for bridge project
# 
# Output:
#   stdout        : total coverage (to be parsed by Gitlab pipeline)
#   coverage.xml  : Cobertura format of covered lines for coverage visualization in Gitlab

# Assuming that test coverages from all jobs were put into `./coverage` folder
# and passed as artifacts. The flags are:
# -covermode=count
# -coverpkg=github.com/ProtonMail/proton-bridge/v3/internal/...,github.com/ProtonMail/proton-bridge/v3/pkg/...,
# -args -test.gocoverdir=$$PWD/coverage/${TOPIC}


ALLINPUTS="coverage$(printf ",%s" coverage/*)"

go tool covdata textfmt \
    -i "$ALLINPUTS" \
    -o coverage_withGen.out

# Filter out auto-generated code
grep -v '\.pb\.go' coverage_withGen.out > coverage.out

# Print out coverage
go tool cover -func=./coverage.out | grep total:

# Convert to Cobertura
#
# NOTE: We are not using the latest `github.com/boumenot/gocover-cobertura`
# because it does not support multiplatform coverage in one profile. See
# https://github.com/boumenot/gocover-cobertura/pull/3#issuecomment-1571586099
go get github.com/t-yuki/gocover-cobertura
go run github.com/t-yuki/gocover-cobertura < ./coverage.out > coverage.xml
