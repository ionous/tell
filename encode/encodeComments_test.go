package encode_test

import (
	_ "embed"
	"encoding/json"
	"errors"
	"io/fs"
	"strings"
	"testing"

	"github.com/ionous/tell/collect/orderedmap"
	"github.com/ionous/tell/encode"
	"github.com/ionous/tell/testdata"
)

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
