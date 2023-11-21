package notes_test

import (
	"slices"
	"strconv"
	"testing"

	"github.com/ionous/tell/notes"
	"github.com/ionous/tell/runes"
)

func TestEmptyish(t *testing.T) {
	const expected = "" +
		"# emptyish"
	b := notes.KeepComments()
	WriteLine(b.OnParagraph(), "emptyish")
	if got := b.GetComments(); got != expected {
		t.Fatalf("got %q expected %q", got, expected)
	}
}

// # header
// # subheader
// "value" # inline
// # footer
func TestDocumentComment(t *testing.T) {
	const expected = "" +
		"# header\n# subheader\r\r# inline\n# footer"
	b := notes.KeepComments()
	WriteLine(b.OnParagraph(), "header")
	WriteLine(b.OnParagraph(), "subheader")
	WriteLine(b.OnScalarValue(), "inline")
	WriteLine(b.OnFootnote(), "footer")
	//
	if got := b.GetComments(); got != expected {
		t.Logf("\nwant %q \nhave %q", expected, got)
		t.Fail()
	}
}

// terms should have form feeds between each other
func TestCommentEmptyTerms(t *testing.T) {
	const expected = "" +
		"\f\f\r\r# comment"
	b := notes.KeepComments()
	// the builder started the collection
	// and the collection has an implicit first term
	// these are the two subsequent terms -- so two newlines
	for i := 0; i < 2; i++ {
		b.OnScalarValue()
	}
	WriteLine(b.OnScalarValue(), "comment")
	if got := b.GetComments(); got != expected {
		t.Logf("\nwant %q \nhave %q", expected, got)
		t.Fail()
	}
}

// headers should appear right after their form feeds
// approximately:
// # 0
// 0
// ...
func TestTermHeaders(t *testing.T) {
	const expected = "" +
		"# 0" +
		"\f# 1" +
		"\f# 2"
	b := notes.KeepComments()
	for i := 0; i < 3; i++ {
		WriteLine(b.OnParagraph(), strconv.Itoa(i))
		b.OnScalarValue()
	}
	if got := b.GetComments(); got != expected {
		t.Logf("\nwant %q \nhave %q", expected, got)
		t.Fail()
	}
}

// the trailing header should be queued and sent to the parent.
// # header
// - "sequence"
// # footer
func TestDocumentFooter(t *testing.T) {
	var expected = []string{
		"# header\r\r\n# footer",
		"",
	}
	b := notes.KeepComments()
	WriteLine(b.OnParagraph(), "header")
	b.OnBeginCollection().OnKeyDecoded().OnScalarValue()
	// re: "header"
	// footer of the sequence would be indented
	// the parser can only assume the
	WriteLine(b.OnParagraph(), "footer")
	//
	if got := b.GetAllComments(); slices.Compare(got, expected) != 0 {
		for i, el := range got {
			t.Logf("%d %q", i, el)
		}
		t.Fatal("mismatch")
	}
}

// test an example similar to the one in the read me
func TestReadmeExample(t *testing.T) {
	const expected = "# header\r\r# inline\n# footer\n# second footer"
	b := notes.KeepComments()
	WriteLine(b.OnParagraph(), "header")
	WriteLine(b.OnScalarValue(), "inline")
	WriteLine(b.OnFootnote(), "footer")
	WriteLine(b.OnFootnote(), "second footer")
	if got := b.GetComments(); got != expected {
		t.Logf("\nwant %q \nhave %q", expected, got)
		t.Fail()
	}
}

func TestPanicNestedFooter(t *testing.T) {
	b := notes.KeepComments()
	WriteLine(b.OnScalarValue().OnFootnote(), "footer")
	func() {
		defer func() { _ = recover() }()
		WriteLine(b, "nesting not allowed")
		t.Fail()
	}()
}

// the comments mention a certain pattern
// make sure to build that pattern correctly
func TestCommentBlock(t *testing.T) {
	const expected = "" +
		"# header\n\t# nested header" +
		"\r# key\n\t# nested key" +
		"\r# inline\n\t# nested inline" +
		"\n# footer\n# extra footer"
	b := notes.KeepComments()
	WriteLine(b.OnParagraph(), "header")
	WriteLine(b, "nested header")
	WriteLine(b.OnKeyDecoded(), "key")
	WriteLine(b, "nested key")
	WriteLine(b.OnScalarValue(), "inline")
	WriteLine(b, "nested inline")
	WriteLine(b.OnFootnote(), "footer")
	WriteLine(b.OnFootnote(), "extra footer")
	got := b.GetComments()
	if got != expected {
		t.Logf("\nwant %q \nhave %q", expected, got)
		t.Fail()
	}
}

