package notes

import (
	"testing"

	"github.com/ionous/tell/charm"
	"github.com/ionous/tell/runes"
)

func doNothing() charm.State {
	return nil
}

// a simple one line header:
//
// # emptyish
//
func TestEmptyish(t *testing.T) {
	const expected = "" +
		"# emptyish"

	ctx := newContext()
	b := build(docStart(ctx, doNothing, doNothing))
	//
	WriteLine(b.OnParagraph(), "emptyish")
	if got := ctx.GetComments(); got != expected {
		t.Fatalf("got %q expected %q", got, expected)
	}
}

// header paragraphs should be separated by newlines
//
// # header
// # subheader
//
func TestHeaderLines(t *testing.T) {
	const expected = "" +
		"# header\n# subheader"

	ctx := newContext()
	b := build(docStart(ctx, doNothing, doNothing))
	WriteLine(b.OnParagraph(), "header")
	WriteLine(b.OnParagraph(), "subheader")
	//
	if got := ctx.GetComments(); got != expected {
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
func TestHeaderNest(t *testing.T) {
	const expected = "" +
		"# header\n\t# nest\n# subheader\n\t# nest"

	ctx := newContext()
	b := build(docStart(ctx, doNothing, doNothing))
	WriteLine(b.OnParagraph(), "header")
	WriteLine(&b, "nest")
	WriteLine(b.OnParagraph(), "subheader")
	WriteLine(&b, "nest")
	//
	if got := ctx.GetComments(); got != expected {
		t.Logf("\nwant %q \nhave %q", expected, got)
		t.Fail()
	}
}

func WriteBreak(w RuneWriter) {
	w.WriteRune(runes.Newline)
}

// for testing: write the whole string and a newline
func WriteLine(w RuneWriter, str string) {
	WriteInline(w, str)
	w.WriteRune(runes.Newline)
}

// for testing: write the whole string and a newline
func WriteInline(w RuneWriter, str string) {
	w.WriteRune(runes.Hash)
	w.WriteRune(runes.Space)
	for _, r := range str {
		w.WriteRune(r)
	}
}
