package charmed

import (
	"github.com/ionous/tell/runes"
)

const (
	tagFailed tagStatus = iota - 1
	tagProgress
	tagSucceeded
)

type tagStatus int

// match an exact set of runes at the start of a line;
// plus one rune of whitespace.
type customTagMatcher struct {
	idx    int
	endTag []rune // custom end tag that has to match exactly
}

// match three contiguous occurrences from a set of runes;
// plus one rune of whitespace.
type tripleTagMatcher struct {
	rep    int
	curr   rune
	endSet []rune // individual runes that can match to close
}

func (m *customTagMatcher) match(q rune) (ret tagStatus) {
	switch i, cnt := m.idx, len(m.endTag); {
	case i < 0 || cnt == 0:
		ret = tagFailed
	case i == cnt:
		if !runes.IsWhitespace(q) {
			ret = tagFailed
		} else {
			ret = tagSucceeded
		}
		m.idx = -1 // either way, if called again, fail.
	case i < cnt:
		if m.endTag[i] != q {
			m.idx = -1
			ret = tagFailed
		} else {
			m.idx++ // advance
			ret = tagProgress
		}
	default:
		panic("unhandled")
	}
	return
}

func (m *tripleTagMatcher) match(q rune) (ret tagStatus) {
	switch i := m.rep; {
	case i < 0: // once failed, forever shy
		ret = tagFailed

	case i == 3: // we matched enough; now we need whitespace
		m.rep = -1 // either way, if called again, fail
		if !runes.IsWhitespace(q) {
			ret = tagFailed
		} else {
			ret = tagSucceeded
		}

	case i > 0: // we've been initialized
		if q == m.curr { // check the initial match
			m.rep++
			ret = tagProgress
		} else {
			m.rep = -1
			ret = tagFailed
		}
	default:
		// initialize:
		var found bool
		for _, n := range m.endSet {
			if q == n {
				found = true
				break
			}
		}
		if !found {
			m.rep = -1
			ret = tagFailed
		} else {
			m.curr, m.rep = q, 1
			ret = tagProgress
		}
	}
	return
}
