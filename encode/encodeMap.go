package encode

import (
	"cmp"
	r "reflect"
	"slices"
)

type sortedMap struct {
	src      r.Value
	keys     []r.Value
	next     int
	comments cit
}

// called with values of reflect kind Map
func SortedMap(enc *Encoder, src r.Value) MappingIter {
	var comments cit
	keys := src.MapKeys()
	if len(keys) > 0 {
		slices.SortFunc(keys, func(a, b r.Value) int {
			return cmp.Compare(a.String(), b.String())
		})
		// fix? for the sake of tapestry,
		// allow comment keys to be other than ""?
		// it uses double dash...
		if first := keys[0]; first.IsZero() {
			// tbd: how much error handling or recovery should there be?
			// and should "not keep comments" be "skip comments"?
			if enc.keep {
				v := src.MapIndex(first)
				comments = makeComments(v.String())
			}
			// tbd: always skip the first blank key?
			keys = keys[1:]
		}
	}
	return &sortedMap{src: src, keys: keys, comments: comments}
}

func (m *sortedMap) Next() (okay bool) {
	if okay = m.next < len(m.keys); okay {
		m.comments.Next() // advance comments without checking if done
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
	return m.GetReflectedValue().Interface()
}

func (m *sortedMap) GetReflectedValue() r.Value {
	key := m.getKey()
	return m.src.MapIndex(key)
}

func (m *sortedMap) GetComment() Comment {
	return m.comments.GetComment()
}
