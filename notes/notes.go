package notes

// comment block creation
// events indicate different sections of a tell document
// during decoding. those methods return "self" for chaining.
type Commentator interface {
	// a key for the start of a new collection has been decoded.
	// the rune writer must be valid until at least the corresponding OnCollectionEnded
	BeginCollection(RuneWriter) Commentator
	// an explicit request to nest the next comment
	// the only valid next input is WriteRune.
	OnNestedComment() Commentator
	// a value has started decoding; even a nil value should trigger this.
	// any inline comments will live to the right of the value on the same line
	// there can only be one inline comment, continued with optional nesting.
	OnScalarValue() Commentator
	// a signature or dash in an existing collection has finished decoding.
	// paragraphs can get treated as key comments,
	// or headers for sub collection elements
	// depending on the number of paragraphs and the following value.
	OnKeyDecoded() Commentator
	// done with the current collection
	// usually followed by "GetComments"
	OnCollectionEnded() Commentator
	// receive text from the decoded comments of a tell document.
	// each comment should start with a hash and space, and should end with a newline.
	// newlines outside of a comment can sometimes alter the meaning of subsequent comments
	// but are otherwise eaten.
	// the signature mirrors strings.StringBuilder, but always returns 0, nil
	WriteRune(rune) (int, error)
}

type RuneWriter interface {
	WriteRune(rune) (int, error)
}

type stringWriter interface {
	WriteString(string) (int, error)
}
