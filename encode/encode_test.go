package encode_test

import (
	_ "embed"
	"strings"
	"testing"

	"github.com/ionous/tell/encode"
)

// write various kinds of values
func TestEncoding(t *testing.T) {
	// -----------------------------
	//  #  | input  | encoded results
	// ------------------------------
	testEncoding(t,
		/* 0 */ true, line(`true`),
		/* 1 */ false, line(`false`),
		/* 2 */ "hello", line(`"hello"`),
		/* 3 */ 10, line(`10`),
		/* 4 */ -23, line(`-23`),
		/* 5 */ uint16(12), line(`0xc`),
		/* 6 */ int64(4294967296), line(`4294967296`),
		/* 7 */ 0.1, line(`0.1`),
		/* 8 */ []any{}, line(`[]`),
		/* 9 */
		lines("hello", "there"), chomp(`
|
  hello
  there
  '''`),
		/* 10 */
		lines("hello", "    indents"), chomp(`
|
  hello
      indents
  '''`),
		/* 11 */
		"trailing newlines\n   in heredocs\ndon't collapse\n", chomp(`
|
  trailing newlines
     in heredocs
  don't collapse
  """`),
		/* 12 */
		lines(
			`this implementation`,
			`prefers \ escaping`), chomp(`
|
  this implementation
  prefers \\ escaping\
  """`),
	)
}

// match "encodedTest.tell"
func TestEncodingMap(t *testing.T) {
	testEncoding(t,
		map[string]any{
			"bool":    true,
			"empty":   []any{},
			"hello":   "there",
			"heredoc": lines("a string", "with several lines", "becomes a heredoc."),
			"map": map[string]any{
				"bool":  true,
				"hello": "world",
				"value": 11,
			},
			"nil": nil,
			"slice": []any{
				"5",
				5,
				false,
			},
			"value": 23,
		},
		string(encodedTest),
	)
}

//go:embed encodedTest.tell
var encodedTest []byte

func line(s string) string {
	return s + "\n"
}

func lines(s ...string) string {
	return strings.Join(s, "\n")
}

// gofmt is problematic for strings ( and comments! )
func chomp(s string) string {
	return s[1:] + "\n"
}

// tests without comments
func testEncoding(t *testing.T, pairs ...any) {
	cnt := len(pairs)
	if cnt&1 != 0 {
		panic("mismatched tests")
	}
	var buf strings.Builder
	enc := encode.MakeEncoder(&buf)
	for i := 0; i < cnt; i += 2 {
		src, expect := pairs[i], pairs[i+1].(string)
		if e := enc.Encode(src); e != nil {
			t.Errorf("failed to marshal test #%d because %v", i/2, e)
		} else {
			if got := buf.String(); got != expect {
				t.Logf("have\n%s", got)
				t.Logf("want\n%s", expect)
				t.Logf("have\n%v", []byte(got))
				t.Logf("want\n%v", []byte(expect))
				//
				t.Errorf("failed test #%d", i/2)
			}
			buf.Reset()
		}
	}
	return
}
