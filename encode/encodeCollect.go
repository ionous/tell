package encode

import r "reflect"

// controls serialization when implemented by an encoding value
// see also MappingFactory
type TellMapping interface {
	TellMapping() MappingIter
}

// controls serialization when implemented by an encoding value
// see also SequenceFactory
type TellSequence interface {
	TellSequence() SequenceIter
}

// factory function for serializing native maps
// r.Value is guaranteed to a kind of reflect.Map
// the default encoder uses SortedMap or SortedMapFactory.
// returning a nil iterator skips the value
type MappingFactory func(r.Value) (MappingIter, error)

// factory function for serializing slices and arrays.
// the default encoder uses Sequence.
// returning a nil iterator skips the value
type SequenceFactory func(r.Value) (SequenceIter, error)

// turns a value representing one or more comments
// into an iterator. the encoder uses the iterator to generate comments for collections.
type CommentFactory func(r.Value) (CommentIter, error)

type MappingIter interface {
	Next() bool // called before every element, false if there are no more elements
	GetKey() string
	GetValue() any // valid after next returns true
}

type SequenceIter interface {
	Next() bool    // called before every element, false if there are no more elements
	GetValue() any // valid after next returns true
}

// if implemented by the implementation of MappingIter or SequenceIter
// will be used instead of GetValue()
type GetReflectedValue interface {
	GetReflectedValue() r.Value
}

type Comment struct {
	Header []string // before the key
	Prefix []string // the key comment, between the key and value
	Suffix []string // trailing comment, the value
}

// comment access for collections
type CommentIter interface {
	Next() bool          // called before every element, false if there are no more elements
	GetComment() Comment // valid after next returns true
}
