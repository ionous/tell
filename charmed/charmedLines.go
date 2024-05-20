package charmed

import (
	"strings"

	"github.com/ionous/tell/charm"
	"github.com/ionous/tell/runes"
)

// helper context to limit parameter passing
type hereLines struct {
	out      *strings.Builder // final target for the heredoc
	lines    indentedBlock    // accumulation of the heredocs
	lineType rune             // describes how to process the incoming text
	endTag   []rune           // custom end tag that has to match exactly
	endSet   []rune           // individual runes that can match to close
}

// decode until a custom tag has been reached
// endTag should be free of escapes and whitespace.
func decodeUntilCustom(out *strings.Builder, lineType rune, endTag []rune) charm.State {
	d := hereLines{
		out:      out,
		lineType: lineType,
		endTag:   endTag, // the default endTag is three quotes
	}
	return d.decodeLines()
}

// decode until a triplet of any of the passed tags has been reached
func decodeUntilTriple(out *strings.Builder, endSet ...rune) charm.State {
	d := hereLines{
		out:    out,
		endSet: endSet,
	}
	return d.decodeLines()
}

// record lines until the closing tag, then returns nil.
// switches to decodeLeft when something other than whitespace is detected.
func (d *hereLines) decodeLines() charm.State {
	var indent int
	return charm.Self("decodeLines", func(self charm.State, q rune) (ret charm.State) {
		switch q {
		case runes.Space:
			indent++
			ret = self
		case runes.Newline:
			d.lines.flushLine(0, 0)
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

func (d *hereLines) report(lineType rune, indent int) error {
	return d.lines.writeHere(d.out, lineType, indent)
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
			// its a closing custom tag
			if e := d.report(d.lineType, depth); e != nil {
				ret = charm.Error(e)
			}

		case tm == tagSucceeded:
			// its a closing triple tag
			if e := d.report(triple.curr, depth); e != nil {
				ret = charm.Error(e)
			}

		case (cm == tagFailed && tm == tagFailed) || runes.IsWhitespace((q)):
			// its a line of content
			d.lines.WriteString(accum.String())
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
			d.lines.flushLine(depth, trailingSpaces)
			ret = d.decodeLines() // done with this line; read more lines!
		case runes.Eof:
			e := charm.InvalidRune(q) // closing tag required before eof
			ret = charm.Error(e)
		default:
			if trailingSpaces > 0 {
				dupe(&d.lines, runes.Space, trailingSpaces)
				trailingSpaces = 0
			}
			d.lines.WriteRune(q)
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
