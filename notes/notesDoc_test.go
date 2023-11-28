package notes

import (
	"testing"
)

// a simple one line header:
//
// # emptyish
//
func TestDocEmptyish(t *testing.T) {
	const expected = "" +
		"# emptyish"

	ctx := newContext()
	b := build(newDocument(ctx))
	//
	WriteLine(b.Inplace(), "emptyish")
	if got := b.GetComments(ctx)[0]; got != expected {
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

	ctx := newContext()
	b := build(newDocument(ctx))
	WriteLine(b.Inplace(), "header")
	WriteLine(b.Inplace(), "subheader")
	//
	if got := b.GetComments(ctx)[0]; got != expected {
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

	ctx := newContext()
	b := build(newDocument(ctx))
	WriteLine(b.Inplace(), "header")
	WriteLine(b.OnNestedComment(), "nest")
	WriteLine(b.Inplace(), "subheader")
	WriteLine(b.OnNestedComment(), "nest")
	//
	if got := b.GetComments(ctx)[0]; got != expected {
		t.Logf("\nwant %q \nhave %q", expected, got)
		t.Fail()
	}
}

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

	ctx := newContext()
	b := build(docEnd(ctx))
	//
	WriteLine(b.Inplace(), "one")
	WriteLine(b.Inplace(), "")
	WriteLine(b.Inplace(), "two")
	if got := b.GetComments(ctx)[0]; got != expected {
		t.Fatalf("got %q expected %q", got, expected)
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

	ctx := newContext()
	b := build(newDocument(ctx))
	//
	WriteLine(b.Inplace(), "header")
	WriteLine(b.Inplace(), "subheader")
	WriteLine(b.OnScalarValue(), "inline")
	WriteLine(b.Inplace(), "footer")
	//
	if got := b.GetComments(ctx)[0]; got != expected {
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

	ctx := newContext()
	b := build(newDocument(ctx))
	//
	WriteLine(b.Inplace(), "header")
	WriteLine(b.Inplace(), "subheader")
	WriteLine(b.OnKeyDecoded().OnScalarValue(), "")
	WriteLine(b.Inplace(), "footer")
	//
	if got := b.GetComments(ctx)[0]; got != expected {
		t.Logf("\nwant %q \nhave %q", expected, got)
		t.Fail()
	}
}
