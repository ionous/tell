package notes

import (
	"strings"
	"testing"

	"github.com/ionous/tell/charm"
)

// decodes a few comments as if they existed
// between a key ( or dash ) and an incoming value.
// looks under the hood to see if the comments went to the right spot.
func TestKeyBuffering(t *testing.T) {
	// test an ever increasing amount of these lines over time:
	// index 0..1, ... 0..number of total lines
	lines := []string{
		"# one",
		"\t# nest",
		"# two",
		"\t# nest",
		"# three",
		"\t# nest",
	}

	// pairs expected container, buffer output
	expectations := []string{
		// the first block always goes to output, no buffering
		/* 0 */ "\r# one", "",
		/* 1 */ "\r# one\n\t# nest", "",
		// the next block goes to the buffer
		/* 2 */ "\r# one\n\t# nest", "# two",
		/* 3 */ "\r# one\n\t# nest", "# two\n\t# nest",
		// flushing should keep at most one block in the buffer
		/* 4 */ "\r# one\n\t# nest\n# two\n\t# nest", "# three",
		/* 5 */ "\r# one\n\t# nest\n# two\n\t# nest", "# three\n\t# nest",
	}

	for i := 0; i < len(lines); i++ {
		test := strings.Join(lines[:i+1], "\n")
		a, b := expectations[i*2+0], expectations[i*2+1]

		var str strings.Builder
		ctx := newContext(&str)
		key := makeKeyComments(ctx)
		if _, e := charm.Parse(test, &key); e != nil {
			t.Fatal(e)
		} else {
			out, buf := str.String(), ctx.resolveBuffer()
			if out != a || buf != b {
				t.Logf("test %d got:\n%q\n:%q", i, out, buf)
				t.Fail()
			}
		}
	}
}

// once it sees a blank line, everything else should buffer
// with no flushing.
func TestKeyBlank(t *testing.T) {
	// test a leading blank line
	lines := []string{
		"\n# one\n# two",
	}

	// pairs expected container, buffer output
	expectations := []string{
		"", // a blank line mean only buffered content
		"# one\n# two",
	}

	for i := 0; i < len(lines); i++ {
		test := lines[0]
		a, b := expectations[i*2+0], expectations[i*2+1]

		var str strings.Builder
		ctx := newContext(&str)
		key := makeKeyComments(ctx)
		if _, e := charm.Parse(test, &key); e != nil {
			t.Fatal(e)
		} else {
			out, buf := str.String(), ctx.resolveBuffer()
			if out != a || buf != b {
				t.Logf("test %d got:\n%q\n:%q", i, out, buf)
				t.Fail()
			}
		}
	}
}
