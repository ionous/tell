package charmed

import (
	_ "embed"
	"strings"
	"testing"

	"github.com/ionous/tell/charm"
	"github.com/ionous/tell/runes"
)

//go:embed hereDoc.test
var hereDoc string

//go:embed hereDocRaw.test
var hereDocRaw string

//go:embed hereExpected.test
var hereExpected string

// test reading a full heredoc
func TestHereNow(t *testing.T) {
	if got, e := testHere(hereDocRaw); e != nil {
		t.Fatal("failed hereDocRaw", e)
	} else if got != hereExpected {
		t.Fatalf("hereDocRaw: \nhave: %q\nwant: %q", got, hereExpected)
	}
	if got, e := testHere(hereDoc); e != nil {
		t.Fatal("failed hereDoc", e)
	} else if got != hereExpected {
		t.Fatalf("hereDoc: \nhave: %q\nwant: %q", got, hereExpected)
	}
}

// test for tokenization of heredoc headers
// ( not every series of tokens form a legal header; this doesn't test for that.
//
//	ex. legal headers allow at most one redirect triplet, and it should always be followed by exactly one word.)
func TestHeader(t *testing.T) {
	if got, e := testHeader("yaml<<<END"); e != nil {
		t.Fatal(e)
	} else if expect := "yaml[headerWord][headerRedirect]END[headerWord]"; got != expect {
		t.Fatal("got:", got)
	} else if got, e := testHeader("yaml  <<<  END"); e != nil {
		t.Fatal(e)
	} else if expect := "yaml[headerWord][headerRedirect]END[headerWord]"; got != expect {
		t.Fatal("got:", got)
	} else if got, e := testHeader("<<<"); e != nil {
		t.Fatal(e)
	} else if expect := "[headerRedirect]"; got != expect {
		t.Fatal("got:", got)
	}
}

// expect three redirect markers; no more, no less.
func TestRedirectCount(t *testing.T) {
	for i := 1; i < 5; i++ {
		str := strings.Repeat("<", i)
		_, e := testHeader(str)
		expectError := i != 3
		ok := (e != nil) == expectError
		if !ok {
			t.Fatalf("expected error %v, have %q", expectError, e)
		}
	}
}

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

func TestBody(t *testing.T) {
	if got, e := testBody("!!"); e != nil {
		t.Fatal(e)
	} else if expect := "[lineClose]"; got != expect {
		t.Fatalf("\nhave: %q\nwant: %q", got, expect)
	}
	if got, e := testBody("boop\nbop\nbeep\n!!"); e != nil {
		t.Fatal(e)
	} else if expect := "boop[lineText]bop[lineText]beep[lineText][lineClose]"; got != expect {
		t.Fatalf("\nhave: %q\nwant: %q", got, expect)
	}
	if got, e := testBody("!partial!\n!!"); e != nil {
		t.Fatal(e)
	} else if expect := "!partial![lineText][lineClose]"; got != expect {
		t.Fatalf("\nhave: %q\nwant: %q", got, expect)
	}
}

func testHere(str string) (ret string, err error) {
	var buf strings.Builder
	quote, str := rune(str[0]), str[3:] // decodeHereAfter starts after the opening
	escape := quote == runes.InterpretQuote
	if e := charm.ParseEof(str, decodeHereAfter(&buf, quote, escape)); e != nil {
		err = e
	} else {
		ret = buf.String()
	}
	return
}

func testBody(str string) (ret string, err error) {
	var escape bool
	var buf strings.Builder
	var endTag = []rune{'!', '!'}
	if e := parse(str, decodeLines(&buf, escape, endTag, func(cat lineType, lhs, rhs int) (_ error) {
		buf.WriteRune('[')
		buf.WriteString(cat.String())
		buf.WriteRune(']')
		return
	})); e != nil {
		err = e
	} else {
		ret = buf.String()
	}
	return
}

func testHeader(str string) (ret string, err error) {
	var buf strings.Builder
	if e := parse(str, decodeHeaderHere(&buf, func(t headerToken) (_ error) {
		buf.WriteRune('[')
		buf.WriteString(t.String())
		buf.WriteRune(']')
		return
	})); e != nil {
		err = e
	} else {
		ret = buf.String()
	}
	return
}

// have to do this manually to avoid issues with eof
// ( probably the set of testHeader functions exposed by charm are less than ideal )
func parse(str string, next charm.State) (err error) {
	for _, q := range str + "\n" {
		if next = next.NewRune(q); next == nil {
			break
		} else if es, ok := next.(charm.Terminal); ok {
			err = es
			break
		}
	}
	return
}

func resolve(buf *strings.Builder) (ret string) {
	ret = buf.String()
	buf.Reset()
	return
}
