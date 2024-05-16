package charm

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

func TestRequires(t *testing.T) {
	isSpace := func(r rune) bool { return r == ' ' }

	// index of the fail point, or -1 if success is expected
	count := func(failPoint int, str string, style State) (err error) {
		var ep EndpointError
		if e := ParseEof(str, style); e == nil && failPoint != -1 {
			err = errors.New("unexpected success")
		} else if !errors.As(e, &ep) {
			err = e
		} else if at := ep.End(); at != failPoint {
			// 0 means okay, -1 incomplete, >0 the one-index of the failure point.
			err = fmt.Errorf("%s len: %d", str, at)
		}
		return
	}
	if e := count(0, "a", AtleastOne(isSpace)); e != nil {
		t.Fatal(e)
	}
	if e := count(0, "a", Optional(isSpace)); e != nil {
		t.Fatal(e)
	}
	if e := count(-1, strings.Repeat(" ", 5), Optional(isSpace)); e != nil {
		t.Fatal(e)
	}
	if e := count(3, strings.Repeat(" ", 3)+"x", Optional(isSpace)); e != nil {
		t.Fatal(e)
	}
}
