package decode

import (
	"github.com/ionous/tell/collect"
	"github.com/ionous/tell/note"
)

// factory for collections, arrays, and comments
type collector struct {
	maps           collect.MapFactory
	seqs           collect.SequenceFactory
	keepComments   bool
	commentContext note.Context
}

func (f *collector) newCollection(key string) pendingValue {
	var p pendingValue
	switch {
	case len(key) == 0:
		p = f.newSequence()
	default:
		p = f.newMapping(key)
	}
	if f.keepComments {
		p.BeginCollection(&f.commentContext)
	}
	return p
}

func (f *collector) newSequence() *pendingSeq {
	return newSequence(f.seqs(f.keepComments), f.keepComments)
}

func (f *collector) newMapping(key string) *pendingMap {
	return newMapping(key, f.maps(f.keepComments))
}

func (f *collector) newArray() pendingValue {
	seq := f.newSequence()
	seq.blockNil = true
	if f.keepComments {
		seq.BeginCollection(&f.commentContext)
	}
	return seq
}

func isMapping(c pendingValue) (okay bool) {
	_, okay = c.(*pendingMap)
	return
}

func isSequence(c pendingValue) (okay bool) {
	_, okay = c.(*pendingSeq)
	return
}
