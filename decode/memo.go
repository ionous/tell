package decode

import (
	"strings"
)

type memo struct {
	buf          strings.Builder // shared buffer for initially ambiguous attribution.
	keepComments bool
}

func (m *memo) Keep() bool {
	return m.keepComments
}

// just started some new collection.
// any buffered comment that is still buffered is a header for the new collection.
func (m *memo) Begin(b *CommentBlock) {
	if !m.Keep() {
		return
	}
	b.started(m)
	// steal the buffer for the sub collection
	if m.buf.Len() > 0 {
		appendLine(&b.out, m.buf.String())
		m.buf.Reset()
	}
}
