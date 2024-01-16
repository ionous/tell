package orderedmap

import "github.com/ionous/tell/encode"

// implements the encoder mapping
// modification of the map during iteration may yield surprising results.
func (o OrderedMap) TellMapping() encode.Iterator {
	return &mapIter{o: &o}
}

type mapIter struct {
	o    *OrderedMap
	next int
}

func (m *mapIter) Next() (okay bool) {
	if okay = m.next < len(m.o.keys); okay {
		m.next++
	}
	return
}

func (m *mapIter) GetKey() string {
	return m.o.keys[m.next-1]
}

func (m *mapIter) GetValue() any {
	k := m.GetKey()
	return m.o.values[k]
}
