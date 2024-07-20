package charmed

import (
	"math"
	"strings"
	"testing"

	"github.com/ionous/tell/charm"
)

func TestNum(t *testing.T) {
	var NaN = math.NaN()
	// returns point of failure
	run := func(str string) (val float64, err error) {
		var num NumParser
		// use the make parser version to gain access to the offset
		p := charm.MakeParser(strings.NewReader(str))
		if e := p.ParseEof(num.Decode()); e != nil {
			val = NaN
			err = endpointError{e, p.Offset()}
		} else if v, e := num.GetFloat(); e != nil {
			err = e // all of the input was okay, but we couldn't make a float of it.
		} else {
			val = v
		}
		return
	}
	tests := []struct {
		input string
		// 0 means success is expected, and the value contains the parsed result
		// -1 means we expect the input to run out before parsing is finished.
		// >0 the one-index of an expected failure point
		endpoint int
		value    float64
	}{
		// bad decimals
		{"0.", -1, NaN}, // it parses just fine, but GetFloat() will error with unknown number
		{".0", 1, NaN},
		// floats
		{"0.0", 0, 0},
		{"0.25", 0, 0.25},
		{"72.40", 0, 72.4},
		{"072.40", 0, 72.4},
		{"2.71828", 0, 2.71828},
		// exponents:
		{"1.e+0", 0, 1},
		{"6.67428e-11", 0, 6.67428e-11},
		{"0e6", 0, 0},
		{"1E6", 0, 1e6},
		// bad exponents
		{"0.12345E+5", 0, 12345},
		{"0.12345E+", -1, NaN},
		{"0.12345E", -1, NaN},
		{"1E6e5", -1, NaN},
		// ints
		{"42", 0, 42},
		{"0600", 0, 600},
		// hex
		{"0xFACADE", 0, 0xfacade},
		{"0Xbadf00d", 0, 0xbadf00d},
		// bad hex:
		{"0x", -1, NaN},
		{"0xg", 3, NaN},
		{"xbadf00d", 1, NaN},
		// other chars:
		{"uncle", 1, NaN},
		{"0uncle", 2, NaN},
		// leading
		{"-5", 0, -5},
		{"+5", 0, 5},
		{"-5.1", 0, -5.1},
		{"+5.1", 0, 5.1},
		// bad leads
		{"-0x5", 3, NaN},
		{"+0x5", 3, NaN},
	}
	// out of range:
	// {"170141183460469231731687303715884105727", 0, 1.7014118346046923e+38},
	for i, test := range tests {
		t.Logf("test%2d: '%s'", i, test.input)
		if v, e := run(test.input); e == nil {
			t.Log("output:", v)
			// no error returned, then our values should match
			if v != test.value {
				t.Fatal("wanted:", test.value)
				break
			}
		} else {
			// error returned, check the expected error
			if test.endpoint == 0 {
				t.Fatal("expected success but received an error", e)
				break
			} else if test.endpoint > 0 {
				// ensure our expected fail point is correct
				if ep := e.(endpointError); ep.pos != test.endpoint-1 {
					t.Fatal("mismatched endpoint at", ep.error)
					break
				}
			}
			t.Log("ok", e)
		}
	}
}

// backwards compatibility for tests
// reports the offset of failure ( if any ) when parsing a number
type endpointError struct {
	error
	pos int
}
