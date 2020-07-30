package parser

import "regexp"

type Visitor struct {
	rules    []*rule
	fallback Rule
}

func NewVisitor(fallback Rule) *Visitor {
	return &Visitor{
		fallback: fallback,
	}
}

type Visit func(*Part) (interface{}, error)

type Rule func(*Part, Visit) (interface{}, error)

type rule struct {
	re string
	fn Rule
}

func (v *Visitor) RegisterRule(contentTypeRegex string, fn Rule) *Visitor {
	v.rules = append(v.rules, &rule{
		re: contentTypeRegex,
		fn: fn,
	})

	return v
}

func (v *Visitor) Visit(p *Part) (interface{}, error) {
	t, _, err := p.Header.ContentType()
	if err != nil {
		return nil, err
	}

	if rule := v.getRuleForContentType(t); rule != nil {
		return rule.fn(p, v.Visit)
	}

	return v.fallback(p, v.Visit)
}

func (v *Visitor) getRuleForContentType(contentType string) *rule {
	for _, rule := range v.rules {
		if regexp.MustCompile(rule.re).MatchString(contentType) {
			return rule
		}
	}

	return nil
}
