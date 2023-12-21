package decode

import (
	"errors"

	"github.com/ionous/tell/note"
	"github.com/ionous/tell/token"
)

type output struct {
	pendingAt
	stack           pendingStack
	waitingForValue bool
	skipTerm        bool // ie. if its already been processed
}

func (out *output) finalizeAll() (ret any, err error) {
	if _, e := out.uncheckedPop(-1); e != nil {
		err = e
	} else {
		if out.pendingValue != nil { // tbd: error on empty document?
			ret = out.finalize()
		}
	}
	return
}

func (out *output) setPending(at token.Pos, p pendingValue) {
	out.pendingAt = pendingAt{pos: at, pendingValue: p}
}

func (out *output) push(at token.Pos, p pendingValue) {
	out.stack = append(out.stack, out.pendingAt)
	out.setPending(at, p)
}

func (out *output) newTerm() {
	if !out.skipTerm {
		out.NextTerm()
		out.skipTerm = true
	}
}

// add a header comment, and wait for a new key ( because that's all that can follow )
func (out *output) newHeader(at token.Pos, str string) (err error) {
	if e := out.popToIndent(at.X); e != nil {
		err = e
	} else {
		out.newTerm()
		out.Comment(note.Header, str)
	}
	return
}

// the key is for the same or an earlier collection
// writes a nil value, before finding the right collection
func (out *output) newKey(at token.Pos, key string) (err error) {
	if e := out.popToIndent(at.X); e != nil {
		err = e
	} else if e := out.setKey(at.Y, key); e != nil {
		err = e
	} else {
		out.newTerm()
		out.skipTerm = false
	}
	return
}

// exposed for use by normal collections and arrays.
func (out *output) setKey(row int, key string) (err error) {
	if e := out.pendingAt.setKey(key); e != nil {
		err = e
	} else {
		out.pos.Y = row
		out.waitingForValue = true
	}
	return
}

// add a prefix or suffix comment
func (out *output) addComment(baseType note.Type, at token.Pos, str string) {
	noteType := baseType
	if at.Y == out.pos.Y {
		noteType++
	}
	out.Comment(noteType, str)
}

func (out *output) setValue(val any) (err error) {
	out.waitingForValue = false
	return out.pendingAt.setValue(val)
}

// internal: find the collection indicated by the passed indentation:
// could be this collection, or a parent.
// generates an implicit nil if needed
func (out *output) popToIndent(at int) (err error) {
	if out.waitingForValue {
		out.setValue(nil)
	}
	if cnt, e := out.uncheckedPop(at); e != nil {
		err = e
	} else if cnt > 0 && at != out.pos.X {
		err = errors.New("mismatched indent")
	}
	return
}

// internal: returns number of pops; doesnt check that the resulting indent is valid.
func (out *output) uncheckedPop(at int) (ret int, err error) {
	for ; at < out.pos.X && len(out.stack) > 0; ret++ {
		if e := out.popTop(); e != nil {
			err = e
			break
		}
	}
	return
}

// end the current collection or array.
func (out *output) popTop() (err error) {
	out.EndCollection()
	prev := out.finalize()  // finalize the current pending value
	next := out.stack.pop() // move this to pending
	if e := next.setValue(prev); e != nil {
		err = e
	} else {
		out.pendingAt = next
	}
	return
}
