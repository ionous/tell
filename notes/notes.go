package notes

// comment block creation
// split into three interfaces for documentation
// no methods return error because, in theory,
// the decoding machine should never make invalid requests.
type Commentator interface {
	Events
	RuneWriter
	Resolver
}

// events indicate different sections of a tell document
// during decoding. the method return "self" for chaining.
type Events interface {
	// a new collection has been *discovered*.
	OnBeginCollection() Commentator
	// an element of a collection has finished decoding
	OnTermDecoded() Commentator
	// called before a header gets read.
	// header comments sit above a value
	// at the same indentation of its term
	// each header starts a left-justified comment
	// reused to indicate *potential* headers in the padding area.
	OnBeginHeader() Commentator
	// a signature or dash has finished decoding.
	// padding is the space between
	// the key ( or dash ) and the value
	// it can contain a single comment, continued with optional nesting.
	OnKeyDecoded() Commentator
	// a value has started decoding
	// inline comments will live to the right of the value on the same line
	// there can only be one inline comment, continued with optional nesting.
	OnScalarValue() Commentator
	// called before a header gets read.
	// footer comments sit below a value
	// at the same indentation of the value.
	// can be called multiple times, once for each new footnote;
	// the footer doesnt nest.( an aesthetic choice. )
	OnBeginFooter() Commentator
}

// receive text from the decoded comments of a tell document.
type RuneWriter interface {
	// lines must be ended with a newline before other comments can be added.
	// writing after a newline "nests" the subsequent comment
	// the signature the StringBuilder interface
	WriteRune(rune) (int, error)
}

// pull finished comments from the commentator
type Resolver interface {
	// end of the current collection
	// returns its comment block
	GetComments() string
}
