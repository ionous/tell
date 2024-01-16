package encode

import (
	"fmt"
	r "reflect"
	"sort"
)

// customization for serializing native maps
// r.Value is guaranteed to a kind of reflect.Map
type MapTransform struct {
	keyLess      func(a, b string) bool
	keyTransform func(r.Value) string
}

// return a factory function for the encoder
func (m *MapTransform) Mapper() StartCollection {
	return m.makeMapping
}

// sort keys; by default keys are written sorted as per standard go string rules.
func (m *MapTransform) Sort(t func(a, b string) bool) *MapTransform {
	m.keyLess = t
	return m
}

// change a reflected key into an encodable string
// the default uses reflect Value.String()
func (m *MapTransform) KeyTransform(t func(keys r.Value) string) *MapTransform {
	m.keyTransform = t
	return m
}

// fix: change to support error?
func keyTransform(v r.Value) (ret string) {
	if k := v.Kind(); k != r.String {
		e := fmt.Errorf("map keys must be string, have %s", k)
		panic(e)
	} else {
		ret = v.String()
	}
	return
}

func (m *MapTransform) makeMapping(src r.Value) (retIt Iterator, err error) {
	keyLess := m.keyLess
	if keyLess == nil {
		keyLess = func(a, b string) bool { return a < b }
	}
	xform := m.keyTransform
	if xform == nil {
		xform = keyTransform
	}

	var mk mapKeys
	if keys := src.MapKeys(); len(keys) > 0 {
		// ugly, but simple:
		str := make([]string, len(keys))
		for i, k := range keys {
			str[i] = xform(k)
		}
		mk = mapKeys{str: str, val: keys, keyLess: keyLess}
		sort.Sort(&mk)
	}
	if err == nil {
		retIt = &mapIter{src: src, mapKeys: mk}
	}
	return
}

// not quite sure how to turn string into an interface without something like this. ugh.
var anyBlank = [1]any{""}
var blank = r.ValueOf(anyBlank).Index(0)

type mapIter struct {
	src     r.Value // the native map
	mapKeys mapKeys
	next    int
}

func (m *mapIter) Next() (okay bool) {
	if okay = m.next < m.mapKeys.Len(); okay {
		m.next++
	}
	return
}

func (m *mapIter) getKey() r.Value {
	return m.mapKeys.val[m.next-1]
}

func (m *mapIter) GetKey() string {
	return m.mapKeys.str[m.next-1]
}

func (m *mapIter) GetValue() any {
	return m.GetReflectedValue().Interface()
}

func (m *mapIter) GetReflectedValue() r.Value {
	key := m.getKey()
	return m.src.MapIndex(key)
}
