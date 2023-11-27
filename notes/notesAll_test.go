package notes

import (
	"strings"
	"testing"

	"github.com/ionous/tell/charm"
)

// ensure multiple lines are joined together correctly
func TestReadAll(t *testing.T) {
	// test an ever increasing amount of these lines over time:
	// index 0..1, ... 0..number of total lines
	lines := []string{
		"# one",
		"\t# nest", // the alert rune, alerts us to nesting.
		"# two",
		"\t# a",
		"\t# b",
		"\t# c",
		"# three",
		"\t# nest",
	}

	// pairs expected container
	expectations := []string{
		/* 0 */ "# one",
		/* 1 */ "# one\n\t# nest",
		/* 2 */ "# one\n\t# nest\n# two",
		/* 3 */ "# one\n\t# nest\n# two\n\t# a",
		/* 4 */ "# one\n\t# nest\n# two\n\t# a\n\t# b",
		/* 5 */ "# one\n\t# nest\n# two\n\t# a\n\t# b\n\t# c",
		/* 6 */ "# one\n\t# nest\n# two\n\t# a\n\t# b\n\t# c\n# three",
		/* 7 */ "# one\n\t# nest\n# two\n\t# a\n\t# b\n\t# c\n# three\n\t# nest",
	}

	for i := 0; i < len(lines); i++ {
		test := strings.Join(lines[:i+1], "\n")
		a := expectations[i]

		var b strings.Builder
		if e := charm.Parse(test, readAll(&b)); e != nil {
			t.Fatal(e)
		} else {
			out := b.String()
			if out != a {
				t.Logf("test %d got:\n%q", i, out)
				t.Fail()
			}
		}
	}
}
