package note

import "strings"

// a comment block generator
// see Nothing ( which discards comments )
// and Book ( which compiles comments into a comment block. )
type Taker interface {
	// start recording comments for a new sequence, mapping, or document.
	// every collection in a document must share the same string builder;
	// but each should probably have its own unique taker.
	// passing nil will disable comment collection.
	BeginCollection(*strings.Builder)
	// record a comment
	// returns error if the the type of comment was unexpected for the current context
	Comment(Type, string) error
	// separates comments for each term within a collection
	// ( terms in a sequence are indicated by a dash
	//   terms in a mapping are indicated by a signature style key )
	NextTerm()
	// stop recording comments for this collection
	// probably best to not reuse the taker after this call.
	EndCollection()
	// return the unified comment block for a collection.
	// initially true, if BeginCollection had been passed a valid string buffer.
	// subsequent calls may return false.
	Resolve() (ret string, okay bool)
}
