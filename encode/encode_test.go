package encode

import (
	"fmt"
	"strings"
	"testing"
)

func TestEncoding(t *testing.T) {
	if e := testEncoding(t,
		/* 0 */ true, line(`true`),
		/* 1 */ false, line(`false`),
		/* 2 */ "hello", line(`"hello"`),
		/* 3 */ 10, line(`10`),
		/* 4 */ -23, line(`-23`),
		/* 5 */ uint16(12), line(`0xc`),
		/* 6 */ int64(4294967296), line(`4294967296`),
		/* 7 */ 0.1, line(`0.1`),
		/* 8 */ map[string]any{
			"hello": "there",
			"value": 23,
			"bool":  true,
			"map": map[string]any{
				"bool":  true,
				"hello": "world",
				"value": 11,
			},
			"slice": []any{
				"5",
				5,
				false,
			},
		},
		`bool: true
hello: "there"
map:
  bool: true
  hello: "world"
  value: 11
slice:
  - "5"
  - 5
  - false
value: 23
`,
	); e != nil {
		t.Fatal(e)
	}
}

func line(s string) string {
	return s + "\n"
}

func testEncoding(t *testing.T, pairs ...any) (err error) {
	cnt := len(pairs)
	if cnt&1 != 0 {
		panic("mismatched tests")
	}
	var buf strings.Builder
	enc := MakeEncoder(&buf)
	for i := 0; i < cnt; i += 2 {
		src, expect := pairs[i], pairs[i+1].(string)
		if e := enc.Encode(src); e != nil {
			err = fmt.Errorf("failed to marshal test #%d because %v", i/2, e)
			break
		} else {
			if got := buf.String(); got != expect {
				t.Logf("have %q", got)
				t.Logf("want %q", expect)

				err = fmt.Errorf("failed test #%d", i/2)
				break
			}
			buf.Reset()
		}
	}
	return
}
