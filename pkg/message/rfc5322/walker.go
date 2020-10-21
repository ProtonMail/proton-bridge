// Copyright (c) 2020 Proton Technologies AG
//
// This file is part of ProtonMail Bridge.
//
// ProtonMail Bridge is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// ProtonMail Bridge is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with ProtonMail Bridge.  If not, see <https://www.gnu.org/licenses/>.

package rfc5322

import (
	"fmt"

	"github.com/ProtonMail/proton-bridge/pkg/message/rfc5322/parser"
	"github.com/antlr/antlr4/runtime/Go/antlr"
)

// walker implements parser.BaseRFC5322ParserListener, defining what to do at
// each node while traversing the syntax tree.
// It also implements antlr.DefaultErrorListener, allowing us to react to
// errors encountered while trying to determine the syntax tree of the input.
type walker struct {
	parser.BaseRFC5322ParserListener
	antlr.DefaultErrorListener

	// nodes acts as a stack; when entering a node, it is pushed here, and when
	// exiting a node, it is popped from here.
	nodes []interface{}

	// res holds the result of walking the parse tree.
	res interface{}

	// err holds the error encountered during parsing, if any.
	err error
}

func (w *walker) enter(b interface{}) {
	w.nodes = append(w.nodes, b)
}

func (w *walker) exit() interface{} {
	b := w.nodes[len(w.nodes)-1]
	w.nodes = w.nodes[:len(w.nodes)-1]
	return b
}

func (w *walker) parent() (b interface{}) {
	return w.nodes[len(w.nodes)-1]
}

func (w *walker) SyntaxError(_ antlr.Recognizer, _ interface{}, _, _ int, msg string, _ antlr.RecognitionException) {
	w.err = fmt.Errorf("error parsing rfc5322 input: %v", msg)
}
