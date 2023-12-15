package decode

import (
	"strings"

	"github.com/ionous/tell/runes"
)

type CommentBlock struct {
	out strings.Builder
	blockState
}

// returns false if not commenting
func (b *CommentBlock) Resolve() (ret string, okay bool) {
	if b.memos != nil {
		ret = b.out.String()
		b.out.Reset()
		b.blockState = blockState{}
		okay = true
	}
	return
}

type blockState struct {
	markerCount int
	nextKeys    int
	totalKeys   int
	memos       *memo
	lastNote    noteType
}

func (b *CommentBlock) started(memos *memo) {
	b.memos = memos
}

func (b *CommentBlock) End() {
	if b.memos == nil { // ex. comments disabled.
		return
	}
	// differentiate the leading header of a collection
	// from an "inter key" footer ( a final element that never existed )
	//
	// # leading header.
	// # ( shouldnt have anything before it. )
	// -
	// # header for a missing next element becomes a footer.
	// # ( requires a leading form feed. )
	//
	if b.totalKeys > 0 {
		b.nextKeys++
	}
	b.flushTerm()
}

// new key in this block
func (b *CommentBlock) NextKey() {
	if b.memos == nil { // ex. comments disabled.
		return
	}
	// note: if there's a sub-collection
	// its begin() will have stolen our buffer away
	b.nextKeys++
	b.totalKeys++
	b.flushTerm()
}

func (b *CommentBlock) Comment(n noteType, str string) {
	if b.memos == nil { // ex. comments disabled.
		return
	}
	switch n {
	case NoteHeader:
		appendLine(&b.memos.buf, str)

	case NotePrefix, NotePrefixInline:
		if n != NotePrefixInline {
			b.memos.buf.WriteRune(runes.Newline)
		}
		b.memos.buf.WriteString(str)

	case NoteSuffix, NoteSuffixInline:
		b.writeKeys()
		b.writeHeader()
		b.writePrefix()
		b.writePadding(2)
		if n != NoteSuffixInline {
			b.out.WriteRune(runes.Newline)
			b.out.WriteRune(runes.HTab)
		}
		b.out.WriteString(str)

	case NoteFooter:
		b.writeKeys()
		b.writeHeader()
		b.writePrefix()
		if b.lastNote != NoteFooter {
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

func (b *CommentBlock) flushTerm() {
	// if there's a buffer, it might be for the prefix or header.
	// either way, we need to write the form feeds first.
	//
	// FirstKey: # inline prefix
	// # header for next key
	// NextKey:
	//
	if b.memos.buf.Len() > 0 {
		b.writeKeys()
		b.writeHeader()
		b.writePrefix()
	}
}

func (b *CommentBlock) writeKeys() {
	if b.nextKeys > 0 {
		for i := 0; i < b.nextKeys; i++ {
			b.out.WriteRune(runes.NextTerm)
		}
		b.nextKeys = 0
		b.markerCount = 0
	}
}

func (b *CommentBlock) writeHeader() {
	if b.lastNote == NoteHeader {
		if str := b.memos.buf.String(); len(str) > 0 {
			b.out.WriteString(str)
			b.memos.buf.Reset()
		}
		b.lastNote = NoteNone
	}
}

func (b *CommentBlock) writePrefix() {
	if b.lastNote.Prefix() {
		if str := b.memos.buf.String(); len(str) > 0 {
			b.writePadding(1)
			b.out.WriteString(str)
			b.memos.buf.Reset()
		}
		b.lastNote = NoteNone
	}
}

func (b *CommentBlock) writePadding(markers int) {
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
