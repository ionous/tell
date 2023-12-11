package charmed

import (
	"strings"

	"github.com/ionous/tell/charm"
	"github.com/ionous/tell/runes"
)

// scans until the matching quote marker is found
func ScanQuote(match rune, useEscapes bool, onDone func(string)) (ret charm.State) {
	d := QuoteDecoder{indent: -1}
	return charm.Step(d.ScanQuote(match, useEscapes),
		charm.OnExit("recite", func() (err error) {
			onDone(d.String())
			return
		}))
}

//

// wraps a string builder to read a quoted string or heredoc.
type QuoteDecoder struct {
	strings.Builder
	indent int
}

// read until an InterpretedString (") end marker is found
// for heredocs: pass the indentation of the starting quote
func (d *QuoteDecoder) Interpret() charm.State {
	return d.ScanQuote(runes.InterpretQuote, true)
}

// read until an RawString (`) end marker is found
// for heredocs: pass the indentation of the starting quote
func (d *QuoteDecoder) Record() charm.State {
	return d.ScanQuote(runes.RawQuote, false)
}

// return a state which reads until the end of string, returns error if finished incorrectly
func (d *QuoteDecoder) ScanQuote(match rune, useEscapes bool) charm.State {
	return charm.Self("scanQuote", func(self charm.State, q rune) (ret charm.State) {
		switch {
		case q == match: // the second quote
			ret = charm.Statement("quoted", func(third rune) (ret charm.State) {
				// when heredocs are disabled; return unhandled on the rune after the closing quote.
				if d.indent >= 0 && third == match {
					ret = decodeHere(&d.Builder, d.indent, match, useEscapes)
				}
				return
			})

		case q == runes.Escape && useEscapes:
			ret = decodeEscape(d, self)

		case q == runes.Newline || q == runes.Eof:
			e := InvalidRune(q)
			ret = charm.Error(e)

		default:
			d.WriteRune(q)
			ret = self // loop...
		}
		return
	})
}
