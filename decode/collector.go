package decode

import (
	"strings"

	"github.com/ionous/tell/collect"
)

// factory for collections, arrays, and comments
type collector struct {
	maps          collect.MapFactory
	seqs          collect.SequenceFactory
	keepComments  bool
	commentBuffer strings.Builder
}

func (f *collector) newCollection(key string) pendingValue {
	var p pendingValue
	switch {
	case len(key) == 0:
		p = newSequence(f.seqs(f.keepComments), f.keepComments)
	default:
		p = newMapping(key, f.maps(f.keepComments))
	}
	if f.keepComments {
		p.BeginCollection(&f.commentBuffer)
	}
	return p
}

func (f *collector) newArray() pendingValue {
	seq := newSequence(f.seqs(f.keepComments), f.keepComments)
	seq.blockNil = true
	if f.keepComments {
		seq.BeginCollection(&f.commentBuffer)
	}
	return seq
}
