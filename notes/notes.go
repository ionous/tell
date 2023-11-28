package notes

// comment block creation
// split into three interfaces for documentation
// no methods return error because, in theory,
// the decoding machine should never make invalid requests.
type Commentator interface {
	Events
	RuneWriter
}

// events indicate different sections of a tell document
// during decoding. the method return "self" for chaining.
type Events interface {
	// an explicit request to nest the next comment
	// the only valid next input is WriteRune.
	OnNestedComment() Commentator
	// a value has started decoding; even a nil value should trigger this.
	// any inline comments will live to the right of the value on the same line
	// there can only be one inline comment, continued with optional nesting.
	OnScalarValue() Commentator
	// a signature or dash in a collection has finished decoding.
	// paragraphs can get treated as key comments,
	// or headers for sub collection elements
	// depending on the number of paragraphs and the following value.
	OnKeyDecoded() Commentator
	// done with the current collection
	// usually followed by "GetComments"
	OnCollectionEnded() Commentator
}

// receive text from the decoded comments of a tell document.
// its signature mirrors strings.StringBuilder.
type RuneWriter interface {
	// each comment should start with a hash and space, and should end with a newline.
	// newlines outside of a comment can sometimes alter the meaning of subsequent comments
	// but are otherwise eaten. other runes should generate an error.
	WriteRune(rune) (int, error)
}

type CommentResolver interface {
	Commentator
	GetComments() string
}
