// Copyright (c) 2024 Proton AG
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
// along with Proton Mail Bridge. If not, see <https://www.gnu.org/licenses/>.

package parser

import "regexp"

type Visitor struct {
	root        *Part
	rules       []*visitorRule
	defaultRule VisitorRule
}

func newVisitor(root *Part, defaultRule VisitorRule) *Visitor {
	return &Visitor{
		root:        root,
		defaultRule: defaultRule,
	}
}

type Visit func(*Part) (interface{}, error)

type VisitorRule func(*Part, Visit) (interface{}, error)

type visitorRule struct {
	re *regexp.Regexp
	fn VisitorRule
}

// RegisterRule defines what to do when visiting a part whose content type
// matches the given regular expression.
// If a part matches multiple rules, the one registered first will be used.
func (v *Visitor) RegisterRule(contentTypeRegex string, fn VisitorRule) *Visitor {
	v.rules = append(v.rules, &visitorRule{
		re: regexp.MustCompile(contentTypeRegex),
		fn: fn,
	})

	return v
}

func (v *Visitor) Visit() (interface{}, error) {
	return v.visit(v.root)
}

func (v *Visitor) visit(p *Part) (interface{}, error) {
	t, _, err := p.ContentType()
	if err != nil {
		return nil, err
	}

	if rule := v.getRuleForContentType(t); rule != nil {
		return rule.fn(p, v.visit)
	}

	return v.defaultRule(p, v.visit)
}

func (v *Visitor) getRuleForContentType(contentType string) *visitorRule {
	for _, rule := range v.rules {
		if rule.re.MatchString(contentType) {
			return rule
		}
	}

	return nil
}
