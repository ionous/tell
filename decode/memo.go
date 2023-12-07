package decode

import (
	"strings"

	"github.com/ionous/tell/notes"
	"github.com/ionous/tell/runes"
	"github.com/ionous/tell/token"
)

func makeMemo(n notes.Commentator) memo {
	m := memo{Commentator: n}
	m.noteAt = m.memoDocHeader()
	return m
}

// memo relays document events to package notes
// using different rules for each section of a document:
// categorizing comments into "normal" and "nested",
// and separating inline comments from any following
// comments using an intervening newline.
type memo struct {
	notes.Commentator
	noteAt memoState
}

type memoState func(el, hash token.Pos, str string)

func (m *memo) popped(cnt int) {
	if cnt > 0 {
		for i := 0; i < cnt; i++ {
			m.Commentator.OnCollectionEnded()
		}
		m.noteAt = m.memoValue()
	}
}

func (m *memo) newComments() (ret *strings.Builder) {
	keepComments := !notes.IsNothing(m.Commentator)
	if keepComments {
		ret = new(strings.Builder)
	}
	m.BeginCollection(ret)
	m.noteAt = m.memoKey()
	return
}

// custom message for this custom case:
func (m *memo) OnDocScalar() {
	m.Commentator.OnScalarValue()
	m.noteAt = m.memoDocScalar()
}

func (m *memo) OnScalarValue() notes.Commentator {
	m.Commentator.OnScalarValue()
	m.noteAt = m.memoValue()
	return m
}

func (m *memo) OnKeyDecoded() notes.Commentator {
	m.Commentator.OnKeyDecoded()
	m.noteAt = m.memoKey()
	return m
}

// everything in a document before the value.
// positions greater than 0 are considered nested.
//
// # ....
// <value>
//
func (m *memo) memoDocHeader() memoState {
	return func(el, hash token.Pos, str string) {
		m.nestComment(0, hash.X, str)
		return
	}
}

// the very last region of a document following a *scalar* value.
// ( comments after a document level collection usually
//   wind up being buffered via inter key. )
//
// has the same behavior as the document header:
// positions greater than 0 are considered nested.
//
// # ....
// <value>
//
func (m *memo) memoDocFooter() memoState {
	return m.memoDocHeader()
}

// the region directly following a document level scalar.
// starts with an optional inline comment,
// everything else is nested until
// it switches to "footer" after a blank line,
// or a zero (fully left aligned) indent.
//
// "value" # <--- trailing inline
//   # <-- nested (aka. a trailing block)
//
func (m *memo) memoDocScalar() memoState {
	return m.checkInline(func(el, hash token.Pos, str string) {
		// switches to block after a blank line, or on zero indent.
		// noting that blank lines always have zero indents.
		if hash.Y == el.Y {
			m.writeComment(str)
		} else if hash.X == 0 || len(str) == 0 {
			m.writeComment(str)
			m.noteAt = m.memoDocHeader()
		} else {
			m.nestComment(0, 1, str)
		}
		return
	})
}

// the region following a scalar value in a mapping or sequence
// until a blank line or the first collection aligned comment
// ( and then switches to "inter key" )
//
// has an optional inline comment;
// everything else is considered nested.
//
// - "value" # <--- trailing inline
//   # <-- nested (aka. a trailing block)
//
func (m *memo) memoValue() memoState {
	return m.checkInline(func(el, hash token.Pos, str string) {
		// switches to block after a blank line, or on zero indent.
		// noting that blank lines always have zero indents.
		if hash.Y == el.Y {
			m.writeComment(str)
		} else if hash.X == el.X || len(str) == 0 {
			m.writeComment(str)
			m.noteAt = m.memoInterKey()
		} else {
			m.nestComment(0, 1, str)
		}
		return
	})
}

// the region following a key until its value;
// relies on the decoder's handling of an collection aligned comment,
// to generate an implicit nil, and switch to interKey.
//
// has an optional inline comment,
// subsequent lines can be normal or nested;
// more than two spaces of indent is considered "nested".
//
// Key: # < -- inline
//   # <-- trailing block
// # .... < -- decoder triggers a nil value
//
func (m *memo) memoKey() memoState {
	return m.checkInline(func(el, hash token.Pos, str string) {
		if hash.Y == el.Y {
			m.writeComment(str)
		} else {
			m.nestComment(el.X+2, hash.X, str)
		}
		return
	})
}

// the region between sibling keys;
// after the trailing comments of a value,
// until the next key ( or end of collection. )
//
// everything deeper than the collection's indent
// is considered nested.
//
// - <value or implicit nil>
// # <-- here and after.
//
func (m *memo) memoInterKey() memoState {
	return func(el, hash token.Pos, str string) {
		m.nestComment(el.X, hash.X, str)
		return
	}
}

// check if the next comment is on the same line as the element
// if not, write a blank line. either way, send the comment to indicated next state.
func (m *memo) checkInline(next memoState) memoState {
	return func(el, hash token.Pos, str string) {
		if hash.Y > el.Y {
			m.writeComment("")
		}
		m.noteAt = next
		next(el, hash, str)
	}
}

func (m *memo) nestComment(threshold, depth int, str string) {
	if depth > threshold && len(str) > 0 {
		m.Commentator.OnNestedComment()
	}
	m.writeComment(str)
}

// fix: just pass the whole thing at once
func (m *memo) writeComment(str string) {
	for _, q := range str {
		m.Commentator.WriteRune(q)
	}
	m.Commentator.WriteRune(runes.Newline)
}
