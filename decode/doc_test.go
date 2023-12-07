package decode_test

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"unicode"

	"github.com/ionous/tell/charm"
	"github.com/ionous/tell/decode"
	"github.com/ionous/tell/maps/imap"
	"github.com/ionous/tell/notes"
)

// test a few scalar document values
// the token parser has more exhaustive tests
func TestDocScalar(t *testing.T) {
	test(t,
		"bool", `false`, false,
		"bool", `true`, true,
		"int", `23`, 23,
		"string", `"hello"`, "hello",
		"multi", "true\n5", errors.New("unexpected"),
	)
}

// replace statename with reflection lookup
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

// name of test, input string, expected result
// leading whitespace is trimmed
func test(t *testing.T, nameInputExpect ...any) {
	for i, cnt := 0, len(nameInputExpect); i < cnt; i += 3 {
		name, input, expect := nameInputExpect[0+i].(string), nameInputExpect[1+i].(string), nameInputExpect[2+i]
		if strings.HasPrefix(name, `x `) {
			// commenting out tests causes go fmt to replace spaces with tabs. *sigh*
			t.Log("skipping", name)
		} else {
			var res any
			str := strings.TrimLeftFunc(input, unicode.IsSpace)
			dec := decode.MakeDecoder(imap.Builder, notes.DiscardComments())
			if val, e := dec.Decode(strings.NewReader(str)); e != nil {
				res = e
			} else {
				res = val
			}
			if e := compare(res, expect); e != nil {
				t.Fatal("ng:", name, e)
			} else {
				t.Log("ok:", name)
			}
		}
	}
}

func compare(have any, want any) (err error) {
	if haveErr, ok := have.(error); !ok {
		if !reflect.DeepEqual(have, want) {
			err = fmt.Errorf("mismatched want: %v(%T) have: %v(%T)", want, want, have, have)
		}
	} else {
		if expectErr, ok := want.(error); !ok {
			err = fmt.Errorf("failed %v", haveErr)
		} else if !strings.HasPrefix(haveErr.Error(), expectErr.Error()) {
			err = fmt.Errorf("failed %v, expected %v", haveErr, expectErr)
		}
	}
	return
}
