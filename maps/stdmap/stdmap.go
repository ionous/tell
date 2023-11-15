package stdmap

import "github.com/ionous/tell/maps"

type StdMap map[string]any

func Builder(reserve bool) (ret maps.Builder) {
	if reserve {
		ret = StdMap{"": nil}
	} else {
		ret = make(StdMap)
	}
	return
}

func (m StdMap) Add(key string, val any) maps.Builder {
	if len(key) == 0 { // there should be only one blank key; at the start
		if _, exists := m[key]; !exists {
			// could adjust the slice. but the program should know better.
			panic("map doesn't have space for a blank key")
		}
	}
	m[key] = val
	return m
}

func (m StdMap) Map() any {
	return (map[string]any)(m)
}
