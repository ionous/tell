package notes

import (
	"slices"
	"testing"
)

// from inlineComment.tell
//
// - 5  # one inline comment
// - 11 # and more
//
func TestInlineCollection(t *testing.T) {
	var expected = []string{
		"",
		"\r\r# one inline comment\f\r\r# and more",
	}
	var stack stringStack
	b := newNotes(stack.new())

	WriteLine(b.BeginCollection(stack.new()).OnScalarValue(), "one inline comment")
	WriteLine(b.OnKeyDecoded().OnScalarValue(), "and more")
	//
	if got := stack.Strings(); slices.Compare(got, expected) != 0 {
		for i, el := range got {
			t.Logf("%d %q", i, el)
			t.Logf("x %q", expected[i])
		}
		t.Fatal("mismatch")
	}
}