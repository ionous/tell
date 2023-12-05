package decode

import (
	"errors"

	"github.com/ionous/tell/maps"
)

type output struct {
	pendingAt
	stack pendingStack
}

func (out *output) finalizeAll() (ret any, err error) {
	if e := out.uncheckedPop(-1); e != nil {
		err = e
	} else if out.pendingValue != nil { // tbd: error on empty document?
		ret = out.finalize()
	}
	return
}

func (out *output) push(at int, p pendingValue) {
	out.stack = append(out.stack, out.pendingAt)
	out.setPending(at, p)
}

func (out *output) setPending(at int, p pendingValue) {
	out.pendingAt = pendingAt{indent: at, pendingValue: p}
}

func (out *output) popToIndent(at int) (err error) {
	if e := out.uncheckedPop(at); e != nil {
		err = e
	} else if at != out.indent {
		err = errors.New("mismatched indent")
	}
	return
}

func (out *output) uncheckedPop(at int) (err error) {
	for at < out.indent && len(out.stack) > 0 {
		prev := out.finalize()
		next := out.stack.pop()
		if e := next.setValue(prev); e != nil {
			err = e
			break
		} else {
			out.pendingAt = next
		}
	}
	return
}

type mapMaker struct {
	create maps.BuilderFactory
}

func (f *mapMaker) newCollection(key string) pendingValue {
	var p pendingValue
	switch {
	case len(key) == 0:
		p = newSequence()
	default:
		//keepComments := !notes.IsNothing(doc.notes)
		p = newMapping(key, f.create(false))
	}
	return p
}
