package charmed

import (
	"strings"

	"github.com/ionous/tell/charm"
	"github.com/ionous/tell/runes"
)

// scans until the matching quote marker is found
func ScanQuote(match rune, escape bool, onDone func(string)) (ret charm.State) {
	var d QuoteDecoder
	return charm.Step(d.ScanQuote(match, escape, false),
		charm.OnExit("recite", func() {
			onDone(d.String())
		}))
}

// wraps a string builder to read a quoted string or heredoc.
type QuoteDecoder struct {
	strings.Builder
}

func (d *QuoteDecoder) Decode(q rune) (ret charm.State) {
	if n := d.Pipe(q); n != nil {
		ret = n
	} else if n = d.Interpret(q); n != nil {
		ret = n
	} else if n = d.Record(q); n != nil {
		ret = n
	}
	return
}

// assumes q is a pipe rune
// read until a heredoc ending marker is found
func (d *QuoteDecoder) Pipe(q rune) (ret charm.State) {
	if q == runes.YamlBlock {
		ret = charm.Self("pipe whitespace", func(self charm.State, q rune) (ret charm.State) {
			switch q {
			case runes.Space: // ignore spaces
				ret = self
			case runes.Newline: // we expect to see a newline after the pipe
				var lines indentedLines
				endSet := []rune{runes.RawQuote, runes.InterpretQuote}
				ret = decodeTripleTag(&lines, endSet,
					func(lineType rune, indent int) error {
						var escape bool
						if lineType == runes.InterpretQuote {
							escape = true
						} else if lineType != runes.RawQuote {
							panic("unexpected lineType")
						}
						return lines.writeLines(&d.Builder, indent, escape)
					})
			default:
				ret = charm.Error(charm.InvalidRune(q))
			}
			return
		})
	}

	return
}

// read until an InterpretedString (") end marker is found
// for heredocs: pass the indentation of the starting quote
func (d *QuoteDecoder) Interpret(q rune) (ret charm.State) {
	if q == runes.InterpretQuote {
		ret = d.ScanQuote(q, true, true)
	}
	return
}

// read until an RawString (`) end marker is found
// for heredocs: pass the indentation of the starting quote
func (d *QuoteDecoder) Record(q rune) (ret charm.State) {
	if q == runes.RawQuote {
		ret = d.ScanQuote(runes.RawQuote, false, true)
	}
	return
}

// return a state which reads until the end of string, returns error if finished incorrectly
func (d *QuoteDecoder) ScanQuote(match rune, escape, allowHere bool) charm.State {
	return charm.Self("scanQuote", func(self charm.State, q rune) (ret charm.State) {
		switch {
		case q == match: // the second quote
			ret = charm.Statement("quoted", func(third rune) (ret charm.State) {
				// when heredocs are disabled; return unhandled on the rune after the closing quote.
				if allowHere && third == match {
					ret = decodeHereAfter(&d.Builder, match, escape)
				}
				return
			})

		case q == runes.Escape && escape:
			ret = charm.Step(decodeEscape(d), self)

		case q == runes.Newline || q == runes.Eof:
			e := charm.InvalidRune(q)
			ret = charm.Error(e)

		default:
			d.WriteRune(q)
			ret = self // loop...
		}
		return
	})
}
