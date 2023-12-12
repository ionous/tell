package charmed

import (
	"strings"
	"testing"

	"github.com/ionous/tell/charm"
)

// test for tokenization of heredoc headers
// ( not every series of tokens form a legal header; this doesn't test for that.
//   ex. legal headers allow at most one redirect triplet, and it should always be followed by exactly one word.)
func TestHeader(t *testing.T) {
	if got, e := parse("yaml<<<END"); e != nil {
		t.Fatal(e)
	} else if expect := "yaml[headerWord][headerRedirect]END[headerWord]"; got != expect {
		t.Fatal("got:", got)
	} else if got, e := parse("yaml  <<<  END"); e != nil {
		t.Fatal(e)
	} else if expect := "yaml[headerWord][headerRedirect]END[headerWord]"; got != expect {
		t.Fatal("got:", got)
	} else if got, e := parse("<<<"); e != nil {
		t.Fatal(e)
	} else if expect := "[headerRedirect]"; got != expect {
		t.Fatal("got:", got)
	}
}

// expect three redirect markers; no more, no less.
func TestRedirectCount(t *testing.T) {
	for i := 1; i < 5; i++ {
		str := strings.Repeat("<", i)
		_, e := parse(str)
		expectError := i != 3
		ok := (e != nil) == expectError
		if !ok {
			t.Fatalf("expected error %v, have %q", expectError, e)
		}
	}
}

// fix: eat trailing space for escaped lines
// how?! ( probably by tracking that in the state and reporting it )
func TestLiteralLines(t *testing.T) {
	var ls indentedLines
	// left side spaces, trailing spaces, and the text.
	ls.addLine(3, 0, "a")
	ls.addLine(4, 2, "b")
	ls.addLine(2, 0, "c")
	var buf strings.Builder
	ls.writeLines(&buf, 2, false /*literalLine*/)
	if got, expect := resolve(&buf),
		" a   b c"; got != expect {
		t.Fatalf("\nhave: %q\nwant: %q", got, expect)
	}
	ls.writeLines(&buf, 2, true /*literalLine*/)
	if got, expect := resolve(&buf),
		" a\n  b  \nc"; got != expect {
		t.Fatalf("\nhave: %q\nwant: %q", got, expect)
	}
}

func resolve(buf *strings.Builder) (ret string) {
	ret = buf.String()
	buf.Reset()
	return
}

// have to do this manually to avoid issues with eof
// ( probably the set of parse functions exposed by charm are less than ideal )
func parse(str string) (ret string, err error) {
	var buf strings.Builder
	next := decodeHeaderHere(&buf, note(&buf))
	for _, q := range str + "\n" {
		if next = next.NewRune(q); next == nil {
			ret = buf.String()
			break
		} else if es, ok := next.(charm.Terminal); ok {
			err = es
			break
		}
	}
	return
}

func note(out *strings.Builder) headerNotifier {
	return func(t headerToken) (_ error) {
		out.WriteRune('[')
		out.WriteString(t.String())
		out.WriteRune(']')
		return
	}
}
