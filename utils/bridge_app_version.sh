#!/bin/sh
sed -n "s/BRIDGE_APP_VERSION?=\(\S*\)/\1/p" "$(dirname $0)/../Makefile"
