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
	// called before a new comment is read.
	// indicates a comment that sits above a value
	// with the same indentation of that value.
	// used for document headers, key comments, element headers.
	OnParagraph() Commentator
	// a value has started decoding; even a nil value should trigger this.
	// any inline comments will live to the right of the value on the same line
	// there can only be one inline comment, continued with optional nesting.
	OnScalarValue() Commentator
	// a new collection has started decoding
	// not expected for the document itself.
	// although it is not explicitly prevented,
	// no runes or paragraphs are expected directly after starting a collection.
	OnBeginCollection() Commentator
	// a signature or dash in a collection has finished decoding.
	// paragraphs can get treated as key comments,
	// or headers for sub collection elements
	// depending on the number of paragraphs and the following value.
	OnKeyDecoded() Commentator
	// footer comments sit below a value, slightly indented.
	// like "paragraph" this can be called multiple times, once for each new footnote;
	// the footer never nests.( an aesthetic choice. )
	OnFootnote() Commentator
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
