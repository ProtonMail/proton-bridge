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
	re string
	fn VisitorRule
}

func (v *Visitor) RegisterRule(contentTypeRegex string, fn VisitorRule) *Visitor {
	v.rules = append(v.rules, &visitorRule{
		re: contentTypeRegex,
		fn: fn,
	})

	return v
}

func (v *Visitor) Visit() (interface{}, error) {
	return v.visit(v.root)
}

func (v *Visitor) visit(p *Part) (interface{}, error) {
	t, _, err := p.Header.ContentType()
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
		if regexp.MustCompile(rule.re).MatchString(contentType) {
			return rule
		}
	}

	return nil
}
