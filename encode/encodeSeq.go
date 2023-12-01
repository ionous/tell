package encode

import (
	r "reflect"
)

type rseq struct {
	slice r.Value
	next  int
}

// called with values of reflect kinds Slice and Array
func OrderedSequence(enc *Encoder, src r.Value) SequenceIter {
	return &rseq{slice: src}
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

func (m *rseq) GetComment() Comment {
	return Comment{}
}
