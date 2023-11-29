package notes

import (
	"slices"
	"strings"
	"testing"
)

// a simple one line header:
//
// # emptyish
//
func TestDocEmptyish(t *testing.T) {
	const expected = "" +
		"# emptyish"
	var str strings.Builder
	b := newNotes(&str)
	WriteLine(b.Inplace(), "emptyish")
	if got := str.String(); got != expected {
		t.Fatalf("got %q expected %q", got, expected)
	}
}

// header paragraphs should be separated by newlines
//
// # header
// # subheader
//
func TestDocHeaderLines(t *testing.T) {
	const expected = "" +
		"# header\n# subheader"
	var str strings.Builder
	b := newNotes(&str)
	WriteLine(b.Inplace(), "header")
	WriteLine(b.Inplace(), "subheader")
	//
	if got := str.String(); got != expected {
		t.Logf("\nwant %q \nhave %q", expected, got)
		t.Fail()
	}
}

// header paragraphs should allow nesting:
//
// # header
//   # nest
// # subheader
//   # nest
//
func TestDocHeaderNest(t *testing.T) {
	const expected = "" +
		"# header\n\t# nest\n# subheader\n\t# nest"

	var str strings.Builder
	b := newNotes(&str)
	WriteLine(b.Inplace(), "header")
	WriteLine(b.OnNestedComment(), "nest")
	WriteLine(b.Inplace(), "subheader")
	WriteLine(b.OnNestedComment(), "nest")
	b.OnCollectionEnded() // flush the document
	//
	if got := str.String(); got != expected {
		t.Logf("\nwant %q \nhave %q", expected, got)
		t.Fail()
	}
}

// newlines should split doc header into element headers
//
// # one
//
// # two
// - "collection"
func TestDocHeaderSplit(t *testing.T) {
	var expected = []string{
		"# one", // 0. the document has a header
		"# two", // 1. the sequence has a header
	}
	//
	var str stringStack
	b := newNotes(str.new())
	WriteLine(b.Inplace(), "one")
	WriteLine(b.Inplace(), "")
	WriteLine(b.Inplace(), "two")
	b.BeginCollection(str.new()).OnScalarValue()
	//
	got := str.Strings()
	if slices.Compare(got, expected) != 0 {
		for i, el := range got {
			t.Logf("%d %q", i, el)
		}
		t.Fatal("mismatch")
	}
}

// nesting should split doc header into element headers
//
// # one
//  # nest
// # two
// - "collection"
func TestDocHeaderSplitNest(t *testing.T) {
	var expected = []string{
		"# one\n\t# nest", // 0. the document has a header
		"# two",           // 1. the sequence has a header
	}
	//
	var str stringStack
	b := newNotes(str.new())
	WriteLine(b.Inplace(), "one")
	WriteLine(b.OnNestedComment(), "nest")
	WriteLine(b.Inplace(), "two")
	b.BeginCollection(str.new())
	//
	got := str.Strings()
	if slices.Compare(got, expected) != 0 {
		for i, el := range got {
			t.Logf("%d %q", i, el)
		}
		t.Fatal("mismatch")
	}
}

// test a document with a scalar and footer.
//
// # header
// # subheader
// "value" # inline
// # footer
func TestDocScalar(t *testing.T) {
	const expected = "" +
		"# header\n# subheader\r# inline\f# footer"

	var str strings.Builder
	b := newNotes(&str)
	WriteLine(b.Inplace(), "header")
	WriteLine(b.Inplace(), "subheader")
	WriteLine(b.OnScalarValue(), "inline")
	WriteLine(b.Inplace(), "footer")
	//
	if got := str.String(); got != expected {
		t.Logf("\nwant %q \nhave %q", expected, got)
		t.Fail()
	}
}

// test a document with a collection and footer.
//
// # header
// - "sequence"
// # footer
func TestDocCollection(t *testing.T) {
	const expected = "" +
		"# header\n# subheader\f# footer"

	var str stringStack
	b := newNotes(str.new())
	WriteLine(b.Inplace(), "header")
	WriteLine(b.Inplace(), "subheader")
	WriteLine(b.BeginCollection(str.new()).OnScalarValue(), "")
	WriteLine(b.Inplace(), "footer")
	b.OnEof() // flush the document

	//
	if got := str.Strings()[0]; got != expected {
		t.Logf("\nwant %q \nhave %q", expected, got)
		t.Fail()
	}
}

// edge case: when there's no trailing newline
// and a nil.
// - # key
// ..# more key<eof>
func TestKeyNil(t *testing.T) {
	var expected = "\r# key\n# more key"

	var str stringStack
	b := newNotes(str.new())
	// documents only have one value, in this case a sequence
	// - # key
	WriteLine(b.BeginCollection(str.new()), "key")
	// ..# more key ( but no eol )
	for _, q := range "# more key" {
		b.WriteRune(q)
	}
	b.OnEof() // flush the document
	//
	if got := str.Strings()[1]; got != expected {
		t.Logf("\nwant %q \nhave %q", expected, got)
		t.Fail()
	}
}
