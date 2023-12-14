package decode

import (
	"strings"

	"github.com/ionous/tell/runes"
)

type memo struct {
	buf strings.Builder  // shared buffer for initially ambiguous attribution.
	doc runes.RuneWriter // hrm.
}

func (m *memo) Keep() bool {
	return m.doc != nil
}

// popped out to an earlier collection
func (m *memo) ended(block *memoBlock) {
	if !m.Keep() {
		return
	}
}

// just started some new collection.
// any buffered comment that is still buffered is a header for the new collection.
func (m *memo) Begin(block *memoBlock) {
	if !m.Keep() {
		return
	}
	block.started(m)
}

func (m *memo) Comment(out *memoBlock, n noteType, str string) {
	if !m.Keep() {
		return
	}
	switch n {
	case NoteHeader:
		appendLine(&out.Builder, str)

	case NoteFooter:

	case NoteInterKey:

	case NotePrefix, NotePrefixInline, NoteSuffix, NoteSuffixInline:
		out.writePadding(n.Padding())
		if n.Newline() {
			out.WriteRune(runes.Newline)
			if n == NoteSuffix {
				out.WriteRune(runes.HTab)
			}
		}
		out.WriteString(str)
	}
}
