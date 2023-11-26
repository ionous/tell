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

	WriteLine(b.OnParagraph(), "header")
	WriteLine(b.OnParagraph(), "subheader")
	WriteLine(b.OnScalarValue(), "inline")
	WriteLine(b.OnParagraph(), "footer")
	//
	if got := ctx.GetComments(); got != expected {
		t.Logf("\nwant %q \nhave %q", expected, got)
		t.Fail()
	}
}
