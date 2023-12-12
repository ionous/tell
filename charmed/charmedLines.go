package charmed

import (
	"strings"

	"github.com/ionous/tell/charm"
	"github.com/ionous/tell/runes"
)

type lineType int

//go:generate stringer -type=lineType
const (
	lineText lineType = iota
	lineClose
)

// the decoder produces indented tokens
// it writes lines other than the closing tag into the provided buffer.
// depth can be -1 for fully blank lines
type lineReporter func(lineType lineType, lhs, rhs int) error

type hereLines struct {
	report lineReporter
	escape bool
	endTag []rune
}

// endTag should be free of escapes and whitespace.
func decodeLines(out *strings.Builder, escape bool, endTag []rune, report lineReporter) charm.State {
	d := hereLines{
		escape: escape,
		endTag: endTag, // the default endTag is three quotes
		report: report,
	}
	return d.decodeLines(out)
}

// reports lines until the closing tag, then returns nil.
func (d *hereLines) decodeLines(out *strings.Builder) charm.State {
	var indent int
	return charm.Self("decodeLines", func(self charm.State, q rune) (ret charm.State) {
		switch q {
		case runes.Space:
			indent++
			ret = self
		case runes.Newline:
			d.report(lineText, 0, 0)
			indent = 0
			ret = self
		case runes.Eof: // expects a closing tag
			e := InvalidRune(q)
			ret = charm.Error(e)
		default:
			ret = charm.RunState(q, d.decodeLeft(out, indent))
		}
		return
	})
}

func (d *hereLines) decodeLeft(out *strings.Builder, depth int) charm.State {
	var idx int // index in tag
	return charm.Self("decodeLeft", func(self charm.State, q rune) (ret charm.State) {
		if cnt := len(d.endTag); idx < cnt && d.endTag[idx] == q {
			// still matching; advance.
			ret, idx = self, idx+1
		} else if idx < cnt || q != runes.Newline {
			// mismatched: write out all the runes that did match
			// ( since those are part of the lines )
			for i := 0; i < idx; i++ {
				out.WriteRune(d.endTag[i])
			}
			ret = charm.RunState(q, d.decodeRight(out, depth))
		} else {
			// otherwise: we have fully matched, and received a newline
			d.report(lineClose, depth, 0)
			ret = charm.UnhandledNext()
		}
		return
	})
}

// after trying to read the closing tag ( and failing )
func (d *hereLines) decodeRight(out *strings.Builder, depth int) charm.State {
	var trailingSpaces int // count trailing spacing
	return charm.Self("decodeRight", func(self charm.State, q rune) (ret charm.State) {
		switch q {
		case runes.Space:
			trailingSpaces++
		case runes.Newline:
			d.report(lineText, depth, trailingSpaces)
			ret = d.decodeLines(out) // done with this line; read more lines!
		case runes.Eof:
			e := InvalidRune(q) // closing tag required before eof
			ret = charm.Error(e)
		default:
			if trailingSpaces > 0 {
				dupe(out, runes.Space, trailingSpaces)
				trailingSpaces = 0
			}
			if q == runes.Escape && d.escape {
				ret = charm.Step(decodeEscape(out), self)
			} else {
				out.WriteRune(q)
				ret = self // keep reading the rest of the line...
			}
		}
		return
	})
}

func dupe(w runes.RuneWriter, q rune, cnt int) {
	for n := 0; n < cnt; n++ {
		w.WriteRune(q)
	}
}
