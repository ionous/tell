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
	b := NewBuilder()
	WriteLine(b.OnBeginHeader(), "emptyish")
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
	b := NewBuilder()
	WriteLine(b.OnBeginHeader(), "header")
	WriteLine(b.OnBeginHeader(), "subheader")
	// i think the key and value for document are nestled together.
	// WriteLine(b.OnKeyDecoded(), "invisible padding")
	WriteLine(b.OnScalarValue(), "inline")
	WriteLine(b.OnBeginFooter(), "footer")
	//
	if got := b.GetComments(); got != expected {
		t.Fatalf("got %q expected %q", got, expected)
	}
}

// terms should have form feeds between each other
func TestCommentEmptyTerms(t *testing.T) {
	const expected = "" +
		"\f\f\r\r# comment"
	b := NewBuilder()
	// the builder started the collection
	// and the collection has an implicit first term
	// these are the two subsequent terms -- so two newlines
	for i := 0; i < 2; i++ {
		b.OnScalarValue()
		b.OnTermDecoded()
	}
	WriteLine(b.OnScalarValue(), "comment")
	if got := b.GetComments(); got != expected {
		t.Fatalf("got %q expected %q", got, expected)
	}
}

// headers should appear right after their form feeds
func TestTermHeaders(t *testing.T) {
	const expected = "" +
		"# 0" +
		"\f# 1" +
		"\f# 2"
	b := NewBuilder()
	for i := 0; i < 3; i++ {
		WriteLine(b.OnBeginHeader(), strconv.Itoa(i))
		b.OnTermDecoded()
	}
	if got := b.GetComments(); got != expected {
		t.Fatalf("got %q expected %q", got, expected)
	}
}

// the trailing header should be queued and sent to the parent.
// # header
// - "sequence"
// # footer
func TestDocumentFooter(t *testing.T) {
	var expected = []string{
		"# header\r\r#footer",
		"",
	}
	b := NewBuilder()
	WriteLine(b.OnBeginHeader(), "header")
	b.OnBeginCollection().OnKeyDecoded().OnScalarValue().OnTermDecoded()
	WriteLine(b.OnBeginFooter(), "footer")
	//
	if got := b.GetAllComments(); slices.Compare(got, expected) != 0 {
		for i, el := range got {
			t.Logf("%d %q", i, el)
		}
		t.Fatal("mismatch")
	}
}

func TestReadmeExample(t *testing.T) {
	const expected = "# header\r\r# inline\n# footer"
	b := NewBuilder()
	WriteLine(b.OnBeginHeader(), "header")
	WriteLine(b.OnScalarValue(), "inline")
	WriteLine(b.OnBeginFooter(), "footer")
	if got := b.GetComments(); got != expected {
		t.Fatalf("got %q expected %q", got, expected)
	}
}

func TestPanics(t *testing.T) {
	b := NewBuilder()
	WriteLine(b.OnKeyDecoded(), "padding")
	func() {
		defer func() { _ = recover() }()
		WriteLine(b.OnKeyDecoded(), "multiple not allowed")
		t.Fail() // expects not to get to this line because of panicking
	}()
	WriteLine(b.OnScalarValue(), "inline")
	func() {
		defer func() { _ = recover() }()
		WriteLine(b.OnScalarValue(), "multiple not allowed")
		t.Fail()
	}()
	WriteLine(b.OnBeginFooter(), "footer")
	func() {
		defer func() { _ = recover() }()
		WriteLine(b, "nesting not allowed")
		t.Fail()
	}()
	func() {
		defer func() { _ = recover() }()
		WriteLine(b.OnBeginHeader(), "rewind not allowed")
		t.Fail()
	}()
}

func TestHeaderPanic(t *testing.T) {
	// check that we cant use the header after nesting.
	b := NewBuilder()
	WriteLine(b.OnBeginHeader(), "header")
	WriteLine(b, "nested")
	func() {
		defer func() { _ = recover() }()
		WriteLine(b.OnBeginHeader(), "header")
		t.Fatalf("should panic when extending the header after nesting")
	}()
	// check that we cant use nesting after extending the header
	b = NewBuilder()
	WriteLine(b.OnBeginHeader(), "header")
	WriteLine(b.OnBeginHeader(), "subheader")
	func() {
		defer func() { _ = recover() }()
		WriteLine(b, "nested")
		t.Fatalf("should panic when extending the header after nesting")
	}()
}

// the comments mention a certain pattern
// make sure to build that pattern correctly
func TestCommentBlock(t *testing.T) {
	const expected = "" +
		"# header\n\t# nested header" +
		"\r# padding\n\t# nested padding" +
		"\r# inline\n\t# nested inline" +
		"\n# footer\n# extra footer"
	b := NewBuilder()
	WriteLine(b.OnBeginHeader(), "header")
	WriteLine(b, "nested header")
	WriteLine(b.OnKeyDecoded(), "padding")
	WriteLine(b, "nested padding")
	WriteLine(b.OnScalarValue(), "inline")
	WriteLine(b, "nested inline")
	WriteLine(b.OnBeginFooter(), "footer")
	WriteLine(b.OnBeginFooter(), "extra footer")
	got := b.GetComments()
	if got != expected {
		t.Fatalf("got %q expected %q", got, expected)
	}
}

