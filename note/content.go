package note

import (
	"fmt"
	"strings"

	"github.com/ionous/tell/runes"
)

type content struct {
	out strings.Builder
	bookState
}

// returns false if not commenting
func (b *content) Resolve() (ret string, okay bool) {
	if b.ctx != nil {
		b.EndCollection() // hrm.
		ret = b.out.String()
		b.out.Reset()
		b.bookState = bookState{}
		b.ctx = nil
		okay = true
	}
	return
}

type bookState struct {
	ctx         *Context
	markerCount int
	nextKeys    int
	totalKeys   int
	lastNote    Type
}

func (b *content) BeginCollection(ctx *Context) {
	b.ctx = ctx
	// if there is a comment pending
	// steal it from the shared buffer, and use it as
	// the header of the first element.
	ctx.writeInto(&b.out)
}

func (b *content) EndCollection() {
	b.flushLast()
}

// new key in this block
func (b *content) NextTerm() {
	// note: if there's a sub-collection
	// its begin() will have stolen our buffer away
	b.flushLast()
	b.nextKeys++
	b.totalKeys++
	b.lastNote = None
}

func (b *content) Comment(n Type, str string) (err error) {
	if was := b.lastNote; n < was {
		// NextTerm would normally handle this.
		err = fmt.Errorf("unexpected transition from %q to %q", n, was)
	} else {
		// advanced the comment type?
		if n.withoutInline() != was.withoutInline() {
			if was != None {
				b.flushLast()
			}
			b.lastNote = n
		}
		switch n {
		default:
			b.ctx.append(str)

		case Header:
			// write headers for following terms straight to the output
			// so that they appear with the *current* collection
			// and dont get stolen by the next begin collection.
			if b.totalKeys == 0 {
				b.ctx.append(str)
			} else if b.writeKeys() {
				b.out.WriteString(str)
			} else {
				appendLine(&b.out, str)
			}

		case Footer:
			if was != Footer {
				b.out.WriteRune(runes.NextTerm)
			} else {
				b.out.WriteRune(runes.Newline)
			}
			b.out.WriteString(str)
		}
	}
	return
}

func (b *content) flushLast() {
	// note: no need to write form feeds or markers
	// unless there are some comments pending
	if b.ctx.pending() {
		// form feeds
		b.writeKeys()
		// markers
		if mark := b.lastNote.mark(); mark > 0 {
			if b.markerCount < mark {
				for i := b.markerCount; i < mark; i++ {
					b.out.WriteRune(runes.KeyValue)
				}
				b.markerCount = mark
			}
			// inline vs trailing
			if !b.lastNote.inline() {
				b.out.WriteRune(runes.Newline)
			}
		}
		b.ctx.writeInto(&b.out)
	}
}

func (b *content) writeKeys() (okay bool) {
	if okay = b.nextKeys > 0; okay {
		for i := 0; i < b.nextKeys; i++ {
			b.out.WriteRune(runes.NextTerm)
		}
		b.nextKeys = 0
		b.markerCount = 0
	}
	return
}

func appendLine(out *strings.Builder, str string) {
	if out.Len() > 0 {
		out.WriteRune(runes.Newline)
	}
	out.WriteString(str)
}
