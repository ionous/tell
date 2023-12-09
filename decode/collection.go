package decode

import (
	"errors"
	"fmt"
	"strings"

	"github.com/ionous/tell/collect"
)

type pendingValue interface {
	setKey(string) error
	setValue(any) error
	finalize() any // return the collection
}

func newMapping(key string, values collect.MapWriter, comments *strings.Builder) pendingValue {
	return &pendingMap{key: key, maps: values, comments: comments}
}

type pendingMap struct {
	key      string
	maps     collect.MapWriter
	comments *strings.Builder
}

func (p *pendingMap) finalize() (ret any) {
	if len(p.key) > 0 {
		p.setValue(nil)
	}
	if p.comments != nil {
		str := clearComments(&p.comments)
		p.maps.MapValue("", str)
	}
	return p.maps.GetMap()
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
		p.maps = p.maps.MapValue(p.key, val)
		p.key = ""
	}
	return
}

func newSequence(values collect.SequenceWriter, comments *strings.Builder) pendingValue {
	var index int
	if comments != nil {
		index++
	}
	return &pendingSeq{dashed: true, index: index, values: values, comments: comments}
}

type pendingSeq struct {
	dashed   bool
	values   collect.SequenceWriter
	comments *strings.Builder
	index    int
}

func (p *pendingSeq) finalize() (ret any) {
	if p.dashed {
		p.setValue(nil)
	}
	if p.comments != nil {
		str := clearComments(&p.comments)
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

func clearComments(a **strings.Builder) (ret string) {
	ret = (*a).String()
	(*a).Reset()
	(*a) = nil
	return
}
