package encode

import (
	r "reflect"
)

type mapKeys struct {
	str     []string
	val     []r.Value
	keyLess func(a, b string) bool
}

func (m *mapKeys) Len() int {
	return len(m.str)
}

func (m *mapKeys) Less(i, j int) (ret bool) {
	a, b := m.str[i], m.str[j]
	switch {
	case len(a) == 0:
		ret = len(b) > 0
	case len(b) > 0: // if b is blank, 'a' can never be less than it.
		ret = m.keyLess(a, b)
	}
	return
}

func (m *mapKeys) Swap(i, j int) {
	m.str[i], m.str[j] = m.str[j], m.str[i]
	m.val[i], m.val[j] = m.val[j], m.val[i]
}
