package stdmap

import "github.com/ionous/tell/collect"

type StdMap map[string]any

func Make(reserve bool) (ret collect.MapWriter) {
	if reserve {
		ret = StdMap{"": nil}
	} else {
		ret = make(StdMap)
	}
	return
}

func (m StdMap) MapValue(key string, val any) collect.MapWriter {
	if len(key) == 0 { // there should be only one blank key; at the start
		if _, exists := m[key]; !exists {
			// could adjust the slice. but the program should know better.
			panic("map doesn't have space for a blank key")
		}
	}
	m[key] = val
	return m
}

// returns map[string]any
func (m StdMap) GetMap() any {
	return (map[string]any)(m)
}
