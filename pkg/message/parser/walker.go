package parser

type Walker struct {
	root *Part

	defaultHandler PartHandler
	typeHandlers   map[string]PartHandler
	dispHandlers   map[string]DispHandler
}

type PartHandler func(*Part) error
type DispHandler func(*Part, PartHandler) error

func newWalker(root *Part) *Walker {
	return &Walker{
		root:           root,
		defaultHandler: func(*Part) (err error) { return },
		typeHandlers:   make(map[string]PartHandler),
		dispHandlers:   make(map[string]DispHandler),
	}
}

func (w *Walker) Walk() (err error) {
	return w.root.visit(w)
}

func (w *Walker) WithDefaultHandler(handler PartHandler) *Walker {
	w.defaultHandler = handler
	return w
}
func (w *Walker) WithContentTypeHandler(contType string, handler PartHandler) *Walker {
	w.typeHandlers[contType] = handler
	return w
}

func (w *Walker) WithContentDispositionHandler(contDisp string, handler DispHandler) *Walker {
	w.dispHandlers[contDisp] = handler
	return w
}
