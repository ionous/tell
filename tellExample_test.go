package tell_test

import (
	"fmt"
	"strings"

	"github.com/ionous/tell"
	"github.com/ionous/tell/maps/imap"
)

func ExampleString() {
	str := `true` // some tell document
	// tell/maps/imap contains a slice based ordered map implementation.
	// tell/maps/stdmap generates standard (unordered) go maps.
	// tell/maps/orderedmap uses Ian Coleman's ordered map implementation.
	doc := tell.NewDocument(imap.Builder, tell.KeepComments)
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
