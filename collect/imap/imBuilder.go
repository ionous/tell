package imap

import "github.com/ionous/tell/collect"

// return a builder that generates ItemMap
func Make(reserve bool) collect.MapWriter {
	var cnt int
	if reserve {
		cnt = 1
	}
	return mapBuilder{values: make(ItemMap, cnt)}
}

type mapBuilder struct {
	values ItemMap
}

// panic if adding the blank key but no space for a blank key was reserved.
func (b mapBuilder) MapValue(key string, val any) collect.MapWriter {
	if len(key) == 0 { // there should be only one blank key; at the start
		if len(b.values) == 0 || len(b.values[0].Key) != 0 {
			// could adjust the slice. but the program should know better.
			panic("map doesn't have space for a blank key")
		}
		b.values[0] = MapItem{Value: val}
	}
	b.values = append(b.values, MapItem{key, val})
	return b
}

// returns ItemMap
func (b mapBuilder) GetMap() any {
	return b.values
}
