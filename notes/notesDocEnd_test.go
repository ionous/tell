package notes

import "testing"

// use the docEnd decoder to read
// two comment lines, split by a blank line
//
// # one
//
// # two
//
func TestDocFooter(t *testing.T) {
	const expected = "" +
		"\f# one\n# two"

	// uses just the end parser
	ctx := newContext()
	b := build(docEnd(ctx))
	//
	WriteLine(b.Inplace(), "one")
	WriteLine(b.Inplace(), "")
	WriteLine(b.Inplace(), "two")
	if got := b.GetAllComments(ctx)[0]; got != expected {
		t.Fatalf("got %q expected %q", got, expected)
	}
}
