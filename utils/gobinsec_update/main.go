// Copyright (c) 2022 Proton AG
//
// This file is part of Proton Mail Bridge.
//
// Proton Mail Bridge is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Proton Mail Bridge is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with Proton Mail Bridge. If not, see <https://www.gnu.org/licenses/>

package main

import (
	"fmt"
	"github.com/intercloud/gobinsec/gobinsec"
	"io/ioutil"
	"regexp"
	"strings"
)

type Depend struct {
	Name    string
	Version string
}

func loadDependencies(file string) []Depend {
	var dependencies []Depend
	txt, err := ioutil.ReadFile(file)
	if err != nil {
		return dependencies
	}
	re := regexp.MustCompile(`\t[a-zA-Z0-9-\/\.]* v.*`)
	matches := re.FindAllString(string(txt), -1)
	for _, str := range matches {
		withoutTab := strings.Split(str, "\t")
		split := strings.Split(withoutTab[1], " ")
		dependencies = append(dependencies, Depend{split[0], split[1]})
	}
	return dependencies
}

func main() {
	dependencies := loadDependencies("../../go.mod")

	if err := gobinsec.LoadConfig("", true, true, true, true); err != nil {
		panic(err)
	}

	if err := gobinsec.BuildCache(); err != nil {
		panic(err)
	}

	for _, dep := range dependencies {
		fmt.Println("... Checking " + dep.Name + " " + dep.Version)
		dep, err := gobinsec.NewDependency(dep.Name, dep.Version)
		if err != nil {
			panic(err)
		}

		if err := dep.LoadVulnerabilities(); err != nil {
			panic(err)
		}

		if err := gobinsec.CacheInstance.Close(); err != nil {
			panic(err)
		}
	}
}
