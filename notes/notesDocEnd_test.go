package notes

import (
	"strings"
	"testing"
)

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
	var str strings.Builder
	ctx := newContext(&str)
	b := makeRunecast(docEnd(ctx))
	//
	WriteLine(b.Inplace(), "one")
	WriteLine(b.Inplace(), "")
	WriteLine(b.Inplace(), "two")
	if got := str.String(); got != expected {
		t.Fatalf("got %q expected %q", got, expected)
	}
}
