package charmed

import (
	"fmt"
	"strings"
	"testing"

	"github.com/ionous/tell/charm"
)

func TestCollapsingSpace(t *testing.T) {
	if str, e := testQuotes(`"backslashes \"\\\" are \
  special,
"`); e != nil {
		t.Error(e)
	} else {
		const expect = `backslashes "\" are special, `
		if str != expect {
			t.Errorf("\nhave: %q\nwant: %q", str, expect)
		}
	}
}

func TestQuotes(t *testing.T) {
	if x, e := testQ(t, "'singles'", "singles"); e != nil {
		t.Fatal(x, e)
	}
	if x, e := testQ(t, `"doubles"`, "doubles"); e != nil {
		t.Fatal(x, e)
	}
	if x, e := testQ(t, "'escape\"'", "escape\""); e != nil {
		t.Fatal(x, e)
	}
	if x, e := testQ(t, `"\\"`, "\\"); e != nil {
		t.Fatal(x, e)
	}
	if x, e := testQ(t, string([]rune{'"', '\\', 'a', '"'}), "\a"); e != nil {
		t.Fatal(x, e)
	}
	if _, e := testQ(t, string([]rune{'"', '\\', 'g', '"'}),
		ignoreResult); e == nil {
		t.Fatal(e)
	}
}

// scans until the matching quote marker is found
func testRemainingString(match rune, onDone func(string)) (ret charm.State) {
	var buf strings.Builder
	return charm.Step(scanRemainingString(&buf, match, AllowEscapes),
		charm.OnExit("recite", func() {
			onDone(buf.String())
		}))
}

func testQ(t *testing.T, str, want string) (ret interface{}, err error) {
	t.Log("test:", str)
	var got string
	p := charm.Statement("quote", func(r rune) (ret charm.State) {
		if r == '\'' || r == '"' {
			ret = testRemainingString(r, func(res string) {
				got = res
			})
		}
		return
	})
	if e := charm.ParseEof(str, p); e != nil {
		err = e
	} else if want != ignoreResult {
		if got != want {
			err = mismatched(want, got)
		} else {
			t.Log("ok", got)
		}
	}
	return str, err
}

func mismatched(want, got string) error {
	return fmt.Errorf("want(%d): %s; != got(%d): %s.", len(want), want, len(got), got)
}

// for testing errors when we want to fail before the match is tested.
const ignoreResult = "~~IGNORE~~"
