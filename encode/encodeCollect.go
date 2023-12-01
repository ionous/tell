package encode

import r "reflect"

type Mapper interface {
	TellMapping(enc *Encoder) MappingIter
}

type Sequencer interface {
	TellSequence(enc *Encoder) SequenceIter
}

type MappingIter interface {
	Next() bool
	GetKey() string
	GetValue() any
	GetComment() Comment
}

type SequenceIter interface {
	Next() bool
	GetValue() any
	GetComment() Comment
}

// if implemented by the implementation of MappingIter or SequenceIter
// will be used instead of GetValue()
type GetReflectedValue interface {
	GetReflectedValue() r.Value
}

type Comment struct {
	// doesnt distinguish between "sub header" and "nested header"
	// rewriting should probably prefer "sub header"
	Header       []string
	OnKeyDecoded []string
	Inline       []string
	Footer       []string
}

// comment access for collections
type CommentIter interface {
	Next() bool
	GetComment() Comment
}
