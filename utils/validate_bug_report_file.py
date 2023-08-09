#!/usr/bin/env python3
# -*- coding: utf-8 -*-

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
# along with Proton Mail Bridge. If not, see <https://www.gnu.org/licenses/>.

import argparse
import json
import re


class BugReportJson:

    def __init__(self, filepath):
        self.filepath = filepath
        self.json = None
        self.metadata = None
        self.version = None
        self.data = None
        self.categories = None
        self.questions = None
        self.questionsID = []
        self.error = ""

    def validate(self):
        with open(self.filepath) as infile:
            self.json = json.load(infile)
        if self.json is None:
            return False, ("JSON cannot be load from %s." % self.filepath)

        for object in self.json:
            if not (object == "metadata" or re.match(r"data_v[0-9]+\.[0-9]+\.[0-9]+", object)) :
                self.error = ("Unexpected object name %s." % object)
                return False

        if not self.parse_metadata():
            return False
        if not self.parse_data():
            return False
        if not self.parse_questions():
            return False
        if not self.parse_categories():
            return False
        return True

    def parse_metadata(self):
        if "metadata" not in self.json:
            self.error = "No metadata object."
            return False
        if not isinstance(self.json["metadata"], dict):
            self.error = "metadata should be a dictionary."
            return False

        self.metadata = self.json["metadata"]
        if "version" not in self.metadata:
            self.error = "No version in metadata object."
            return False

        self.version = self.metadata["version"]
        if not re.match(r"[0-9]+\.[0-9]+\.[0-9]+", self.version):
            self.error = ("Version (%s) doesn't match pattern." % self.version)
            return False
        return True

    def parse_data(self):
        data_version = ("data_v%s" % self.version)
        if data_version not in self.json:
            self.error = ("No data object matching version %s." % self.version)
            return False

        if not isinstance(self.json[data_version], dict):
            self.error = ("%s should be a dictionary." %data_version)
            return False

        self.data = self.json[data_version]

        if "categories" not in self.data:
            self.error = "No categories object in data."
            return False
        self.categories = self.data["categories"]
        if not isinstance(self.categories, list):
            self.error = "categories should be an array."
            return False

        if "questions" not in self.data:
            self.error = "No questions object in data."
            return False
        self.questions = self.data["questions"]
        if not isinstance(self.questions, list):
            self.error = "questions should be an array."
            return False
        return True

    def parse_questions(self):
        for question in self.questions:
            if not isinstance(question, dict):
                self.error = ("Question should be a dictionary.")
                return False
            for option in question:
                if option not in ["id", "text", "tips", "type", "mandatory", "maxChar", "answerList"]:
                    self.error = ("Unexpected option '%s' in question." % option)
                    return False
            # check mandatory field
            if "id" not in question:
                self.error = ("Missing id in question %s." % question)
                return False
            if question["id"] in self.questionsID:
                self.error = ("Question id should be unique (%d)." % question["id"])
                return False
            self.questionsID.append(question["id"])

            if "text" not in question:
                self.error = ("Missing text in question %s." % question)
                return False

            if "type" not in question:
                self.error = ("Missing type in question %s." % question)
                return False

            # check type restriction
            if question["type"] == "open":
                if "maxChar" in question:
                    if question["maxChar"] > 1000:
                        self.error = ("MaxChar is too damn high in question %s." % question)
                        return False
                    if "answerList" in question:
                        self.error = ("AnswerList should not be present in open question %s." % question)
                        return False
            elif question["type"] == "choice" or question["type"] == "multichoice":
                if "answerList" not in question:
                    self.error = ("Missing answerList in question %s." % question)
                    return False
                if not isinstance(question["answerList"], list):
                    self.error = ("AnswerList should be an array in question %s." % question)
                    return False
                if "maxChar" in question:
                    self.error = ("maxChar should not be present in choice/multichoice question %s." % question)
                    return False
            else:
                self.error = ("Wrong type in question %s." % question)
                return False
        return True

    def parse_categories(self):
        for category in self.categories:
            if not isinstance(category, dict):
                self.error = ("category should be a dictionary.")
                return False
            for option in category:
                if option not in ["name", "questions", "hint"]:
                    self.error = ("Unexpected option '%s' in category." % option)
                    return False
            if "name" not in category:
                self.error = ("Missing name in category %s." % category)
                return False
            if "questions" not in category:
                self.error = ("Missing questions in category %s." % category)
                return False
            unique_list = []
            for question in category["questions"]:
                if question not in self.questionsID:
                    self.error = ("Questions referring to non-existing question in category %s." % category)
                    return False
                if question in unique_list:
                    self.error = ("Questions contains duplicate in category %s." % category)
                    return False
                unique_list.append(question)
        return True

    def preview(self):
        for category in self.categories:
            self.preview_category(category)

    def preview_category(self, category):
        print(" > %s" % category["name"])
        if "hint" in category and category["hint"]:
            print("(%s)" % category["hint"])
        for question in category["questions"]:
            self.preview_question(self.questions[question])
        print("\n\r")
        return 0

    def preview_question(self, question):
        # ["id", "text", "tips", "type", "mandatory", "maxChar", "answerList"]
        mandatory = ("mandatory" in question and question["mandatory"])
        mandatory_sym = " *"
        if not mandatory:
            mandatory_sym = ""
        print("\t - %s%s" % (question["text"], mandatory_sym))
        if "tips" in question and question["tips"]:
            print("\t (%s)" % question["tips"])
        if "answerList" in question:
            for answer in question["answerList"]:
                print("\t\t - %s" % answer)
        return 0

def parse_args():
    parser = argparse.ArgumentParser(description='Validate Bug Report File.')
    parser.add_argument('--file', required=True, help='JSON file to validate.')
    parser.add_argument('--preview', action='store_true', help='Output a preview of the parsed file.')
    return parser.parse_args()


def main():
    args = parse_args()
    report = BugReportJson(args.file)

    if not report.validate():
        print("Validation FAILED for %s. Error: %s" %(report.filepath, report.error))
        exit(1)
    print("Validation SUCCEED for %s." % report.filepath)
    if args.preview:
        report.preview()
    exit(0)


if __name__ == "__main__":
    main()