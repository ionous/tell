package decode

import (
	"errors"
	"strings"

	"github.com/ionous/tell/collect"
	"github.com/ionous/tell/token"
)

type output struct {
	pendingAt
	stack pendingStack
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
	}
	return
}

func (out *output) setPending(at token.Pos, p pendingValue) {
	out.pendingAt = pendingAt{pos: at, pendingValue: p}
}

// returns number of pops
func (out *output) popToIndent(at int) (ret int, err error) {
	if cnt, e := out.uncheckedPop(at); e != nil {
		err = e
	} else if cnt > 0 && at != out.pos.X {
		err = errors.New("mismatched indent")
	} else {
		ret = cnt
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
}

func (f *collector) newCollection(key string, comments *strings.Builder) pendingValue {
	var p pendingValue
	switch {
	case len(key) == 0:
		p = newSequence(f.seqs(comments != nil), comments)
	default:
		p = newMapping(key, f.maps(comments != nil), comments)
	}
	return p
}

func (f *collector) newArray(comments *strings.Builder) pendingValue {
	seq := newSequence(f.seqs(comments != nil), comments)
	seq.blockNil = true
	return seq
}
