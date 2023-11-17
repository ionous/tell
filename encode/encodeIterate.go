package encode

import (
	"cmp"
	r "reflect"
	"slices"
	"strings"
)

type rseq struct {
	slice r.Value
	next  int
}

func (m *rseq) Next() (okay bool) {
	if okay = m.next < m.slice.Len(); okay {
		m.next++
	}
	return
}

func (m *rseq) GetValue() any {
	at := m.next - 1
	return m.slice.Index(at).Interface()
}

// --
type sortedMap struct {
	src  r.Value
	keys []r.Value
	next int
}

func makeSortedMap(src r.Value) *sortedMap {
	keys := src.MapKeys()
	slices.SortFunc(keys, func(a, b r.Value) int {
		return cmp.Compare(strings.ToLower(a.String()),
			strings.ToLower(b.String()))
	})
	return &sortedMap{src: src, keys: keys}
}

func (m *sortedMap) Next() (okay bool) {
	if okay = m.next < len(m.keys); okay {
		m.next++
	}
	return
}

func (m *sortedMap) getKey() r.Value {
	return m.keys[m.next-1]
}

func (m *sortedMap) GetKey() string {
	at := m.getKey()
	return at.String()
}

func (m *sortedMap) GetValue() any {
	key := m.getKey()
	return m.src.MapIndex(key).Interface()
}

// --
// question of whether to encode the struct name
// ( thindex would help for a custom deserializer to handle pointers )
// but currently, loading only happens by map anyway

// type rstruct struct {
// 	obj   r.Value
// 	field int
// }

// func (m *rstruct) Next() (okay bool) {
// 	return m.it.Next()
// }

// func (m *rstruct) GetKey() string {
// 	return m.it.Key().String()
// }

// func (m *rstruct) GetValue() (ret any) {
// 	return m.it.Value().Interface()
// }
