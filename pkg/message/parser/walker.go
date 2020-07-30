package parser

type Walker struct {
	root *Part

	defaultHandler handler
	typeHandlers   map[string]handler
	dispHandlers   map[string]handler
}

type handler interface {
	handleEnter(*Walker, *Part) error
	handleExit(*Walker, *Part) error
}

func newWalker(root *Part) *Walker {
	return &Walker{
		root:           root,
		defaultHandler: NewPartHandler(),
		typeHandlers:   make(map[string]handler),
		dispHandlers:   make(map[string]handler),
	}
}

func (w *Walker) Walk() (err error) {
	return w.visitPart(w.root)
}

func (w *Walker) visitPart(p *Part) (err error) {
	hdl := w.getHandler(p)

	if err = hdl.handleEnter(w, p); err != nil {
		return
	}

	for _, child := range p.children {
		if err = w.visitPart(child); err != nil {
			return
		}
	}

	return hdl.handleExit(w, p)
}

func (w *Walker) WithDefaultHandler(handler handler) *Walker {
	w.defaultHandler = handler
	return w
}
func (w *Walker) RegisterContentTypeHandler(contType string) *PartHandler {
	hdl := NewPartHandler()

	w.typeHandlers[contType] = hdl

	return hdl
}

func (w *Walker) RegisterContentDispositionHandler(contDisp string) *DispHandler {
	hdl := NewDispHandler()

	w.dispHandlers[contDisp] = hdl

	return hdl
}

func (w *Walker) getHandler(p *Part) handler {
	if dispHandler := w.getDispHandler(p); dispHandler != nil {
		return dispHandler
	}

	return w.getTypeHandler(p)
}

// getTypeHandler returns the appropriate PartHandler to handle the given part.
// If no specialised handler exists, it returns the default handler.
func (w *Walker) getTypeHandler(p *Part) handler {
	t, _, err := p.Header.ContentType()
	if err != nil {
		return w.defaultHandler
	}

	hdl, ok := w.typeHandlers[t]
	if !ok {
		return w.defaultHandler
	}

	return hdl
}

// getDispHandler returns the appropriate DispHandler to handle the given part.
// If no specialised handler exists, it returns nil.
func (w *Walker) getDispHandler(p *Part) handler {
	t, _, err := p.Header.ContentDisposition()
	if err != nil {
		return nil
	}

	return w.dispHandlers[t]
}
