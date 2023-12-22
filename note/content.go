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
	if b.buf != nil {
		b.EndCollection() // hrm.
		ret = b.out.String()
		b.out.Reset()
		b.bookState = bookState{}
		b.buf = nil
		okay = true
	}
	return
}

type bookState struct {
	markerCount int
	nextKeys    int
	totalKeys   int
	buf         *strings.Builder
	lastNote    Type
}

func (b *content) BeginCollection(buf *strings.Builder) {
	b.buf = buf
	// if there is a (footer or prefix) comment pending
	// steal it from the shared buffer, and use it as
	// the header of the first element.
	if buf.Len() > 0 {
		appendLine(&b.out, buf.String())
		buf.Reset()
	}
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
		if n != Footer {
			appendLine(b.buf, str)
		} else {
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
	if str := b.buf.String(); len(str) > 0 {
		b.buf.Reset()
		// form feeds
		if b.nextKeys > 0 {
			for i := 0; i < b.nextKeys; i++ {
				b.out.WriteRune(runes.NextTerm)
			}
			b.nextKeys = 0
			b.markerCount = 0
		}
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
		// the buffered content
		b.out.WriteString(str)
	}
}

func appendLine(out *strings.Builder, str string) {
	if out.Len() > 0 {
		out.WriteRune(runes.Newline)
	}
	out.WriteString(str)
}
