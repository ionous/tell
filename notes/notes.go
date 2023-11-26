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
	// precedes a comment that sits above a value
	// with the same indentation of that value.
	// trailing comments never use ¶.
	OnParagraph() Commentator
	// a value has started decoding; even a nil value should trigger this.
	// any inline comments will live to the right of the value on the same line
	// there can only be one inline comment, continued with optional nesting.
	OnScalarValue() Commentator
	// a signature or dash in a collection has finished decoding.
	// paragraphs can get treated as key comments,
	// or headers for sub collection elements
	// depending on the number of paragraphs and the following value.
	//
	// fix? GetComments() is the implicit "EndCollection" --
	// maybe better would be an explicit EndCollection that hands back the resolver or comments
	OnKeyDecoded() Commentator
}

// receive text from the decoded comments of a tell document.
// its signature mirrors strings.StringBuilder.
type RuneWriter interface {
	// each comment should start with a hash and space, and should end with a newline.
	// without an intervening OnParagraph, comments after a newline automatically "nest".
	WriteRune(rune) (int, error)
}

// pull finished comments from the commentator
type Resolver interface {
	// end of the current collection
	// returns its comment block
	GetComments() string
}
