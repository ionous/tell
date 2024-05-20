package charmed

import (
	"embed"
	"fmt"
	"io/fs"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/ionous/tell/charm"
	"github.com/ionous/tell/runes"
)

//go:embed _testdata/here*
var hereTest embed.FS

//go:embed _testdata/expectedResults
var expectedResults string

// test reading a full heredoc
func TestHereNow(t *testing.T) {
	if e := fs.WalkDir(hereTest, ".", func(path string, d fs.DirEntry, e error) (err error) {
		if e != nil {
			err = e
		} else if !d.IsDir() {
			if b, e := fs.ReadFile(hereTest, path); e != nil {
				err = e
			} else {
				str, name := string(b), d.Name()
				if got, e := testHere(str); e != nil {
					t.Errorf("failed %s %s", name, e)
				} else if got != expectedResults {
					t.Errorf("here %s: \nhave: %q\nwant: %q", name, got, expectedResults)
				}
			}
		}
		return
	}); e != nil {
		t.Fatal(e)
	}
}

// test for tokenization of heredoc headers
// ( not every series of tokens form a legal header; this doesn't test for that.
// - ex. legal headers allow at most one redirect triplet, and it should always be followed by exactly one word.)
func TestHeader(t *testing.T) {
	if got, e := testHeader("lang<<<END"); e != nil {
		t.Fatal(e)
	} else if expect := "lang[headerWord][headerRedirect]END[headerWord]"; got != expect {
		t.Fatal("got:", got)
	} else if got, e := testHeader("lang  <<<  END"); e != nil {
		t.Fatal(e)
	} else if expect := "lang[headerWord][headerRedirect]END[headerWord]"; got != expect {
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

// use the lower level "indentLines" writer
// and verify escaping and final line chomping
func TestHereLines(t *testing.T) {
	var ls indentedBlock
	// left side spaces, and the text.
	ls.addLine(3, "a\n")
	ls.addLine(4, "b  \\\n")
	ls.addLine(2, "c\n")
	var buf strings.Builder
	// raw, keep trailing newline
	ls.writeHere(&buf, runes.QuoteRaw, 2)
	if got, expect := resolve(&buf),
		" a\n  b  \\\nc\n"; got != expect {
		t.Errorf("\nhave: %q\nwant: %q", got, expect)
	}
	// raw, discard trailing newline
	ls.writeHere(&buf, runes.QuoteSingle, 2)
	if got, expect := resolve(&buf),
		" a\n  b  \\\nc"; got != expect {
		t.Errorf("\nhave: %q\nwant: %q", got, expect)
	}
	// interpreted, keep trailing newline
	ls.writeHere(&buf, runes.QuoteDouble, 2)
	if got, expect := resolve(&buf),
		" a\n  b  c\n"; got != expect {
		t.Errorf("\nhave: %q\nwant: %q", got, expect)
	}
}

func TestCustomTag(t *testing.T) {
	if got, e := testCustomTag("!!"); e != nil {
		t.Fatal(e)
	} else if expect := ""; got != expect {
		t.Errorf("\nhave: %q\nwant: %q", got, expect)
	}
	if got, e := testCustomTag("boop\nbop\nbeep\n!!"); e != nil {
		t.Fatal(e)
	} else if expect := "boop\nbop\nbeep"; got != expect {
		t.Errorf("\nhave: %q\nwant: %q", got, expect)
	}
	if got, e := testCustomTag("!partial!\n!!"); e != nil {
		t.Fatal(e)
	} else if expect := "!partial!"; got != expect {
		t.Errorf("\nhave: %q\nwant: %q", got, expect)
	}
}

func testHere(str string) (ret string, err error) {
	var d QuoteDecoder
	q, size := utf8.DecodeRuneInString(str)
	if next, ok := d.DecodeQuote(q); !ok {
		err = fmt.Errorf("unhandled rune %q", q)
	} else if e := charm.ParseEof(str[size:], next); e != nil {
		err = e
	} else {
		ret = d.String()
	}
	return
}

func testCustomTag(str string) (ret string, err error) {
	var buf strings.Builder // for no particular reason trims the final newline
	next := decodeUntilCustom(&buf, runes.QuoteSingle, []rune{'!', '!'})
	if e := charm.ParseEof(str, next); e != nil {
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
