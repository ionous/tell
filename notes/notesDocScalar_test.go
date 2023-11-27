package notes

import (
	"testing"
)

// minimalist testing of document scalar comments
func TestDocScalar(t *testing.T) {
	const expected = "" +
		"\r# one\n\t# two\n\t# three"

	ctx := newContext()
	b := build(docScalar(ctx, doNothing))
	//
	WriteLine(b.Inplace(), "one")
	WriteLine(b.OnNestedComment(), "two")
	WriteLine(b.OnNestedComment(), "three")
	if got := ctx.GetComments(); got != expected {
		t.Fatalf("got %q expected %q", got, expected)
	}
}
