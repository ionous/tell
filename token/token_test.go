package token_test

import (
	"errors"
	"fmt"
	"log"
	"reflect"
	"strings"
	"testing"

	"github.com/ionous/tell/charm"
	"github.com/ionous/tell/runes"
	"github.com/ionous/tell/token"
)

func TestError(t *testing.T) {
	expect := errors.New("couldn't read words.")
	if e := testOne(token.Invalid, "beep", expect); e != nil {
		t.Fatal(e)
	} else if e := testOne(token.Invalid, "falsey", expect); e != nil {
		t.Fatal(e)
	}
}

// test (at least) one of each of the possible tokens produced
func TestArray(t *testing.T) {
	var tests = []struct {
		str    string
		expect []result
	}{{
		`[]`, []result{{
			tokenType: token.Array, tokenValue: runes.ArrayOpen,
		}, {
			tokenType: token.Array, tokenValue: runes.ArrayClose,
		}},
	}, {
		`[1,,2, "hello"]`, []result{{
			tokenType: token.Array, tokenValue: runes.ArrayOpen,
		}, {
			tokenType: token.Number, tokenValue: 1,
		}, {
			tokenType: token.Array, tokenValue: runes.ArraySeparator,
		}, {
			tokenType: token.Array, tokenValue: runes.ArraySeparator,
		}, {
			tokenType: token.Number, tokenValue: 2,
		}, {
			tokenType: token.Array, tokenValue: runes.ArraySeparator,
		}, {
			tokenType: token.String, tokenValue: "hello",
		}, {
			tokenType: token.Array, tokenValue: runes.ArrayClose,
		}},
	}}
	for i, test := range tests {
		var pairs results
		run := token.NewTokenizer(&pairs)
		if _, e := charm.Parse(test.str, run); e != nil {
			t.Fatal("failed test", i, e)
		} else if e := pairs.compare(test.expect); e != nil {
			t.Fatal("failed test", i, e)
		}
	}
}

// test (at least) one of each of the possible tokens produced
func TestTokens(t *testing.T) {
	tests := []any{
		// token, string to parse, result:
		/*1*/ token.Bool, `true`, true,
		/*2*/ token.Number, `5`, 5,
		/*3*/ token.Number, `0x20`, uint(0x20),
		/*4*/ token.Number, `5.4`, 5.4,
		/*5*/ token.String, `"5.4"`, "5.4",

		// ----------
		/*6*/ token.String,
		`"hello\\world"`,
		`hello\world`,

		// ----------
		/*7*/ token.String,
		"`" + `hello\\world` + "`",
		`hello\\world`,

		// -----
		/*8*/ token.Comment, "# comment", "# comment",
		/*9*/ token.Key, "-", "",
		/*10*/ token.Key, "hello:world:", "hello:world:",
		// make sure dash numbers are treated as negative numbers
		/*11*/ token.Number, `-5`, -5,
		// ----------
		token.String,
		`"""
hello
doc
"""`,
		`hello doc`,
		// ----------
		token.String,
		`|
yaml compatibility block
"""`,
		`yaml compatibility block`,
		// -------------
		token.String,
		strings.Join([]string{
			"```",
			"hello",
			"line",
			"```"}, "\n"),
		`hello
line`,
	}

	// test all of the above in both the same and separate buffers
	// at the very least it helps to validate tokens must be separated by whitespace.
	var combined results
	run := token.NewTokenizer(&combined)

	for i := 0; i < len(tests); i += 3 {
		wantType := tests[i+0].(token.Type)
		testStr := tests[i+1].(string)
		wantVal := tests[i+2]
		whichTest := 1 + i/3
		if e := testOne(wantType, testStr, wantVal); e != nil {
			t.Logf("failed single %d: %s", whichTest, e)
			t.Fail()
		} else {
			sep := " "
			if wantType == token.Comment {
				sep = "\n" // comments have to be ended with a newlne
			}
			if next, e := charm.Parse(testStr+sep, run); e != nil {
				t.Logf("failed combine parse %d: %s", whichTest, e)
				t.Fail()
			} else {
				last := combined[len(combined)-1]
				if e := last.compare(wantType, wantVal); e != nil {
					t.Logf("failed combine compare %d: %s", whichTest, e)
					t.Fail()
				} else {
					run = next
				}
			}
		}
	}
}

func testOne(tokenType token.Type, testStr string, tokenValue any) (err error) {
	var pairs results
	run := token.NewTokenizer(&pairs)
	if _, e := charm.Parse(testStr+"\n", run); e != nil {
		err = compare(e, tokenValue)
	} else if cnt := len(pairs); cnt == 0 {
		err = errors.New("didn't collect any tokens")
	} else {
		last := pairs[cnt-1]
		if e := compare(last.pos, token.Pos{}); e != nil {
			err = e
		} else {
			err = last.compare(tokenType, tokenValue)
		}
	}
	return
}

type results []result

type result struct {
	pos        token.Pos
	tokenType  token.Type
	tokenValue any
}

func (res *results) Decoded(pos token.Pos, tokenType token.Type, tokenValue any) (_ error) {
	(*res) = append((*res), result{pos, tokenType, tokenValue})
	return
}

// compare everything except pos
func (res results) compare(expects results) (err error) {
	if have, want := len(res), len(expects); have != want {
		log.Println(res)
		err = fmt.Errorf("failed test have %d != want %d", have, want)
	} else {
		for k, el := range res {
			want := expects[k]
			if e := el.compare(want.tokenType, want.tokenValue); e != nil {
				err = e
				break
			}
		}
	}
	return
}

func (p result) compare(wantType token.Type, wantValue any) (err error) {
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
