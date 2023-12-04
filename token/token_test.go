package token_test

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/ionous/tell/charm"
	"github.com/ionous/tell/token"
)

func TestError(t *testing.T) {
	expect := errors.New("couldn't read words.")
	if e := testOne(token.Invalid, "beep", expect); e != nil {
		t.Fatal(e)
	}
}

func TestTokens(t *testing.T) {
	tests := []any{
		// token, string to parse, result:
		token.Comment, "\n", "",
		token.Bool, `true`, true,
		token.Number, `5`, int64(5),
		token.Number, `0x20`, uint64(0x20),
		token.Number, `5.4`, 5.4,
		token.InterpretedString, `"5.4"`, "5.4",

		// ----------
		token.InterpretedString,
		`"hello\\world"`,
		`hello\world`,

		// ----------
		token.RawString,
		"`" + `hello\\world` + "`",
		`hello\\world`,

		// -----
		token.Comment, "# comment", "# comment",
		token.Key, "-", "",
		token.Key, "hello:world:", "hello:world:",
	}

	// test all of the above in both the same and separate buffers
	// at the very least it helps to validate tokens must be separated by whitespace.
	var combined tokenPairs
	run := token.MakeTokenizer(&combined)

	for i := 0; i < len(tests); i += 3 {
		wantType := tests[i+0].(token.Type)
		testStr := tests[i+1].(string)
		wantVal := tests[i+2]
		if e := testOne(wantType, testStr, wantVal); e != nil {
			t.Logf("failed single %d: %s", i/3, e)
			t.Fail()
		} else {
			sep := " "
			if wantType == token.Comment {
				sep = "\n" // comments have to be ended with a newlne
			}
			if next, e := charm.Parse(testStr+sep, run); e != nil {
				t.Logf("failed combine parse %d: %s", i/3, e)
				t.Fail()
			} else {
				last := combined[len(combined)-1]
				if e := last.compare(wantType, wantVal); e != nil {
					t.Logf("failed combine compare %d: %s", i/3, e)
					t.Fail()
				} else {
					run = next
				}
			}
		}
	}
}

func testOne(tokenType token.Type, testStr string, tokenValue any) (err error) {
	var pairs tokenPairs
	run := token.MakeTokenizer(&pairs)
	if _, e := charm.Parse(testStr+"\n", run); e != nil {
		err = compare(e, tokenValue)
	} else if cnt := len(pairs); cnt == 0 {
		err = errors.New("didn't collect any tokens")
	} else {
		last := pairs[cnt-1]
		err = last.compare(tokenType, tokenValue)
	}
	return
}

type tokenPairs []tokenPair

type tokenPair struct {
	tokenType  token.Type
	tokenValue any
}

func (s *tokenPairs) Decoded(tokenType token.Type, tokenValue any) (_ error) {
	(*s) = append((*s), tokenPair{tokenType, tokenValue})
	return
}

func (p tokenPair) compare(wantType token.Type, wantValue any) (err error) {
	if tt := p.tokenType; tt != wantType {
		err = fmt.Errorf("mismatched types want: %s, have: %s", wantType, tt)
	} else {
		err = compare(p.tokenValue, wantValue)
	}
	return
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
