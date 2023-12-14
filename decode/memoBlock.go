package decode

import (
	"strings"

	"github.com/ionous/tell/runes"
)

// replaces comments *strings.Builder
type memoBlock struct {
	strings.Builder
	// have the comment markers been written?
	// wroteKey can happen after the key, up to interkey
	// wroteValue can happen after the value, up to intervalue
	markerCount int
	// incremented at the start of every interkey,
	// whenever markerCount is 0
	emptyKeys int
	memos     *memo
}

func (m *memoBlock) started(memos *memo) {
	m.memos = memos
}

func (m *memoBlock) End() {
	if m.memos != nil {
		m.memos.ended(m)
		m.memos = nil
	}
}

// on interKey / returning from a sub-collection
func (m *memoBlock) endTerm() {
	if m.markerCount > 0 {
		m.markerCount = 0
		m.emptyKeys++
	}
}

func (m *memoBlock) writePadding(markers int) {
	if m.emptyKeys > 0 {
		for i := 0; i < m.emptyKeys; i++ {
			m.WriteRune(runes.NextRecord)
		}
		m.emptyKeys = 0
	}
	if m.markerCount < markers {
		for i := m.markerCount; i < markers; i++ {
			m.WriteRune(runes.KeyValue)
		}
		m.markerCount = markers
	}
}

func appendBuffer(dst, src *strings.Builder) {
	if src.Len() > 0 {
		appendLine(dst, src.String())
		src.Reset()
	}
}

func appendLine(out *strings.Builder, str string) {
	if out.Len() > 0 {
		out.WriteRune(runes.Newline)
	}
	out.WriteString(str)
}
