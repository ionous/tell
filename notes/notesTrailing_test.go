package notes

import (
	"testing"

	"github.com/ionous/tell/charm"
)

// after a collection's scalar value:
// there might be an inline comment or a trailing comment
// ( whichever the case, all comments are aligned )
//
// "value" # inline
//         # aligned
//
// "value"
//   # block
//   # also aligned
//
func TestTrailingInline(t *testing.T) {
	tests := []string{
		"# inline\n\t# aligned",
		"\n\t# block\n\t# also aligned",
	}
	expected := []string{
		"\r# inline\n\t# aligned",
		"\r\n\t# block\n\t# also aligned",
	}

	for i, test := range tests {
		expect := expected[i]
		ctx := newContext()
		if e := charm.Parse(test, readTrailing(ctx, true)); e != nil {
			t.Fatal(e)
		} else if got := ctx.out.Resolve(); got != expect {
			t.Logf("test %d: \nwant %q \nhave %q", i, expect, got)
			t.Fail()
		}
	}
}

// minimalist testing of document scalar comments
func TestInlineOnly(t *testing.T) {
	const expected = "" +
		"\r# one\n\t# two\n\t# three"

	ctx := newContext()
	b := build(readInline(ctx))
	//
	WriteLine(b.Inplace(), "one")
	WriteLine(b.OnNestedComment(), "two")
	WriteLine(b.OnNestedComment(), "three")
	if got := b.GetComments(ctx)[0]; got != expected {
		t.Fatalf("got %q expected %q", got, expected)
	}
}
