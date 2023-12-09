package notes

import (
	"slices"
	"testing"
)

// from inlineComment.tell
// ( fixed: misordering of \f and \r )
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

// from subComments1.tell
// ( fixed: redefined left-alignment now means key comments,
// and key comments are headers for the sub collection's first element )
//
// - # one
// ..# two
// ..# three
// ..- "sequence"
func TestSubComments(t *testing.T) {
	var expected = []string{
		"",                      // doc
		"",                      // outer sequence
		"# one\n# two\n# three", // inner sequence
	}
	var stack stringStack
	b := newNotes(stack.new())

	b.BeginCollection(stack.new())
	WriteLine(b, "one")
	WriteLine(b, "two")
	WriteLine(b, "three")
	b.BeginCollection(stack.new()).
		OnScalarValue().
		OnCollectionEnded().
		OnCollectionEnded()
	//
	if got := stack.Strings(); slices.Compare(got, expected) != 0 {
		for i, el := range got {
			t.Logf("%d %q", i, el)
			// t.Logf("x %q", expected[i])
		}
		t.Fatal("mismatch")
	}
}

// from entryExampleComments.tell
// ( fixed: extra newline b/t two and three )
//
// - # one
// ....# two
// ....# three
// .."value"
func TestNestedKeyComment(t *testing.T) {
	var expected = []string{
		"",
		"\r# one\n\t# two\n\t# three",
	}
	var stack stringStack
	b := newNotes(stack.new())

	b.BeginCollection(stack.new())
	WriteLine(b, "one")
	WriteLine(b.OnNestedComment(), "two")
	WriteLine(b.OnNestedComment(), "three")
	b.OnScalarValue().
		OnCollectionEnded()
	//
	if got := stack.Strings(); slices.Compare(got, expected) != 0 {
		for i, el := range got {
			t.Logf("%d %q", i, el)
			// t.Logf("x %q", expected[i])
		}
		t.Fatal("mismatch")
	}
}
