package notes

import (
	"slices"
	"testing"

	"github.com/ionous/tell/runes"
)

// for testing: write a comment and a newline
// to write a fully blank line, pass the empty string
func WriteLine(w RuneWriter, str string) {
	if len(str) > 0 {
		w.WriteRune(runes.Hash)
		w.WriteRune(runes.Space)
		for _, r := range str {
			w.WriteRune(r)
		}
	}
	w.WriteRune(runes.Newline)
}

// test an example similar to the one in the read me
//
// # header
// - "value" # inline
// # footer
//
func TestReadmeExample(t *testing.T) {
	var expected = []string{
		"# header\f# footer",
		"\r\r# inline",
	}
	var stack stringStack
	b := newNotes(stack.new())
	WriteLine(b, "header")
	WriteLine(b.BeginCollection(stack.new()).OnScalarValue(), "inline")
	WriteLine(b, "footer")
	b.OnEof()
	//
	if got := stack.Strings(); slices.Compare(got, expected) != 0 {
		for i, el := range got {
			t.Logf("%d %q", i, el)
			t.Logf("x %q", expected[i])
		}
		t.Fatal("mismatch")
	}
}

// the comments mention a certain pattern
// make sure to build that pattern correctly
func TestCommentBlock(t *testing.T) {
	var expected = []string{
		"# header\n\t# nested header" +
			"\f# footer\n# extra footer",
		"\r# key\n\t# nested key" +
			"\r# inline\n\t# nested inline",
	}

	var stack stringStack
	b := newNotes(stack.new())
	WriteLine(b, "header")
	WriteLine(b.OnNestedComment(), "nested header")
	WriteLine(b.BeginCollection(stack.new()), "key")
	WriteLine(b.OnNestedComment(), "nested key")
	WriteLine(b.OnScalarValue(), "inline")
	WriteLine(b.OnNestedComment(), "nested inline")
	WriteLine(b, "footer")
	WriteLine(b, "extra footer")
	b.OnEof()
	//
	if got := stack.Strings(); slices.Compare(got, expected) != 0 {
		for i, el := range got {
			t.Logf("%d %q", i, el)
			t.Logf("x %q", expected[i])
		}
		t.Fatal("mismatch")
	}
}
