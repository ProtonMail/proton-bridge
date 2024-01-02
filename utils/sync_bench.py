#!/usr/bin/env python
# -*- coding: utf-8 -*-

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


# first try
data = {

        "mails"         : [ 500        , 1000       , 2000       , 10*1000     ] ,
        "retries"       : [ 1          , 1          , 1          , 1           ] ,
        "time[ns]"      : [ 2386939573 , 5082383333 , 9836788842 , 46028190203 ] ,
        "mem[MB]"       : [ 33.28      , 56.44      , 138.73     , 732.02      ] ,
        "time[s]"       : [ 0          , 0          , 0          , 0           ] ,
        "time/mail[ms]" : [ 0          , 0          , 0          , 0           ] ,
        "mem/mail[kB]"  : [ 0          , 0          , 0          , 0           ] ,

  }


# using Put in loop
data = {

        "mails"         : [ 500        , 1000       , 2000       , 10*1000     ] ,
        "retries"       : [ 1          , 1          , 1          , 1           ] ,
        "time[ns]"      : [ 2289519548 , 4469488350 , 8935239780 , 50641402322 ] ,
        "mem[MB]"       : [ 27.21      , 60.51      , 127.06     , 763.31      ] ,
        "time[s]"       : [ 0          , 0          , 0          , 0           ] ,
        "time/mail[ms]" : [ 0          , 0          , 0          , 0           ] ,
        "mem/mail[kB]"  : [ 0          , 0          , 0          , 0           ] ,

  }

# after PutMany rewrite
data = {

        "mails"         : [ 500      , 1000     , 2000      , 10*1000    , 50*1000    , 100*1000    ] ,
        "retries"       : [ 20       , 20       , 10        , 1          , 1          , 1           ] ,
        "time[ns]"      : [ 53334780 , 70278788 , 122645482 , 1098762327 , 7565912638 , 28189738329 ] ,
        "mem[MB]"       : [ 61.82    , 162.69   , 191.60    , 36.53      , 258.58     , 523.37      ] ,

        "mem[MB/op]"    : [ 0        , 0        , 0         , 0          , 0          , 0           ] ,
        "time[s]"       : [ 0        , 0        , 0         , 0          , 0          , 0           ] ,
        "time/mail[ms]" : [ 0        , 0        , 0         , 0          , 0          , 0           ] ,
        "mem/mail[kB]"  : [ 0        , 0        , 0         , 0          , 0          , 0           ] ,

  }

# after refactor
data = {

        "mails"    : [ 500      , 1000     , 2000      , 10*1000    , 50*1000     , 100*1000    ] ,
        "retries"  : [ 20       , 20       , 10        , 1          , 1           , 1           ] ,
        "time[ns]" : [ 65210959 , 84387204 , 154611276 , 1450409808 , 12866781601 , 53265480248 ] ,
        "mem[MB]"  : [ 0        , 0        , 0         , 0          , 60.39       , 523.37      ] ,



        "mem[MB/op]"    : [ 0        , 0        , 0         , 0          , 0          , 0           ] ,
        "time[s]"       : [ 0        , 0        , 0         , 0          , 0          , 0           ] ,
        "time/mail[ms]" : [ 0        , 0        , 0         , 0          , 0          , 0           ] ,
        "mem/mail[kB]"  : [ 0        , 0        , 0         , 0          , 0          , 0           ] ,

  }

allColumns= [ "mails", "retries" , "time[ns]" , "time[s]" , "time/mail[ms]" , "mem[MB]", "mem[MB/op]"  , "mem/mail[kB]" ]

for i,n in enumerate(data["mails"]):
    data["time[s]"][i]       = data["time[ns]"][i] /1000/1000/1000
    data["time/mail[ms]"][i] = data["time[ns]"][i] /1000/1000 /n
    data["mem[MB/op]"][i]    = data["mem[MB]"][i] /data["retries"][i]
    data["mem/mail[kB]"][i]  = data["mem[MB/op]"][i] *1024 / n
    pass

print("| "+" | ".join(allColumns) + " |")
print("| :-- "*len(allColumns) + "|")
for i,n in enumerate(data["mails"]) :
    val = [ str(data[col][i]) for col in allColumns ]
    print("| "+" | ".join(val) + " |")
