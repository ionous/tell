package notes

import (
	"slices"
	"strconv"
	"testing"

	"github.com/ionous/tell/runes"
)

// test a sequence with two key comments and a scalar value
//
// - # key
// ..# more key
// .."value"
//
func TestCollection(t *testing.T) {
	const expected = "\r# key\n# more key"
	ctx := newContext()
	n := newCollection(ctx, doNothing)
	b := build(n)
	// we just created the collection above, so write the key comment:
	// - # key
	WriteLine(b.Inplace(), "key")
	// ..# more key
	WriteLine(b.Inplace(), "more key")
	// .."value"
	b.OnScalarValue()
	//
	ctx.flush(runes.Newline) // hrm
	got := ctx.out.String()
	if got != expected {
		t.Logf("\nwant %q \nhave %q", expected, got)
		t.Fatal("mismatch")
	}

}

// when there's a subcollection, the key should split
// between the parent container and the header of the first element.
// - # key
// ..# buffered header
// ....- "subcollection"
func TestKeyHeaderSplit(t *testing.T) {
	var expected = []string{
		"",
		"\r# key",           // the sequence has key
		"# buffered header", // the sub sequence has a header
	}
	ctx := newContext()
	n := newCollection(ctx, doNothing)
	b := build(n)

	// we just created the collection above, so write the key comment:
	// - # key
	WriteLine(b.Inplace(), "key")
	// ..# buffered header
	WriteLine(b.Inplace(), "buffered header")
	// ....- "subcollection"
	b.OnKeyDecoded().OnScalarValue()
	//
	if got := ctx.GetAllComments(); slices.Compare(got, expected) != 0 {
		for i, el := range got {
			t.Logf("%d %q", i, el)
		}
		t.Fatal("mismatch")
	}
}

// when there's a scalar value, the key should stick
// with the parent container ( there is no sub collection )
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
	ctx := newContext()
	n := newCollection(ctx, doNothing)
	b := build(n)

	// documents only have one value, in this case a sequence
	// - # key
	WriteLine(b.Inplace(), "key")
	// ..# buffered key
	WriteLine(b.Inplace(), "buffered key")
	// ..# more key
	WriteLine(b.Inplace(), "more key")
	// ..- "scalar" # inline
	WriteLine(b.OnScalarValue(), "inline")
	//
	// egad... 1 "\r# key\n# buffered key\r# inline# more key"
	// it must be buffering and not flushing when it sees the value
	//
	got := ctx.GetAllComments()
	if slices.Compare(got, expected) != 0 {
		for i, el := range got {
			t.Logf("%d %q", i, el)
		}
		t.Fatal("mismatch")
	}
}

// the document parser doesnt really hhandle this
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
	ctx := newContext()
	n := newCollection(ctx, doNothing)
	b := build(n)

	// documents only have one value, in this case a sequence
	// - # key & nesting
	WriteLine(b.Inplace(), "key")
	WriteLine(b.OnNestedComment(), "nested key")
	// ..# buffered key & nesting
	WriteLine(b.Inplace(), "second key")
	WriteLine(b.OnNestedComment(), "second nesting")
	// ..# buffered key & nesting
	WriteLine(b.Inplace(), "third key")
	WriteLine(b.OnNestedComment(), "third nesting")
	b.OnScalarValue()
	//
	got := ctx.GetAllComments()
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
	ctx := newContext()
	n := newCollection(ctx, doNothing)
	b := build(n)

	// documents only have one value, in this case a sequence
	// - # key & nesting
	WriteLine(b.Inplace(), "key")
	WriteLine(b.OnNestedComment(), "nested key")
	// ..# buffered key & nesting
	WriteLine(b.Inplace(), "second key")
	WriteLine(b.OnNestedComment(), "second nesting")
	// ..# buffered key & nesting
	WriteLine(b.Inplace(), "buffered header")
	WriteLine(b.OnNestedComment(), "nested header")
	//
	// ..- "subcollection scalar"
	b.OnKeyDecoded().OnScalarValue()
	got := ctx.GetAllComments()
	if slices.Compare(got, expected) != 0 {
		for i, el := range got {
			t.Logf("%d %q", i, el)
			t.Logf("x %q", expected[i])
		}
		t.Fatal("mismatch")
	}
}

// terms should have form feeds between each other
func TestEmptyTerms(t *testing.T) {
	const expected = "" +
		"\f\f\r\r# comment"
	ctx := newContext()
	n := newCollection(ctx, doNothing)
	b := build(n)
	// the builder started the collection
	// and the collection has an implicit first term
	// these are the two subsequent terms -- so two newlines
	for i := 0; i < 2; i++ {
		b.OnScalarValue().OnKeyDecoded()
	}
	WriteLine(b.OnScalarValue(), "comment")
	//
	if got := ctx.GetComments(); got != expected {
		t.Logf("\nwant %q \nhave %q", expected, got)
		t.Fail()
	}
}

// headers should appear right after their form feeds
// approximately:
// - 0
// # 1
// - 1
// # 2
// - 2
// ...
func xTestTermHeaders(t *testing.T) {
	const expected = "" +
		"# 0" +
		"\f# 1" +
		"\f# 2"
	ctx := newContext()
	n := newCollection(ctx, doNothing)
	b := build(n)
	//
	for i := 0; i < 3; i++ {
		if i > 0 {
			b.OnKeyDecoded()
		}
		// a scalar value followed by a newline:
		WriteLine(b.OnScalarValue(), "")
		// on the next line: a comment
		WriteLine(b.Inplace(), strconv.Itoa(i+1))
	}
	// fix: have to determine how to end correctly.
	// b.send(runePopped)
	//
	if got := ctx.GetComments(); got != expected {
		t.Logf("\nwant %q \nhave %q", expected, got)
		t.Fail()
	}
}