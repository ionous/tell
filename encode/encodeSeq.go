package encode

import (
	"fmt"
	r "reflect"
)

type SequenceTransform struct{}

// return a factory function for the encoder
func (n *SequenceTransform) Sequencer() Collection {
	return sequenceStarter(n.makeSequence)
}

// todo? sort values; by default sequences are not sorted
// func (m *SequenceTransform) Sort(t func(a, b r.Value) bool) {
// 	m.sort = t
// }

func (n *SequenceTransform) makeSequence(src r.Value) (ret Iterator, err error) {
	if e := validateSeq(src); e != nil {
		err = e
	} else if cnt := src.Len(); cnt > 0 {
		ret = &rseq{slice: src}
	}
	return
}

func validateSeq(src r.Value) (err error) {
	if k := src.Kind(); k != r.Slice && k != r.Array {
		err = fmt.Errorf("slices must be of interface type, have %s(%s)", k, src.Type())
	}
	return
}

type rseq struct {
	slice r.Value
	next  int
}

// always returns "-" for sequences
func (m *rseq) GetKey() string {
	return Dashing
}

func (m *rseq) Next() (okay bool) {
	if okay = m.next < m.slice.Len(); okay {
		m.next++
	}
	return
}

func (m *rseq) GetValue() any {
	return m.GetReflectedValue().Interface()
}

func (m *rseq) GetReflectedValue() r.Value {
	at := m.next - 1
	return m.slice.Index(at)
}
