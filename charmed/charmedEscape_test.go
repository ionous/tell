package charmed

import (
	"strings"
	"testing"

	"github.com/ionous/tell/charm"
)

func TestEscape(t *testing.T) {
	// in all of these, an initial backslash is implied:
	// \<test>, expected value
	tests := []string{
		// single character escapes:
		"a", "\a",
		"b", "\b",
		"f", "\f",
		"n", "\n",
		"r", "\r",
		"t", "\t",
		"v", "\v",
		`\`, `\`,
		`"`, `"`,
		//
		"x26", "\x26",
		"u0026", "\u0026",
		"uFFFD", "\uFFFD",
		"U00000026", "\U00000026",
		//

		// octal not supported... fix?
		"000", "error: '0' is not recognized after a backslash",
		"007", "error: '0' is not recognized after a backslash",
		// from go's ref spec https://go.dev/ref/spec
		"k", "error: 'k' is not recognized after a backslash",
		"xa", "error: expected 2 hex values", // illegal: too few hexadecimal digits
		"uDFFF", "error: invalid rune", // illegal: surrogate half
		"U00110000", "error: invalid rune", // illegal: invalid Unicode code point
	}
	for i, cnt := 0, len(tests); i < cnt; i += 2 {
		test, expect := tests[i+0], tests[i+1]
		var buf strings.Builder
		if e := charm.ParseEof(test, decodeEscape(&buf)); e != nil {
			if !strings.HasPrefix(expect, "error:") {
				t.Logf("failed test %d (%q) because %v", i, test, e)
				t.Fail()
			} else {
				got := "error: " + e.Error()
				if !strings.HasPrefix(got, expect) {
					t.Logf("failed test %d (%q), unexpected error %q", i, test, e)
					t.Fail()
				}
			}
		} else {
			if got := buf.String(); got != expect {
				t.Logf("failed test %d  (%q) because got %q", i, test, got)
				t.Fail()
			}
		}
	}
}
