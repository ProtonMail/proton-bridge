#!/usr/bin/env python

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
# along with Proton Mail Bridge. If not, see <https://www.gnu.org/licenses/>.


import glob
import pandas as pd
import re


def print_header(report_file):
    print('\n======== %s ========' %
          (report_file.replace("./bench-", "").replace(".log", "")))


rx_line = {
    'exists': re.compile(r'.*Res[A-Za-z]?: [*] (?P<mails>\d+) EXISTS.*\n'),
    'bench': re.compile(r'Benchmark(?P<name>[^ \t]+)[ \t]+(?P<rpts>\d+)[ \t]+(?P<ns>\d+) ns/op.*\n'),
    # 'total' : re.compile(r'ok[ \t]+(?P<pkg>[^ \t]+)[ \t]+(?P<time>[^ \t\n]+)[ \t]*\n'),
}


def parse_line(line):
    for key, rx in rx_line.items():
        match = rx.search(line)
        if match:
            return key, match
    # if there are no matches
    return None, None


rx_count = re.compile(r'Fetch/1:(?P<count>\d+)-')


def parse_file(filepath):
    data = []  # create an empty list to collect the data
    # open the file and read through it line by line
    with open(filepath, 'r') as file_object:
        line = file_object.readline()
        last_count = 0
        mails = 1
        while line:
            # at each line check for a match with a regex
            key, match = parse_line(line)
            # print(line, key, match)
            if key != None:
                row = match.groupdict()
                if key == 'exists':
                    mails = int(row['mails'])
                    last_count = 0
                if key == 'bench':
                    match = rx_count.search(row['name'])
                    row['mails'] = mails - last_count
                    if match:
                        count = int(match.group('count'))
                        if count < mails:
                            row['mails'] = count - last_count
                        last_count = count
                    row['rpts'] = int(row['rpts'])
                    row['ns'] = int(row['ns'])
                    row['time/op'] = human_duration(row['ns'])
                    if row['mails'] > 0:
                        row['time/mails'] = human_duration(
                            row['ns']/row['mails']
                        )
                    data.append(row)
                if key == 'total':
                    row['name'] = key
                    data.append(row)
            line = file_object.readline()

    return data


def human_duration(duration_ns):
    unit = 'ns'
    factor = 1.
    unit_factors = [
        ('us', 1.e3),
        ('ms', 1.e3),
        ('s ', 1.e3),
        ('m ', 60.),
        ('h ', 60.),
        ('d ', 24.),
        ('w ', 7.),
        ('m ', 30./7.),
        ('y ', 12.),
    ]
    for unit_factor in unit_factors:
        if (abs(duration_ns) / factor / unit_factor[1]) < 1.0:
            break
        unit = unit_factor[0]
        factor *= unit_factor[1]
    return "%4.2f%s" % (duration_ns/factor, unit)


def print_table(data):
    data = pd.DataFrame(data)
    data.set_index('name', inplace=True)
    print(data)


if __name__ == "__main__":
    # for d in [ 0.5, 1, 2, 5, 1e3, 5e3, 1e4, 1e5, 1e6, 1e9, 2e9, 1e10, 1e11, 1e12, ]: print(human_duration(int(d)))
    for report_file in glob.glob("./*.log"):
        print_header(report_file)
        print_table(parse_file(report_file))