// when there's a subcollection, the key should split
// between the parent container and the header of the first element.
// - # key
// ..# buffered header
// ....- "subcollection"
func TestKeyHeaderSplit(t *testing.T) {
	var expected = []string{
		"",                  // the document has no comments
		"\r# key",           // the sequence has key
		"# buffered header", // the sub sequence has a header
	}
	b := notes.KeepComments()
	// documents only have one value, in this case a sequence
	// - # key
	WriteLine(b.OnBeginCollection().OnKeyDecoded(), "key")
	// ..# buffered header
	WriteLine(b.OnParagraph(), "buffered header")
	// ....- "subcollection"
	b.OnBeginCollection().OnKeyDecoded().OnScalarValue()
	if got := b.GetAllComments(); slices.Compare(got, expected) != 0 {
		for i, el := range got {
			t.Logf("%d %q", i, el)
		}
		t.Fatal("mismatch")
	}
}

// when there's a scalar value, the key should stick
// with the parent container because there is no first element
// - # key
// ..# buffered key
// ..# more key
// .."scalar" # inline
func TestKeyHeaderJoin(t *testing.T) {
	var expected = []string{
		// 0. the document has no comments
		"",
		// 1. the sequence has key
		"\r# key" +
			"\n# buffered key" +
			"\n# more key" +
			"\r# inline",
	}
	b := notes.KeepComments()
	// documents only have one value, in this case a sequence
	// - # key
	WriteLine(b.OnBeginCollection().OnKeyDecoded(), "key")
	// ..# buffered key
	WriteLine(b.OnParagraph(), "buffered key")
	// ..# more key
	WriteLine(b.OnParagraph(), "more key")
	// ..- "scalar" # inline
	WriteLine(b.OnScalarValue(), "inline")
	got := b.GetAllComments()
	if slices.Compare(got, expected) != 0 {
		for i, el := range got {
			t.Logf("%d %q", i, el)
		}
		t.Fatal("mismatch")
	}
}

// the document parser doesnt handle this
// but the comment builder can....
// - # key
// ....# nested key
// ..# second key
// ....# second nesting
// ..# third key
// ....# third nesting
// .."scalar"
func TestKeyNest(t *testing.T) {
	var expected = []string{
		// 0. the document has no comments
		"",
		// 1. the sequence has key
		"\r# key" +
			"\n\t# nested key" +
			"\n# second key" +
			"\n\t# second nesting" +
			"\n# third key" +
			"\n\t# third nesting",
	}
	b := notes.KeepComments()
	// documents only have one value, in this case a sequence
	// - # key & nesting
	WriteLine(b.OnBeginCollection().OnKeyDecoded(), "key")
	WriteLine(b, "nested key")
	// ..# buffered key & nesting
	WriteLine(b.OnParagraph(), "second key")
	WriteLine(b, "second nesting")
	// ..# buffered key & nesting
	WriteLine(b.OnParagraph(), "third key")
	WriteLine(b, "third nesting")
	b.OnScalarValue()
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

// the nested sequence version.
// - # key
// ....# nested key
// ..# second key
// ....# second nesting
// ..# buffered header
// ....# nested header
// ..- "subcollection scalar"
func TestKeyNestCollection(t *testing.T) {
	var expected = []string{
		// 0. the document has no comments
		"",
		// 1. the sequence has key
		"\r# key" +
			"\n\t# nested key" +
			"\n# second key" +
			"\n\t# second nesting",
		// 2.
		"# buffered header" +
			"\n\t# nested header",
	}
	b := notes.KeepComments()
	// documents only have one value, in this case a sequence
	// - # key & nesting
	WriteLine(b.OnBeginCollection().OnKeyDecoded(), "key")
	WriteLine(b, "nested key")
	// ..# buffered key & nesting
	WriteLine(b.OnParagraph(), "second key")
	WriteLine(b, "second nesting")
	// ..# buffered key & nesting
	WriteLine(b.OnParagraph(), "buffered header")
	WriteLine(b, "nested header")
	//
	// ..- "subcollection scalar"
	b.OnBeginCollection().OnKeyDecoded().OnScalarValue()
	got := b.GetAllComments()
	if slices.Compare(got, expected) != 0 {
		for i, el := range got {
			t.Logf("%d %q", i, el)
			t.Logf("x %q", expected[i])
		}
		t.Fatal("mismatch")
	}
}

// for testing: write the whole string and a newline
func WriteLine(w notes.RuneWriter, str string) {
	w.WriteRune(runes.Hash)
	w.WriteRune(runes.Space)
	for _, r := range str {
		w.WriteRune(r)
	}
	w.WriteRune(runes.Newline)
}