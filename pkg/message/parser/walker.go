// Copyright (c) 2023 Proton AG
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

func (w *Walker) Walk() error {
	return w.walkOverPart(w.root)
}
func (w *Walker) WalkSkipAttachment() error {
	return w.walkOverPartSkipAttachment(w.root)
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

func (w *Walker) walkOverPartSkipAttachment(p *Part) error {
	if err := w.getHandlerFuncSkipAttachment(p)(p); err != nil {
		return err
	}

	for _, child := range p.children {
		if err := w.walkOverPartSkipAttachment(child); err != nil {
			return err
		}
	}

	return nil
}

// RegisterDefaultHandler registers a handler that will be called on every part
// that doesn't match a registered content type/disposition handler.
func (w *Walker) RegisterDefaultHandler(fn HandlerFunc) *Walker {
	w.defaultHandler = fn
	return w
}

// RegisterContentTypeHandler registers a handler that will be called when a
// part's content type matches the given regular expression.
// If a part matches multiple handlers, the one registered first will be chosen.
func (w *Walker) RegisterContentTypeHandler(typeRegExp string, fn HandlerFunc) *Walker {
	w.handlers = append(w.handlers, &handler{
		typeRegExp: regexp.MustCompile(typeRegExp),
		fn:         fn,
	})

	return w
}

// RegisterContentDispositionHandler registers a handler that will be called
// when a part's content disposition matches the given regular expression.
// If a part matches multiple handlers, the one registered first will be chosen.
func (w *Walker) RegisterContentDispositionHandler(dispRegExp string, fn HandlerFunc) *Walker {
	w.handlers = append(w.handlers, &handler{
		dispRegExp: regexp.MustCompile(dispRegExp),
		fn:         fn,
	})

	return w
}

func (w *Walker) getHandlerFunc(p *Part) HandlerFunc {
	for _, handler := range w.handlers {
		if handler.matchPart(p) {
			return handler.fn
		}
	}

	return w.defaultHandler
}

func (w *Walker) getHandlerFuncSkipAttachment(p *Part) HandlerFunc {
	for _, handler := range w.handlers {
		if handler.matchPartSkipAttachment(p) {
			return handler.fn
		}
	}

	return w.defaultHandler
}
