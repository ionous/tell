package note

import (
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
	b.flushTerm()
}

// new key in this block
func (b *content) NextTerm() {
	// note: if there's a sub-collection
	// its begin() will have stolen our buffer away

	b.flushTerm()
	b.nextKeys++
	b.totalKeys++
}

func (b *content) Comment(n Type, str string) {
	switch n {
	case Header:
		appendLine(b.buf, str)

	case Prefix, PrefixInline:
		if n != PrefixInline {
			b.buf.WriteRune(runes.Newline)
		}
		b.buf.WriteString(str)

	case Suffix, SuffixInline:
		b.writeKeys()
		b.writeHeader()
		b.writePrefix()
		b.writePadding(2)
		if n != SuffixInline {
			b.out.WriteRune(runes.Newline)
		}
		b.out.WriteString(str)

	case Footer:
		b.writeKeys()
		b.writeHeader()
		b.writePrefix()
		if b.lastNote != Footer {
			b.out.WriteRune(runes.NextTerm)
		} else {
			b.out.WriteRune(runes.Newline)
		}
		b.out.WriteString(str)

	default:
		panic("unknown comment")
	}
	b.lastNote = n
}

func (b *content) flushTerm() {
	// if there's a buffer, it might be for the prefix or header.
	// either way, we need to write the form feeds first.
	//
	// FirstKey: # inline prefix
	// # header for next key
	// NextTerm:
	//
	if b.buf.Len() > 0 {
		b.writeKeys()
		b.writeHeader()
		b.writePrefix()
	}
}

func (b *content) writeKeys() {
	if b.nextKeys > 0 {
		for i := 0; i < b.nextKeys; i++ {
			b.out.WriteRune(runes.NextTerm)
		}
		b.nextKeys = 0
		b.markerCount = 0
	}
}

func (b *content) writeHeader() {
	if b.lastNote == Header {
		if str := b.buf.String(); len(str) > 0 {
			b.out.WriteString(str)
			b.buf.Reset()
		}
		b.lastNote = None
	}
}

func (b *content) writePrefix() {
	if b.lastNote.Prefix() {
		if str := b.buf.String(); len(str) > 0 {
			b.writePadding(1)
			b.out.WriteString(str)
			b.buf.Reset()
		}
		b.lastNote = None
	}
}

func (b *content) writePadding(markers int) {
	b.writeKeys()
	if b.markerCount < markers {
		for i := b.markerCount; i < markers; i++ {
			b.out.WriteRune(runes.KeyValue)
		}
		b.markerCount = markers
	}
}

func appendLine(out *strings.Builder, str string) {
	if out.Len() > 0 {
		out.WriteRune(runes.Newline)
	}
	out.WriteString(str)
}
