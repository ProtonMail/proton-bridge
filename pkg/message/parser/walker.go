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
	return w.root.visit(w)
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