// when there's a subcollection, the padding should split
// between the parent container and the header of the first element.
// - # padding
// ..# buffered header
// ....- "subcollection"
func TestPaddingHeaderSplit(t *testing.T) {
	var expected = []string{
		"",                  // the document has no comments
		"\r# padding",       // the sequence has padding
		"# buffered header", // the sub sequence has a header
	}
	b := NewBuilder()
	// documents only have one value, in this case a sequence
	// - # padding
	WriteLine(b.OnBeginCollection().OnKeyDecoded(), "padding")
	// ..# buffered header
	WriteLine(b.OnBeginHeader(), "buffered header")
	// ....- "subcollection"
	b.OnBeginCollection().OnKeyDecoded().OnScalarValue()
	if got := b.GetAllComments(); slices.Compare(got, expected) != 0 {
		for i, el := range got {
			t.Logf("%d %q", i, el)
		}
		t.Fatal("mismatch")
	}
}

// when there's a scalar value, the padding should stick
// with the parent container because there is no first element
// - # padding
// ..# buffered padding
// ..# more padding
// .."scalar" # inline
func TestPaddingHeaderJoin(t *testing.T) {
	var expected = []string{
		// 0. the document has no comments
		"",
		// 1. the sequence has padding
		"\r# padding" +
			"\n# buffered padding" +
			"\n# more padding" +
			"\r# inline",
	}
	b := NewBuilder()
	// documents only have one value, in this case a sequence
	// - # padding
	WriteLine(b.OnBeginCollection().OnKeyDecoded(), "padding")
	// ..# buffered padding
	WriteLine(b.OnBeginHeader(), "buffered padding")
	// ..# more padding
	WriteLine(b.OnBeginHeader(), "more padding")
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
// - # padding
// ....# nested padding
// ..# buffered padding
// ....# nested buffer
// ..# buffered padding
// ....# nested buffer
func TestPaddingNest(t *testing.T) {
	var expected = []string{
		// 0. the document has no comments
		"",
		// 1. the sequence has padding
		"\r# padding" +
			"\n\t# nested padding" +
			"\n# buffered padding" + // plus buffered padding
			"\n\t# nested buffer" +
			"\n# buffered padding" + // plus buffered padding
			"\n\t# nested buffer",
	}
	b := NewBuilder()
	// documents only have one value, in this case a sequence
	// - # padding & nesting
	WriteLine(b.OnBeginCollection().OnKeyDecoded(), "padding")
	WriteLine(b, "nested padding")
	// ..# buffered padding & nesting
	WriteLine(b.OnBeginHeader(), "buffered padding")
	WriteLine(b, "nested buffer")
	// ..# buffered padding & nesting
	WriteLine(b.OnBeginHeader(), "buffered padding")
	WriteLine(b, "nested buffer")
	//
	got := b.GetAllComments()
	if slices.Compare(got, expected) != 0 {
		for i, el := range got {
			t.Logf("%d %q", i, el)
		}
		t.Fatal("mismatch")
	}
}

// the nested sequence version.
// - # padding
// ....# nested padding
// ..# buffered padding
// ....# nested buffer
// ..# buffered padding
// ....# nested buffer
// ....- "subcollection scalar"
func TestPaddingNestCollection(t *testing.T) {
	var expected = []string{
		// 0. the document has no comments
		"",
		// 1. the first term has padding:
		"\r# padding\n\t# nested padding",
		// 2. the sub sequence gets a header
		"# buffered heading\n\t# nested buffer" +
			//... a second header
			"\n# buffered padding\n\t# nested buffer",
	}
	b := NewBuilder()
	// documents only have one value, in this case a sequence
	// - # padding & nesting
	WriteLine(b.OnBeginCollection().OnKeyDecoded(), "padding")
	WriteLine(b, "nested padding")
	// ..# buffered padding & nesting
	WriteLine(b.OnBeginHeader(), "buffered heading")
	WriteLine(b, "nested buffer")
	// ..# buffered padding & nesting
	WriteLine(b.OnBeginHeader(), "buffered padding")
	WriteLine(b, "nested buffer")
	//
	// ....- "subcollection scalar"
	b.OnBeginCollection().OnKeyDecoded().OnScalarValue()
	got := b.GetAllComments()
	if slices.Compare(got, expected) != 0 {
		for i, el := range got {
			t.Logf("%d %q", i, el)
		}
		t.Fatal("mismatch")
	}
}

func NewBuilder() *notes.Builder {
	var b notes.Builder
	b.OnBeginCollection()
	return &b
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
