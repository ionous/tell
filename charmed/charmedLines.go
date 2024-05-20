package charmed

import (
	"strings"

	"github.com/ionous/tell/charm"
	"github.com/ionous/tell/runes"
)

// the decoder produces indented tokens
// it writes lines other than the closing tag into the provided buffer.
// depth can be -1 for fully blank lines
type lineReporter func(lineType rune, indent int) error

// helper context to limit parameter passing
type hereLines struct {
	out    *indentedLines
	report lineReporter
	endTag []rune // custom end tag that has to match exactly
	endSet []rune // individual runes that can match to close
}

// endTag should be free of escapes and whitespace.
func decodeCustomTag(out *indentedLines, endTag []rune, report lineReporter) charm.State {
	d := hereLines{
		out:    out,
		endTag: endTag, // the default endTag is three quotes
		report: report,
	}
	return d.decodeLines()
}

func decodeTripleTag(out *indentedLines, endSet []rune, report lineReporter) charm.State {
	d := hereLines{
		out:    out,
		endSet: endSet,
		report: report,
	}
	return d.decodeLines()
}

// reports lines until the closing tag, then returns nil.
func (d *hereLines) decodeLines() charm.State {
	var indent int
	return charm.Self("decodeLines", func(self charm.State, q rune) (ret charm.State) {
		switch q {
		case runes.Space:
			indent++
			ret = self
		case runes.Newline:
			d.out.nextLine(0, 0)
			indent = 0
			ret = self
		case runes.Eof: // expects a closing tag
			e := charm.InvalidRune(q)
			ret = charm.Error(e)
		default:
			ret = charm.RunState(q, d.decodeLeft(indent))
		}
		return
	})
}

// match at the start of a line, past any initial whitespace.
// depth is the size of that whitespace.
func (d *hereLines) decodeLeft(depth int) charm.State {
	var accum strings.Builder
	custom := customTagMatcher{endTag: d.endTag}
	triple := tripleTagMatcher{endSet: d.endSet}
	return charm.Self("decodeLeft", func(self charm.State, q rune) (ret charm.State) {
		cm, tm := custom.match(q), triple.match(q)
		switch {
		case cm == tagSucceeded:
			if e := d.report(0, depth); e != nil {
				ret = charm.Error(e)
			}

		case tm == tagSucceeded:
			if e := d.report(triple.curr, depth); e != nil {
				ret = charm.Error(e)
			}

		case (cm == tagFailed && tm == tagFailed) || runes.IsWhitespace((q)):
			// whitespace that wasn't a success or if both tags have failed
			// then we are done.
			d.out.WriteString(accum.String())
			ret = charm.RunState(q, d.decodeRight(depth))

		default: // still some tagProgress going on.
			accum.WriteRune(q)
			ret = self
		}
		return
	})
}

// after trying to read the closing tag at the start of a line, and failing.
func (d *hereLines) decodeRight(depth int) charm.State {
	var trailingSpaces int // count trailing spacing
	return charm.Self("decodeRight", func(self charm.State, q rune) (ret charm.State) {
		switch q {
		case runes.Space:
			trailingSpaces++
			ret = self
		case runes.Newline:
			d.out.nextLine(depth, trailingSpaces)
			ret = d.decodeLines() // done with this line; read more lines!
		case runes.Eof:
			e := charm.InvalidRune(q) // closing tag required before eof
			ret = charm.Error(e)
		default:
			if trailingSpaces > 0 {
				dupe(d.out, runes.Space, trailingSpaces)
				trailingSpaces = 0
			}
			d.out.WriteRune(q)
			ret = self // keep reading the rest of the line...
		}
		return
	})
}

func dupe(w runes.RuneWriter, q rune, cnt int) {
	for n := 0; n < cnt; n++ {
		w.WriteRune(q)
	}
}
