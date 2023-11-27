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
		"# inline\n# aligned",
		"\n# block\n# also aligned",
	}
	expected := []string{
		"\r# inline\n\t# aligned",
		"\r\n\t# block\n\t# also aligned",
	}

	for i, test := range tests {
		expect := expected[i]
		ctx := newContext()
		if e := charm.Parse(test, readTrailing(ctx)); e != nil {
			t.Fatal(e)
		} else if got := ctx.buf.String(); got != expect {
			t.Logf("test %d: \nwant %q \nhave %q", i, expect, got)
			t.Fail()
		}
	}
}
