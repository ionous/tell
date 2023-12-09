package stdseq

import "github.com/ionous/tell/collect"

type StdSequence []any

func Make(reserve bool) (ret collect.SequenceWriter) {
	if reserve {
		ret = StdSequence{""}
	} else {
		ret = StdSequence{} // nil doesnt work. needs some place for the interface value i guess.
	}
	return
}

func (m StdSequence) IndexValue(idx int, val any) (ret collect.SequenceWriter) {
	if idx < len(m) {
		m[idx] = val
		ret = m
	} else {
		ret = append(m, val)
	}
	return
}

// returns []any
func (m StdSequence) GetSequence() any {
	return ([]any)(m)
}
