package notes

import (
	"strings"
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
		var buf strings.Builder
		if e := charm.Parse(test, readTrailing(&buf, true)); e != nil {
			t.Fatal(e)
		} else if got := buf.String(); got != expect {
			t.Logf("test %d: \nwant %q \nhave %q", i, expect, got)
			t.Fail()
		}
	}
}
