package encode_test

import (
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"strings"
	"testing"

	"github.com/ionous/tell/collect/orderedmap"
	"github.com/ionous/tell/encode"
	"github.com/ionous/tell/testdata"
)

//go:embed encodedTest.tell
var encodedTest []byte

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
		/* 8 */ []any{}, line(`[]`),
		/* 9 */ map[string]any{
			"hello": "there",
			"empty": []any{},
			"value": 23,
			"bool":  true,
			"nil":   nil,
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
			"heredoc": `a string
with several lines
becomes a heredoc.`,
		},
		string(encodedTest),
	); e != nil {
		t.Fatal(e)
	}
}

func TestCommentEncoding(t *testing.T) {
	filePath := "smallCatalogComments"
	if e := func() (err error) {
		if want, e := fs.ReadFile(testdata.Tell, filePath+".tell"); e != nil {
			err = e
		} else if b, e := fs.ReadFile(testdata.Json, filePath+".json"); e != nil {
			err = e
		} else {
			var m orderedmap.OrderedMap
			if e := json.Unmarshal(b, &m); e != nil {
				err = e
			} else if src, ok := m.Get("content"); !ok {
				err = errors.New("missing content")
			} else {
				var buf strings.Builder
				enc := encode.MakeCommentEncoder(&buf)
				if e := enc.Encode(&src); e != nil {
					err = e
				} else if have, want := buf.String(), string(want); have != want {
					err = errors.New("mismatched")
					t.Log(want)
					t.Log(have)
				}
			}
		}
		return
	}(); e != nil {
		t.Fatal(e)
	}
}

func line(s string) string {
	return s + "\n"
}

// tests without comments
func testEncoding(t *testing.T, pairs ...any) (err error) {
	cnt := len(pairs)
	if cnt&1 != 0 {
		panic("mismatched tests")
	}
	var buf strings.Builder
	enc := encode.MakeEncoder(&buf)
	for i := 0; i < cnt; i += 2 {
		src, expect := pairs[i], pairs[i+1].(string)
		if e := enc.Encode(src); e != nil {
			err = fmt.Errorf("failed to marshal test #%d because %v", i/2, e)
			break
		} else {
			if got := buf.String(); got != expect {
				t.Logf("have\n%s", got)
				t.Logf("want\n%s", expect)
				//
				err = fmt.Errorf("failed test #%d", i/2)
				break
			}
			buf.Reset()
		}
	}
	return
}
