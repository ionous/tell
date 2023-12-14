package decode

import (
	"errors"

	"github.com/ionous/tell/collect"
	"github.com/ionous/tell/token"
)

type output struct {
	pendingAt
	stack           pendingStack
	waitingForValue bool
}

func (out *output) finalizeAll() (ret any, err error) {
	if _, e := out.uncheckedPop(-1); e != nil {
		err = e
	} else if out.pendingValue != nil { // tbd: error on empty document?
		ret = out.finalize()
	}
	return
}

func (out *output) push(at token.Pos, p pendingValue) {
	out.stack = append(out.stack, out.pendingAt)
	out.setPending(at, p)
}

func (out *output) setKey(row int, key string) (err error) {
	if e := out.pendingAt.setKey(key); e != nil {
		err = e
	} else {
		out.pos.Y = row
		out.waitingForValue = true
	}
	return
}

func (out *output) setValue(val any) (err error) {
	out.waitingForValue = false
	return out.pendingAt.setValue(val)
}

func (out *output) setPending(at token.Pos, p pendingValue) {
	out.pendingAt = pendingAt{pos: at, pendingValue: p}
}

// returns number of pops
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

// returns number of pops; doesnt check that the resulting indent is valid.
func (out *output) uncheckedPop(at int) (ret int, err error) {
	for ; at < out.pos.X && len(out.stack) > 0; ret++ {
		if e := out.popTop(); e != nil {
			err = e
			break
		}
	}
	return
}

func (out *output) popTop() (err error) {
	prev := out.finalize()  // finalize the current pending value
	next := out.stack.pop() // move this to pending
	if e := next.setValue(prev); e != nil {
		err = e
	} else {
		out.pendingAt = next
	}
	return
}

type collector struct {
	maps collect.MapFactory
	seqs collect.SequenceFactory
	memo memo
}

func (f *collector) newCollection(key string) pendingValue {
	var p pendingValue
	keepComments := f.memo.Keep()
	switch {
	case len(key) == 0:
		p = newSequence(f.seqs(keepComments), keepComments)
	default:
		p = newMapping(key, f.maps(keepComments))
	}
	f.memo.Begin(p.comments())
	return p
}

func (f *collector) newArray() pendingValue {
	keepComments := f.memo.Keep()
	seq := newSequence(f.seqs(keepComments), keepComments)
	f.memo.Begin(seq.comments())
	seq.blockNil = true
	return seq
}
