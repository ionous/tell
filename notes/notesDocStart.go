package notes

import (
	"github.com/ionous/tell/charm"
	"github.com/ionous/tell/runes"
)

// the starting point of every tell document.
// assumes all comments are document headers
// until seeing a nested comment or a fully blank line.
// after "opting in" -- keeps one paragraphed buffered
// ready to be given to the first collection as an element header.
// if there is no such collection, treats the buffered para as header.
type docStartDecoder struct {
	*context
	inlineScalar,
	newCollection makeState
}

func docStart(ctx *context, inlineScalar, newCollection makeState) charm.State {
	d := docStartDecoder{ctx, inlineScalar, newCollection}
	return charm.Step(d.awaitHeader(), d.awaitValue())
}

// parent state: awaits for doc scalar or collection
func (d *docStartDecoder) awaitValue() charm.State {
	return charm.Statement("awaitValue", func(q rune) (ret charm.State) {
		switch q {
		case runeKey:
			ret = d.newCollection()
		case runeValue:
			ret = d.inlineScalar()
		default:
			ret = charm.Error(invalidRune("awaitValue", q))
		}
		return
	})
}

// child state: awaits to decode the first header comment;
// allow parent state to decode values
func (d *docStartDecoder) awaitHeader() (ret charm.State) {
	return charm.Self("awaitHeader", func(self charm.State, q rune) (ret charm.State) {
		switch q {
		case runeParagraph:
			ret = d.firstLines()
		case runes.Hash:
			ret = charm.RunState(q, d.firstLines())
		case runes.Newline:
			ret = self // keep looping on fully blank lines
		default:
			// unhandled
		}
		return
	})
}

// output runes until the end of line,
// then shift to extending the header.
func (d *docStartDecoder) firstLines() charm.State {
	return readLine("firstLines", d.out, d.extendHeader)
}

// immediately after the first comment:
// keep reading paragraphs of header, nest the header, or switch to buffering lines.
// allow parent state to decode values
func (d *docStartDecoder) extendHeader() charm.State {
	return charm.Statement("extendHeader", func(q rune) (ret charm.State) {
		switch q {
		case runeParagraph:
			d.out.WriteRune(runes.Newline)
			ret = d.firstLines()
		case runes.Newline: // fully blank line
			ret = d.awaitParagraph() // buf is empty, so dont need to flush
		case runes.Hash: // nested header, switch to buffering after done
			ret = nestLine("nestHeader", d.out, d.awaitNest)
		default:
			// unhandled
		}
		return
	})
}

// after a blank line, start looking for new paragraphs.
// allow parent state to decode values,
func (d *docStartDecoder) awaitParagraph() charm.State {
	if d.buf.Len() > 0 {
		panic("expects the buffer has been flushed to the doc header")
	}
	return awaitParagraph("awaitParagraph", d.bufferLine)
}

// keep nesting the output, or start buffering.
// allow parent state to decode values
func (d *docStartDecoder) awaitNest() charm.State {
	return charm.Statement("awaitNest", func(q rune) (ret charm.State) {
		switch q {
		case runes.Hash: // nest
			ret = nestLine("nestHeader", d.out, d.awaitNest)
		default:
			ret = d.awaitBuf().NewRune(q)
		}
		return
	})
}

// buffer runes until the end of line, then wait for the next buffered line.
func (d *docStartDecoder) bufferLine() charm.State {
	return readLine("bufferLine", &d.buf, d.awaitBuf)
}

// keep the buffer filled with a maximum of one paragraph.
// allow parent state to decode values
func (d *docStartDecoder) awaitBuf() charm.State {
	return charm.Statement("awaitBuf", func(q rune) (ret charm.State) {
		switch q {
		case runes.Hash: // nest into the current buffer
			ret = nestLine("nestBuffer", &d.buf, d.awaitBuf)
		case runeParagraph:
			d.flush(runes.Newline) // flush
			ret = d.bufferLine()   // and begin buffering a new paragraph
		case runes.Newline:
			d.flush(runes.Newline)   // flush
			ret = d.awaitParagraph() // and begin waiting for a new paragraph
		}
		return
	})
}
