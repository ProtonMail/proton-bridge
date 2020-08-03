package parser

import "regexp"

type HandlerFunc func(*Part) error

type handler struct {
	typeRegExp, dispRegExp string
	fn                     HandlerFunc
}

func (h *handler) matchPart(p *Part) bool {
	return h.matchType(p) || h.matchDisp(p)
}

func (h *handler) matchType(p *Part) bool {
	if h.typeRegExp == "" {
		return false
	}

	t, _, err := p.Header.ContentType()
	if err != nil {
		t = ""
	}

	return regexp.MustCompile(h.typeRegExp).MatchString(t)
}

func (h *handler) matchDisp(p *Part) bool {
	if h.dispRegExp == "" {
		return false
	}

	disp, _, err := p.Header.ContentDisposition()
	if err != nil {
		disp = ""
	}

	return regexp.MustCompile(h.dispRegExp).MatchString(disp)
}
