package decode_test

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"unicode"

	"github.com/ionous/tell/charm"
	"github.com/ionous/tell/collect/imap"
	"github.com/ionous/tell/collect/stdseq"
	"github.com/ionous/tell/decode"
)

// test a few scalar document values
// the token parser has more exhaustive tests
func TestDocScalar(t *testing.T) {
	test(t,
		// testName, result, docValue:
		"bool", false, `false`,
		"bool", true, `true`,
		"int", 23, `23`,
		"string", "hello", `"hello"`,
		// a document shouldn't allow multiple scalar values
		"multi", errors.New("unexpected"), "true\n5",
		"empty array", []any{}, `[]`,
		"array of one", []any{1}, `[1]`,
	)
}

// replace state name with reflection lookup
// could be put in a charm helper package
func init() {
	charm.StateName = func(n charm.State) (ret string) {
		if s, ok := n.(interface{ String() string }); ok {
			ret = s.String()
		} else if n == nil {
			ret = "null"
		} else {
			ret = reflect.TypeOf(n).Elem().Name()
		}
		return
	}
}

// name of test, expected result, input string
// leading whitespace is trimmed
func test(t *testing.T, nameInputExpect ...any) {
	for i, cnt := 0, len(nameInputExpect); i < cnt; i += 3 {
		name, input, expect := nameInputExpect[0+i].(string), nameInputExpect[2+i].(string), nameInputExpect[1+i]
		if strings.HasPrefix(name, `x `) {
			// commenting out tests causes go fmt to replace spaces with tabs. *sigh*
			t.Log("skipping", name)
		} else {
			var res any
			if v, e := decodeString(input); e != nil {
				res = e
			} else {
				res = v
			}
			if e := compare(t, res, expect); e != nil {
				t.Fatal("ng:", name, e)
			} else {
				t.Log("ok:", name)
			}
		}
	}
}

func decodeString(input string) (ret any, err error) {
	str := strings.TrimLeftFunc(input, unicode.IsSpace)
	var dec decode.Decoder
	dec.SetMapper(imap.Make)
	dec.SetSequencer(stdseq.Make)
	return dec.Decode(strings.NewReader(str))
}

func compare(t *testing.T, have any, want any) (err error) {
	if haveErr, ok := have.(error); !ok {
		if !reflect.DeepEqual(have, want) {
			err = fmt.Errorf("mismatched want: %v(%T) have: %v(%T)", want, want, have, have)
		}
	} else {
		if expectErr, ok := want.(error); !ok {
			err = fmt.Errorf("failed %v", haveErr)
		} else if !strings.Contains(haveErr.Error(), expectErr.Error()) {
			err = fmt.Errorf("failed %v, expected %v", haveErr, expectErr)
		} else {
			t.Logf("okay, expected error and got %q", haveErr)
		}
	}
	return
}
