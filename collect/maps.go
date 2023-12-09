package collect

// write to a map
type MapWriter interface {
	// add the passed pair to the in-progress map
	// returns a new writer ( not guaranteed to be the original one )
	MapValue(key string, val any) MapWriter
	// return the implementation specific representation of a map
	GetMap() any
}

// a function which returns a new writer
// reserve indicates whether to keep space for a comment key
type MapFactory func(reserve bool) MapWriter

// a function which returns a new writer
// reserve indicates whether to keep space for comments
type SequenceFactory func(reserve bool) SequenceWriter

type SequenceWriter interface {
	// add the passed value to the in-progress sequence
	// returns a new writer ( not guaranteed to be the original one )
	// indices are guaranteed to increase by one each time
	// ( stating with 1 if reserve was true )
	// except for comments, which are written last at index 0
	IndexValue(idx int, val any) SequenceWriter
	// return the implementation specific representation of a sequence
	GetSequence() any
}
