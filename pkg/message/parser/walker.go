package parser

type Walker struct {
	root *Part

	handlers       []*handler
	defaultHandler HandlerFunc
}

func newWalker(root *Part) *Walker {
	return &Walker{
		root:           root,
		defaultHandler: func(*Part) error { return nil },
	}
}

func (w *Walker) Walk() (err error) {
	return w.walkOverPart(w.root)
}

func (w *Walker) walkOverPart(p *Part) error {
	if err := w.getHandlerFunc(p)(p); err != nil {
		return err
	}

	for _, child := range p.children {
		if err := w.walkOverPart(child); err != nil {
			return err
		}
	}

	return nil
}

func (w *Walker) RegisterDefaultHandler(fn HandlerFunc) *Walker {
	w.defaultHandler = fn
	return w
}

func (w *Walker) RegisterContentTypeHandler(typeRegExp string, fn HandlerFunc) *Walker {
	w.handlers = append(w.handlers, &handler{
		typeRegExp: typeRegExp,
		fn:         fn,
	})

	return w
}

func (w *Walker) RegisterContentDispositionHandler(dispRegExp string, fn HandlerFunc) *Walker {
	w.handlers = append(w.handlers, &handler{
		dispRegExp: dispRegExp,
		fn:         fn,
	})

	return w
}

func (w *Walker) getHandlerFunc(p *Part) HandlerFunc {
	for _, hdl := range w.handlers {
		if hdl.matchPart(p) {
			return hdl.fn
		}
	}

	return w.defaultHandler
}
