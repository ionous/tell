package notes

import "testing"

// a simple footer test:
//
// # emptyish
//
func TestFooter(t *testing.T) {
	const expected = "" +
		"\f# one\n# two"

	ctx := newContext()
	b := build(docEnd(ctx))
	//
	WriteLine(b.OnParagraph(), "one")
	WriteBreak(&b)
	WriteLine(b.OnParagraph(), "two")
	if got := ctx.GetComments(); got != expected {
		t.Fatalf("got %q expected %q", got, expected)
	}
}
