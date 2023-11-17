package decode_test

import (
	"fmt"
	"strings"

	"github.com/ionous/tell/decode"
	"github.com/ionous/tell/maps/imap"
)

func ExampleDocument() {
	str := `true` // some tell document
	// maps/imap contains a slice based ordered map implementation.
	// maps/stdmap generates standard (unordered) go maps.
	// maps/orderedmap uses Ian Coleman's ordered map implementation.
	doc := decode.NewDocument(imap.Builder, decode.KeepComments)
	// ReadDoc takes a string reader
	if res, e := doc.ReadDoc(strings.NewReader(str)); e != nil {
		panic(e)
	} else {
		// the results contains document level comments
		// and the content that was read.
		fmt.Println(res.Content)
	}
	// Output: true
}
