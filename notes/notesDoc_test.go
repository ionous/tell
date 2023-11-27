package notes

import (
	"testing"

	"github.com/ionous/tell/charm"
)

// test a document with a scalar and footer.
//
// # header
// # subheader
// "value" # inline
// # footer
func TestDocumentComment(t *testing.T) {
	const expected = "" +
		"# header\n# subheader\r# inline\f# footer"

	ctx := newContext()
	b := build(docStart(ctx, func() charm.State {
		return docScalar(ctx, func() charm.State {
			return docEnd(ctx)
		})
	}, doNothing))
	//

	WriteLine(b.Inplace(), "header")
	WriteLine(b.Inplace(), "subheader")
	WriteLine(b.OnScalarValue(), "inline")
	WriteLine(b.Inplace(), "footer")
	//
	if got := ctx.GetComments(); got != expected {
		t.Logf("\nwant %q \nhave %q", expected, got)
		t.Fail()
	}
}
