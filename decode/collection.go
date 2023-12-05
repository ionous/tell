package decode

import (
	"errors"
	"fmt"

	"github.com/ionous/tell/maps"
)

type pendingValue interface {
	setKey(string) error
	setValue(any) error
	finalize() any // return the collection
}

func newMapping(key string, values maps.Builder) pendingValue {
	return &pendingMap{key: key, values: values}
}

type pendingMap struct {
	values maps.Builder
	key    string
}

func (p *pendingMap) finalize() (ret any) {
	if len(p.key) > 0 {
		p.setValue(nil)
	}
	return p.values.Map()
}

func (p *pendingMap) setKey(key string) (err error) {
	if len(p.key) > 0 {
		err = fmt.Errorf("unused key %s", p.key)
	} else if len(key) == 0 {
		err = errors.New("cant add indexed elements to mapping")
	} else {
		p.key = key
	}
	return
}

func (p *pendingMap) setValue(val any) (err error) {
	if len(p.key) == 0 {
		err = errors.New("missing key")
	} else {
		p.values = p.values.Add(p.key, val)
		p.key = ""
	}
	return
}

func newSequence() pendingValue {
	return &pendingSeq{dashed: true}
}

type pendingSeq struct {
	values []any
	dashed bool
}

func (p *pendingSeq) finalize() (ret any) {
	if p.dashed {
		p.setValue(nil)
	}
	return p.values
}

func (p *pendingSeq) setKey(key string) (err error) {
	if p.dashed {
		err = fmt.Errorf("expected an element")
	} else if len(key) > 0 {
		err = errors.New("cant add keyed elements to a sequence")
	} else {
		p.dashed = true
	}
	return
}

func (p *pendingSeq) setValue(val any) (err error) {
	if !p.dashed {
		err = errors.New("expected a dash before adding values to a sequence")
	} else {
		p.values = append(p.values, val)
		p.dashed = false
	}
	return
}

func newScalar(val any) pendingValue {
	return &pendingScalar{val}
}

// for document scalars
type pendingScalar struct {
	value any
}

func (p *pendingScalar) finalize() any {
	return p.value
}

func (p *pendingScalar) setKey(key string) error {
	return fmt.Errorf("unexpected key for document scalar %s", key)
}

func (p *pendingScalar) setValue(val any) (err error) {
	return fmt.Errorf("unexpected value for document scalar %v(%T)", val, val)
}
