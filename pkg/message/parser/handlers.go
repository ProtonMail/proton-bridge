package parser

type PartHandlerFunc func(*Part) error
type DispHandlerFunc func(*Part, PartHandlerFunc) error

type PartHandler struct {
	enter, exit PartHandlerFunc
}

func NewPartHandler() *PartHandler {
	return &PartHandler{
		enter: partNoop,
		exit:  partNoop,
	}
}

func (h *PartHandler) OnEnter(fn PartHandlerFunc) *PartHandler {
	h.enter = fn
	return h
}

func (h *PartHandler) OnExit(fn PartHandlerFunc) *PartHandler {
	h.exit = fn
	return h
}

func (h *PartHandler) handleEnter(_ *Walker, p *Part) error {
	return h.enter(p)
}

func (h *PartHandler) handleExit(_ *Walker, p *Part) error {
	return h.exit(p)
}

type DispHandler struct {
	enter, exit DispHandlerFunc
}

func NewDispHandler() *DispHandler {
	return &DispHandler{
		enter: dispNoop,
		exit:  dispNoop,
	}
}

func (h *DispHandler) OnEnter(fn DispHandlerFunc) *DispHandler {
	h.enter = fn
	return h
}

func (h *DispHandler) OnExit(fn DispHandlerFunc) *DispHandler {
	h.exit = fn
	return h
}

func (h *DispHandler) handleEnter(w *Walker, p *Part) error {
	// NOTE: This is hacky -- is there a better solution?
	return h.enter(p, func(p *Part) error {
		return w.getTypeHandler(p).handleEnter(w, p)
	})
}

func (h *DispHandler) handleExit(w *Walker, p *Part) error {
	// NOTE: This is hacky -- is there a better solution?
	return h.exit(p, func(p *Part) error {
		return w.getTypeHandler(p).handleExit(w, p)
	})
}

func partNoop(*Part) error { return nil }

func dispNoop(*Part, PartHandlerFunc) error { return nil }
