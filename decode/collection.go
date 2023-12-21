package decode

import (
	"errors"
	"fmt"

	"github.com/ionous/tell/collect"
	"github.com/ionous/tell/note"
)

type pendingValue interface {
	setKey(string) error
	setValue(any) error
	finalize() any // return the collection
	note.Taker
}

func newMapping(key string, values collect.MapWriter) *pendingMap {
	return &pendingMap{key: key, maps: values}
}

type pendingMap struct {
	key  string
	maps collect.MapWriter
	note.Book
}

func (p *pendingMap) finalize() (ret any) {
	if len(p.key) > 0 {
		p.setValue(nil)
	}
	if str, ok := p.Resolve(); ok {
		p.maps.MapValue("", str)
	}
	return p.maps.GetMap()
}

func (p *pendingMap) setKey(key string) (err error) {
	if len(p.key) > 0 {
		err = fmt.Errorf("unused key %s", p.key)
	} else if len(key) == 0 {
		err = errors.New("cant add indexed elements to map ping")
	} else {
		p.key = key
	}
	return
}

func (p *pendingMap) setValue(val any) (err error) {
	if len(p.key) == 0 {
		err = errors.New("missing key")
	} else {
		p.maps = p.maps.MapValue(p.key, val)
		p.key = ""
	}
	return
}

func newSequence(values collect.SequenceWriter, reserve bool) *pendingSeq {
	var index int
	if reserve {
		index++
	}
	return &pendingSeq{dashed: true, index: index, values: values}
}

type pendingSeq struct {
	dashed   bool
	blockNil bool /// fix: subcase this for arrays?
	values   collect.SequenceWriter
	note.Book
	index int
}

func (p *pendingSeq) finalize() (ret any) {
	// i dont like that this *and* the output can generate an implicit nil
	// maybe "pendingAt" could handle it all in one place.
	if p.dashed && !p.blockNil {
		p.setValue(nil)
	}
	if str, ok := p.Resolve(); ok {
		p.values = p.values.IndexValue(0, str)
	}
	return p.values.GetSequence()
}

func (p *pendingSeq) setKey(key string) (err error) {
	if p.dashed {
		err = fmt.Errorf("expected an element")
	} else if len(key) > 0 {
		err = errors.New("cant add keyed elements to a sequence")
	} else {
		p.dashed = true
		p.blockNil = false
	}
	return
}

func (p *pendingSeq) setValue(val any) (err error) {
	if !p.dashed {
		err = errors.New("expected a dash before adding values to a sequence")
	} else {
		p.values = p.values.IndexValue(p.index, val)
		p.dashed = false
		p.index++
	}
	return
}

// for document scalars
type pendingScalar struct {
	value any
	note.Taker
}

func (p pendingScalar) finalize() any {
	return p.value
}

func (pendingScalar) setKey(key string) error {
	return fmt.Errorf("unexpected key for document scalar %s", key)
}

func (pendingScalar) setValue(val any) (err error) {
	return fmt.Errorf("unexpected value for document scalar %v(%T)", val, val)
}
