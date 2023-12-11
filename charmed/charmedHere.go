package charmed

import (
	"errors"
	"strings"

	"github.com/ionous/tell/charm"
	"github.com/ionous/tell/runes"
)

// three quotes have been found:
// 1. skip until the end of line; ( future: read the heredoc closing tag. )
// 2. read here doc lines until the closing tag
//
func decodeHere(out *strings.Builder, indent int, quote rune, useEscapes bool) charm.State {
	var end = [...]rune{quote, quote, quote} // future: <<<END
	d := hereDoc{end: end[:], useEscapes: useEscapes}
	// fix? ugly.
	return skipToNextLine(charm.Step(
		d.decodeLines(),
		// after decode lines has finished ( returned unhandled )
		// flush the heredoc out once the closing indentation is known.
		charm.OnExit("rightHereWriteNow", func() error {
			return d.flush(out)
		})))
}

type hereDoc struct {
	end        []rune // closing tag
	useEscapes bool
	// the line currently being decoded
	buf   strings.Builder
	depth int
	// previously decoded lines
	lines []docLine
}

// track the indentation of each line
type docLine struct {
	depth int
	str   string
}

// at the beginning of a heredoc line:
// 1. eat indentation ( too little is an error)
// 2. check for the closing tag
// 3. read/escape rest of line
// loop or return unhandled after reading the heredoc closing tag and an eol
func (d *hereDoc) decodeLines() charm.State {
	return charm.Step(d.decodeDepth(), d.decodeLeft())
}

func (d *hereDoc) flush(out *strings.Builder) (err error) {
	if d.buf.Len() > 0 {
		// tbd: panic? because this should never happen
		err = errors.New("unexpected trailing heredoc text")
	} else {
		depth := d.depth
		for _, el := range d.lines {
			if spaces := el.depth - depth; spaces < 0 {
				err = errors.New("bad indent") // FIX -- would want the full pos here; or at least line of doc
				break
			} else {
				for i := 0; i < spaces; i++ {
					out.WriteRune(runes.Space)
				}
				out.WriteString(el.str)
			}
		}
	}
	return
}

// returns unhandled ( success ) on a matching closing tag.
func (d *hereDoc) decodeLeft() charm.State {
	return charm.Statement("decodeLeft", func(q rune) charm.State {
		// try to match the closing tag, cnt is the number of matched characters.
		return decodeEnding(d.end, func(cnt int) (ret charm.State) {
			// on mismatch, write the parts that did match, then read the rest of the line.
			if cnt < len(d.end) {
				for i := 0; i < cnt; i++ {
					d.buf.WriteRune(d.end[i])
				}
				ret = d.decodeRight()
			}
			return
		})
	})
}

// after trying to decode the closing tag, read the rest of the line
// then loop to decode more lines.
func (d *hereDoc) decodeRight() charm.State {
	return charm.Self("decodeRight", func(self charm.State, q rune) (ret charm.State) {
		switch {
		case q == runes.Newline: // done with this line; read more lines!
			line := docLine{depth: d.depth, str: d.buf.String()}
			d.lines = append(d.lines, line)
			d.buf.Reset()
			d.depth = 0
			ret = d.decodeLines()

		case q == runes.Eof: // should have closing tag before eof
			e := InvalidRune(q)
			ret = charm.Error(e)

		case q == runes.Escape && d.useEscapes:
			ret = decodeEscape(&d.buf, self)

		default:
			d.buf.WriteRune(q)
			ret = self // keep reading the current line...
		}
		return
	})
}

// determine the amount of indentation on the current line.
// returns unhandled on the first unknown rune.
// FIX: output newlines when raw, require two newlines for interpreted
func (d *hereDoc) decodeDepth() charm.State {
	return charm.Self("decodeDepth", func(self charm.State, q rune) (ret charm.State) {
		switch q {
		case runes.Space:
			d.depth++
			ret = self
		case runes.Newline:
			d.depth = 0
			ret = self // restart
		case runes.Eof:
			e := InvalidRune(q)
			ret = charm.Error(e)
		}
		return
	})
}

// runs "after" a mismatch of the closing tag
// or returns unhandled ( success for quote strings )
// if the closing tag matched and there was a newline.
func decodeEnding(tag []rune, after func(cnt int) charm.State) charm.State {
	var idx int // index in tag
	return charm.Self("decodeEnding", func(self charm.State, q rune) (ret charm.State) {
		if cnt := len(tag); idx < cnt && tag[idx] == q {
			ret, idx = self, idx+1
		} else if idx < cnt || q != runes.Newline {
			ret = charm.RunState(q, after(idx))
		}
		// otherwise: we have fully matched, and received a newline
		return
	})
}

// read up to, and including a newline;
// eof is an error
func skipToNextLine(next charm.State) charm.State {
	return charm.Self("skipToNextLine", func(self charm.State, q rune) (ret charm.State) {
		switch q {
		case runes.Eof:
			e := InvalidRune(q)
			ret = charm.Error(e)
		case runes.Newline:
			ret = next
		default:
			ret = self
		}
		return
	})
}
