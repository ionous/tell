package encode

import r "reflect"

// for sequences, value is guaranteed to be a reflect.Slice
// for mappings, value is guaranteed to be a reflect.Map
type StartCollection func(r.Value) (Iterator, error)

// controls serialization when implemented by a value that's being encoded
type TellMapping interface {
	TellMapping() Iterator
}

// controls serialization when implemented by a value that's being encoded
type TellSequence interface {
	TellSequence() Iterator
}

// the key for sequences
const Dashing = "-"

// walk the elements of a collection
type Iterator interface {
	// return true if there are elements left
	Next() bool
	// return "-" for sequences
	GetKey() string
	// can panic if there are no remaining elements
	GetValue() any
}

// if implemented by one of the iterators
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
type Comments interface {
	Next() bool          // called before every element, false if there are no more elements
	GetComment() Comment // valid after next returns true
}
