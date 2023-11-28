package notes

import (
	"slices"
	"testing"

	"github.com/ionous/tell/charm"
	"github.com/ionous/tell/runes"
)

func doNothing() charm.State {
	return nil
}

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
func TestReadmeExample(t *testing.T) {
	var expected = []string{
		"# header\f# footer",
		"\r\r# inline",
	}

	b := newNotes()
	WriteLine(b.Inplace(), "header")
	WriteLine(b.OnKeyDecoded().OnScalarValue(), "inline")
	WriteLine(b.Inplace(), "footer")
	//
	//
	got := b.GetAllComments()
	if slices.Compare(got, expected) != 0 {
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

	b := newNotes()
	WriteLine(b.Inplace(), "header")
	WriteLine(b.OnNestedComment(), "nested header")
	WriteLine(b.OnKeyDecoded(), "key")
	WriteLine(b.OnNestedComment(), "nested key")
	WriteLine(b.OnScalarValue(), "inline")
	WriteLine(b.OnNestedComment(), "nested inline")
	WriteLine(b.Inplace(), "footer")
	WriteLine(b.Inplace(), "extra footer")
	//
	got := b.GetAllComments()
	if slices.Compare(got, expected) != 0 {
		for i, el := range got {
			t.Logf("%d %q", i, el)
			t.Logf("x %q", expected[i])
		}
		t.Fatal("mismatch")
	}
}
