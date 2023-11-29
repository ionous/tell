package notes

import (
	"slices"
	"strconv"
	"strings"
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
	var str strings.Builder
	ctx := newContext(&str)
	b := newCommentBuilder(ctx, newCollection(ctx))
	// we just created the collection above, so write the key comment:
	// - # key
	WriteLine(b.Inplace(), "key")
	// ..# more key
	WriteLine(b.Inplace(), "more key")
	// .."value"
	b.OnScalarValue()
	//
	ctx.flush(runes.Newline) // hrm
	got := str.String()
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
		"\r# key",           // the sequence has key
		"# buffered header", // the sub sequence has a header
	}
	var stack stringStack
	ctx := newContext(stack.new())
	b := newCommentBuilder(ctx, newCollection(ctx))

	// we just created the collection above, so write the key comment:
	// - # key
	WriteLine(b.Inplace(), "key")
	// ..# buffered header
	WriteLine(b.Inplace(), "buffered header")
	// ....- "subcollection"
	b.BeginCollection(stack.new()).OnScalarValue()
	//
	if got := stack.Strings(); slices.Compare(got, expected) != 0 {
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
		// 1. the sequence has key
		"\r# key" +
			"\n# buffered key" +
			"\n# more key" +
			"\r# inline",
	}
	var stack stringStack
	ctx := newContext(stack.new())
	b := newCommentBuilder(ctx, newCollection(ctx))

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
	got := stack.Strings()
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
		// 1. the sequence has key
		"\r# key" +
			"\n\t# nested key" +
			"\n# second key" +
			"\n\t# second nesting" +
			"\n# third key" +
			"\n\t# third nesting",
	}
	var stack stringStack
	ctx := newContext(stack.new())
	b := newCommentBuilder(ctx, newCollection(ctx))

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
	got := stack.Strings()
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
		// 1. the sequence has key
		"\r# key" +
			"\n\t# nested key" +
			"\n# second key" +
			"\n\t# second nesting",
		// 2.
		"# buffered header" +
			"\n\t# nested header",
	}

	var stack stringStack
	ctx := newContext(stack.new())
	b := newCommentBuilder(ctx, newCollection(ctx))

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
	b.BeginCollection(stack.new()).OnScalarValue()
	got := stack.Strings()
	if slices.Compare(got, expected) != 0 {
		for i, el := range got {
			t.Logf("%d %q", i, el)
			t.Logf("x %q", expected[i])
		}
		t.Fatal("mismatch")
	}
}

// terms should have form feeds between each other
//
// - 1
// - 2
// - 3 # comment
//
func TestEmptyTerms(t *testing.T) {
	const expected = "" +
		"\f\f\r\r# comment"

	var str strings.Builder
	ctx := newContext(&str)
	b := newCommentBuilder(ctx, newCollection(ctx))
	// the builder started the collection
	// and the collection has an implicit first term
	// these are the two subsequent terms -- so two newlines
	for i := 0; i < 2; i++ {
		b.OnScalarValue().OnKeyDecoded()
	}
	WriteLine(b.OnScalarValue(), "comment")
	//
	if got := str.String(); got != expected {
		t.Logf("\nwant %q \nhave %q", expected, got)
		t.Fail()
	}
}

// headers should appear right after their form feeds:
// - 0
// # 1
// - 1
// # 2
// - 2
//
func TestTermHeaders(t *testing.T) {
	const expected = "" +
		"\f# 1" +
		"\f# 2"

	var str strings.Builder
	ctx := newContext(&str)
	b := newCommentBuilder(ctx, newCollection(ctx))
	//
	for i := 0; i < 3; i++ {
		// the zeroth key exists because of newCollection
		// for all subsequent entries: write a header.
		if i > 0 {
			WriteLine(b.Inplace(), strconv.Itoa(i))
			b.OnKeyDecoded()
		}
		// a scalar value followed by a newline:
		WriteLine(b.OnScalarValue(), "")
	}
	if got := str.String(); got != expected {
		t.Logf("\nwant %q \nhave %q", expected, got)
		t.Fail()
	}
}

// handle sequences that begin and end
//
// - 1
// - - 2
//   - 3
// - - 4
//   - - 5
// - 6
//
// ie. [1,[2,3],[4,[5]],6]
func TestCollectBeginEnd(t *testing.T) {
	var stack stringStack
	ctx := newContext(stack.new())
	b := newCommentBuilder(ctx, newCollection(ctx))
	// in order of left bracket
	// everything here appears as inline comments (\r\r)
	// each comma is a formfeed, with trailing sub-collections are trimmed.
	expected := []string{
		"\r\r# 1\f\f\f\r\r# 6", // the outer most array ends last
		"\r\r# 2\f\r\r# 3",
		"\r\r# 4",
		"\r\r# 5", // the array closest to 5 ends before, 4...5
		// [1,*,*,6] 3 comma separators, 3 form feeds.
	}

	// no initial key because "newCollection" was our key
	WriteLine(b.OnScalarValue(), "1")
	WriteLine(b.OnKeyDecoded().BeginCollection(stack.new()).OnScalarValue(), "2")
	WriteLine(b.OnKeyDecoded().OnScalarValue(), "3")
	b.OnCollectionEnded()
	WriteLine(b.OnKeyDecoded().BeginCollection(stack.new()).OnScalarValue(), "4")
	WriteLine(b.OnKeyDecoded().BeginCollection(stack.new()).OnScalarValue(), "5")
	b.OnCollectionEnded().OnCollectionEnded()
	WriteLine(b.OnKeyDecoded().OnScalarValue(), "6")
	//
	got := stack.Strings()
	if slices.Compare(got, expected) != 0 {
		for i, el := range got {
			t.Logf("%d %q", i, el)
		}
		t.Fatal("mismatch")
	}
}

type stringStack []*strings.Builder

func (f *stringStack) new() *strings.Builder {
	next := new(strings.Builder)
	(*f) = append(*f, next)
	return next
}

func (f *stringStack) Strings() []string {
	out := make([]string, len(*f))
	for i, buf := range *f {
		out[i] = buf.String()
	}
	return out
}
