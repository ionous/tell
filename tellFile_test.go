package tell_test

import (
	"bufio"
	"embed"
	"encoding/json"
	"io"
	"path"
	"reflect"
	"strings"
	"testing"

	"github.com/ionous/tell"
	"github.com/ionous/tell/decode"
)

//go:embed testdata/*.tell
var tellData embed.FS

//go:embed testdata/*.json
var jsonData embed.FS

const testFolder = "testdata"

// helper for debugging specific tests
var focus string

func TestFiles(t *testing.T) {
	// focus = "trailingComments2"
	if files, e := tellData.ReadDir(testFolder); e != nil {
		t.Fatal(e)
	} else {
		for _, info := range files {
			shortName := info.Name()
			tellName := path.Join(testFolder, shortName)
			jsonName := tellName[:len(tellName)-4] + "json"
			if (len(focus) > 0 && !strings.Contains(tellName, focus)) ||
				strings.HasPrefix(shortName, "x_") {
				// t.Log("skipping", tellName)
				continue
			}
			//
			t.Log("trying", tellName)
			if got, e := readTell(tellName); e != nil {
				t.Log("error", e)
				t.Fail()
			} else if want, e := readJson(jsonName); e != nil {
				t.Log(stringify(got))
				t.Fail()
			} else {
				if !reflect.DeepEqual(got, want) {
					t.Log("ng: ", jsonName)
					t.Log("got:", stringify(got))
					t.Fail()
				} else {
					t.Log("ok: ", jsonName)
				}
			}
		}
	}
}

func stringify(got any) (ret string) {
	if a, e := json.MarshalIndent(got, "", " "); e != nil {
		panic(e)
	} else {
		ret = string(a)
	}
	return
}

func readTell(filePath string) (ret any, err error) {
	if fp, e := tellData.Open(filePath); e != nil {
		err = e
	} else {
		var res any
		var docComments decode.CommentBlock
		dec := tell.NewDecoder(bufio.NewReader(fp))
		dec.UseFloats() // because json does
		if keepComments := strings.Contains(strings.ToLower(filePath), "comment"); keepComments {
			dec.UseNotes(&docComments)
		}
		if e := dec.Decode(&res); e != nil {
			err = e
		} else if str, _ := docComments.Resolve(); len(str) > 0 {
			ret = map[string]any{
				"content": res,
				"comment": str,
			}
		} else {
			ret = map[string]any{
				"content": res,
			}
		}
	}
	return
}

func readJson(filePath string) (ret any, err error) {
	if fp, e := jsonData.Open(filePath); e != nil {
		err = e
	} else if b, e := io.ReadAll(fp); e != nil {
		err = e
	} else if e := json.Unmarshal(b, &ret); e != nil {
		err = e
	}
	return
}
