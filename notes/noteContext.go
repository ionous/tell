package notes

import (
	"strings"

	"github.com/ionous/tell/runes"
)

type context struct {
	out   pendingBlock
	stack stack
	// alt: each state could have its own buffer
	// the code uses some implicit hand offs across states
	// letting things stick in the buf and on the start of the new state flushing it.
	// ( ex. buffered doc headers, or inter key comments which get pulled into the next element )
	buf            strings.Builder
	nextCollection RuneWriter // from BeginCollection
}

// fix: might be cleaner for tell to have a "BeginCollection" for document too
// especially because it sends a mismatched EndCollection to flush the document...
// ( rather than passing the runewriter at the start )
func newContext(w RuneWriter) *context {
	ctx := &context{nextCollection: w, out: makeBlock(w)}
	return ctx
}

func (p *context) newBlock() {
	var w RuneWriter
	w, p.nextCollection = p.nextCollection, nil
	if w == nil {
		panic("missing begin collection?")
	}
	prev, next := p.out, makeBlock(w)
	p.stack.push(prev) // remember the former block
	p.out = next
	p.flush(-1) // write the current buffer to out ( the new collection comment )
}

func (p *context) resolveBuffer() (ret string) {
	if ret = p.buf.String(); len(ret) > 0 {
		p.buf.Reset()
	}
	return
}

// write passed runes, and then the buffer, to out
func (p *context) flush(q rune) {
	if str := p.resolveBuffer(); len(str) > 0 {
		p.out.writeTerms()
		writeBuffer(&p.out, str, q)
	}
}

// called on end collection.
func (p *context) pop() {
	// any buffer right now is for the parent container
	parent := p.stack.pop()
	p.out = parent
	// and now it has those contents
	p.flush(runes.Newline)
}
