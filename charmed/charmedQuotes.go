package charmed

import (
	"fmt"
	"strings"

	"github.com/ionous/tell/charm"
	"github.com/ionous/tell/runes"
)

// wraps a string builder to read a quoted string
type QuoteDecoder struct {
	strings.Builder
}

// read until an InterpretedString (") end marker is found
func (d *QuoteDecoder) Interpret() charm.State {
	return d.ScanQuote(runes.InterpretQuote, true)
}

// read until an RawString (`) end marker is found
func (d *QuoteDecoder) Record() charm.State {
	return d.ScanQuote(runes.RawQuote, false)
}

// return a state which reads until the end of string, returns error if finished incorrectly
func (d *QuoteDecoder) ScanQuote(match rune, useEscapes bool) charm.State {
	const escape = '\\'
	return charm.Self("scanQuote", func(self charm.State, q rune) (ret charm.State) {
		switch {
		case q == match:
			// returns unhandled for the net rune:
			ret = charm.Statement("quoted",
				func(rune) charm.State { return nil })

		case q == escape && useEscapes: // alt: could use Step and keep.
			ret = charm.Statement("escaping", func(q rune) (ret charm.State) {
				if x, ok := escapes[q]; !ok {
					e := fmt.Errorf("unknown escape %s", runes.RuneName(q))
					ret = charm.Error(e)
				} else {
					d.WriteRune(x)
					ret = self // loop...
				}
				return
			})

		case q == runes.Newline || q == runes.Eof:
			e := fmt.Errorf("unexpected %s", runes.RuneName(q))
			ret = charm.Error(e)

		default:
			d.WriteRune(q)
			ret = self // loop...
		}
		return
	})
}

// scans until the matching quote marker is found
func ScanQuote(match rune, useEscapes bool, onDone func(string)) (ret charm.State) {
	var d QuoteDecoder
	return charm.Step(d.ScanQuote(match, useEscapes), charm.OnExit("recite", func() {
		onDone(d.String())
	}))
}

var escapes = map[rune]rune{
	'a':  '\a',
	'b':  '\b',
	'f':  '\f',
	'n':  '\n',
	'r':  '\r',
	't':  '\t',
	'v':  '\v',
	'\\': '\\',
	'"':  '"',
}
