# Outline
The `rfc5322` package implements a parser for `address-list` and `date-time` strings, as defined in RFC5322.
It also supports encoded words (RFC2047) and has international tokens (RFC6532).

# `rfc5322/parser` directory
The lexer and parser are generated using ANTLR4.
The grammar is defined in the g4 files:
- RFC5322Parser.g4 defines the parser grammar,
- RFC5322Lexer.g4 defines the lexer grammar.

These grammars are derived from the ABNF grammar provided in the RFCs above, 
albeit with some relaxations added to support "nonstandard" (bad) input.

Running `antlr4` on these g4 files generates a parser which recognises strings conforming to the grammar:
- rfc5322_lexer.go
- rfc5322parser_base_listener.go
- rfc5322_parser.go
- rfc5322parser_listener.go

The generated parser can then be used to convert a valid address/date into an abstract syntax tree.

# `rfc5322` directory
Once we have an abstract syntax tree, we must turn it into something usable, namely a `mail.Address` or `time.Time`.

The generated code in the `rfc5322/parser` directory implements a walker.
This walker walks over the abstract syntax tree, 
calling a callback when entering and another when when exiting each node.
By default, the callbacks are no-ops, unless they are overridden.

## `walker.go`
The `walker` type extends the base walker, overriding the default no-op callbacks
to do something specific when entering and exiting certain nodes. 

The goal of the walker is to traverse the syntax tree, picking out relevant information from each node's text.
For example, when parsing a `mailbox` node, the relevant information to pick out from the parse tree is the
name and address of the mailbox. This information can appear in a number of different ways, e.g. it might be
RFC2047 word-encoded, it might be a string with escaped chars that need to be handled, it might have comments
that should be ignored, and so on.

So while walking the syntax tree, each node needs to ask its children what their "value" is.
The `mailbox` needs to ask its child nodes (either a `nameAddr` node or an `addrSpec` node)
what the name and address are.
If the child node is a `nameAddr`, it needs to ask its `displayName` child what the name is
and the `angleAddr` what the address is; these in turn ask `word` nodes, `addrSpec` nodes, etc.

Each child node is responsible for telling its parent what its own value is.
The parent is responsible for assembling the children into something useful.

Ideally, this would be done with the visitor pattern. But unfortunately, the generated parser only
provides a walker interface. So we need to make use of a stack, pushing on nodes when we enter them
and popping off nodes when we exit them, to turn the walker into a kind of visitor.

## `parser.go`
This file implements two methods, 
`ParseAddressList(string) ([]*mail.Address, error)` 
and
`ParseDateTime(string) (time.Time, error)`.

These methods set up a parser from the raw input, start the walker, and convert the walker result
into an object of the correct type.


# Example: Parsing `dateTime`
Parsing a date-time is rather simple. The implementation begins in `date_time.go`. The abridged code is below:

```
type dateTime struct {
	year   int
	...
}

func (dt *dateTime) withYear(year *year) {
	dt.year = year.value
}

...

func (w *walker) EnterDateTime(ctx *parser.DateTimeContext) {
	w.enter(&dateTime{
		loc: time.UTC,
	})
}

func (w *walker) ExitDateTime(ctx *parser.DateTimeContext) {
	dt := w.exit().(*dateTime)
	w.res = time.Date(dt.year, ...)
}
```

As you can see, when the walker reaches a `dateTime` node, it pushes a `dateTime` object onto the stack:
```
w.enter(&dateTime{
	loc: time.UTC,
})
```

and when it leaves a `dateTime` node, it pops it off the stack, 
converting it from `interface{}` to the concrete type,
and uses the parsed `dateTime` values like day, month, year etc 
to construct a go `time.Time` object to set the walker result:
```
dt := w.exit().(*dateTime)
w.res = time.Date(dt.year, ...)
```

These parsed values were discovered while the walker continued to walk across the date-time node.

Let's see how the walker discovers the `year`.
Here is the abridged code of what happens when the walker enters a `year` node:
```
type year struct {
	value int
}

func (w *walker) EnterYear(ctx *parser.YearContext) {
	var text string

	for _, digit := range ctx.AllDigit() {
		text += digit.GetText()
	}

	val, err := strconv.Atoi(text)
	if err != nil {
		w.err = err
	}

	w.enter(&year{
		value: val,
	})
}
```

When entering the `year` node, it collects all the raw digits, which are strings, then
converts them to an integer, and sets that as the year's integer value while pushing it onto the stack.

When exiting, it pops the year off the stack and gives itself to the parent (now on the top of the stack).
It doesn't know what type of object the parent is, it just checks to see if anything above it on the stack
is expecting a `year` node:
```
func (w *walker) ExitYear(ctx *parser.YearContext) {
	type withYear interface {
		withYear(*year)
	}

	res := w.exit().(*year)

	if parent, ok := w.parent().(withYear); ok {
		parent.withYear(res)
	}
}
```

In our case, the `date` is expecting a `year` node because it implements `withYear`,
```
func (dt *dateTime) withYear(year *year) {
	dt.year = year.value
}
```
and that is how the `dateTime` data members are collected.

