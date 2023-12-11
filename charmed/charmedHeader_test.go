package charmed

import (
	"strings"
	"testing"

	"github.com/ionous/tell/charm"
)

// test for tokenization of heredoc headers
// ( not every series of tokens form a legal header; this doesn't test for that.
//   ex. legal headers allow at most one stream, and it should always be followed by exactly one word.)
func TestHeader(t *testing.T) {
	if got, e := parse("yaml<<<END"); e != nil {
		t.Fatal(e)
	} else if expect := "yaml[headerWord][headerStream]END[headerWord]"; got != expect {
		t.Fatal("got:", got)
	} else if got, e := parse("yaml  <<<  END"); e != nil {
		t.Fatal(e)
	} else if expect := "yaml[headerWord][headerStream]END[headerWord]"; got != expect {
		t.Fatal("got:", got)
	} else if got, e := parse("<<<"); e != nil {
		t.Fatal(e)
	} else if expect := "[headerStream]"; got != expect {
		t.Fatal("got:", got)
	}
}

// expect three stream markers; no more, no less.
func TestStreamCount(t *testing.T) {
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
