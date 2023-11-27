package notes

import "testing"

// a simple footer test:
// two comment lines, split by a blank line
//
// # one
//
// # two
//
func TestFooter(t *testing.T) {
	const expected = "" +
		"\f# one\n# two"

	ctx := newContext()
	b := build(docEnd(ctx))
	//
	WriteLine(b.Inplace(), "one")
	WriteLine(b.Inplace(), "")
	WriteLine(b.Inplace(), "two")
	if got := ctx.GetComments(); got != expected {
		t.Fatalf("got %q expected %q", got, expected)
	}
}
