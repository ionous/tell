package charm_test

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/ionous/tell/charm"
)

func TestRequires(t *testing.T) {
	isSpace := func(r rune) bool { return r == ' ' }

	// index of the fail point, or -1 if success is expected
	count := func(failPoint int, str string, style charm.State) (err error) {
		p := charm.MakeParser(strings.NewReader(str))
		if e := p.ParseEof(style); e == nil && failPoint != -1 {
			err = errors.New("unexpected success")
		} else if failPoint == -1 {
			err = e // expected success; if err is not nil caller will fail.
		} else if at := p.Offset(); at != failPoint {
			// 0 means okay, -1 incomplete, >0 the one-index of the failure point.
			err = fmt.Errorf("%s len: %d", str, at)
		}
		return
	}
	if e := count(0, "a", charm.AtleastOne(isSpace)); e != nil {
		t.Fatal(e)
	}
	if e := count(0, "a", charm.Optional(isSpace)); e != nil {
		t.Fatal(e)
	}
	if e := count(-1, strings.Repeat(" ", 5), charm.Optional(isSpace)); e != nil {
		t.Fatal(e)
	}
	if e := count(3, strings.Repeat(" ", 3)+"x", charm.Optional(isSpace)); e != nil {
		t.Fatal(e)
	}
}
