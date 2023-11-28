package notes

import (
	"github.com/ionous/tell/runes"
)

type context struct {
	out   *pendingBlock
	stack stack
	// alt: each state could have its own buffer
	// the code uses some implicit hand offs across states
	// letting things stick in the buf and on the start of the new state flushing it.
	// ( ex. buffered doc headers, or inter key comments which get pulled into the next element )
	buf stringsBuilder
	res string
}

func newContext() *context {
	return &context{out: new(pendingBlock)}
}

func (p *context) newBlock() {
	prev, next := p.out, new(pendingBlock)
	p.stack.push(prev) // remember the former block
	p.out = next
	p.flush(-1) // write the current buffer to out ( the new collection comment )
}

// write passed runes, and then the buffer, to out
func (p *context) flush(q rune) {
	if str := p.buf.Resolve(); len(str) > 0 {
		p.out.writeTerms()
		writeBuffer(p.out, str, q)
	}
}

// called on end collection.
func (p *context) pop() {
	// whatever we have is what we have
	p.res = p.out.Resolve()
	// any buffer right now is for the parent container
	parent := p.stack.pop()
	p.out = parent
	// and now it has those contents
	p.flush(runes.Newline)
}
