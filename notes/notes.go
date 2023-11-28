package notes

//
type Commentator interface {
	Events
	GetComments() string
}

// comment block creation
// events indicate different sections of a tell document
// during decoding. those methods return "self" for chaining.
type Events interface {
	// an explicit request to nest the next comment
	// the only valid next input is WriteRune.
	OnNestedComment() Events
	// a value has started decoding; even a nil value should trigger this.
	// any inline comments will live to the right of the value on the same line
	// there can only be one inline comment, continued with optional nesting.
	OnScalarValue() Events
	// a signature or dash in a collection has finished decoding.
	// paragraphs can get treated as key comments,
	// or headers for sub collection elements
	// depending on the number of paragraphs and the following value.
	OnKeyDecoded() Events
	// done with the current collection
	// usually followed by "GetComments"
	OnCollectionEnded() Events
	// receive text from the decoded comments of a tell document.
	// each comment should start with a hash and space, and should end with a newline.
	// newlines outside of a comment can sometimes alter the meaning of subsequent comments
	// but are otherwise eaten.
	// the signature mirrors strings.StringBuilder, but always returns 0, nil
	WriteRune(rune) (int, error)
}

type runeWriter interface {
	WriteRune(rune) (int, error)
}
