package tell_test

import (
	"errors"
	"strings"
	"testing"
	"unicode"

	"github.com/ionous/tell"
	"github.com/ionous/tell/maps/imap"
)

func TestScalars(t *testing.T) {
	testValue(t,
		"Test number", `5.4`, 5.4,
		"Test string", `"5.4"`, "5.4",
		"Test bool", `true`, true,

		// ----------
		"Test interpreted",
		`"hello\\world"`,
		`hello\world`,

		// ----------
		"Test raw",
		"`"+`hello\\world`+"`",
		`hello\\world`,

		// -----
		"Test unquoted value",
		"beep",
		errors.New("signature must end with a colon"),
	)
}

//  name of test, input string, expected result
// leading whitespace is trimmed
func testValue(t *testing.T, nameInputExpect ...any) {
	for i, cnt := 0, len(nameInputExpect); i < cnt; i += 3 {
		name, input, expect := nameInputExpect[0+i].(string), nameInputExpect[1+i].(string), nameInputExpect[2+i]
		if strings.HasPrefix(name, `x `) {
			// commenting out tests causes go fmt to replace spaces with tabs. *sigh*
			t.Log("skipping", name)
		} else {
			var res any
			doc := tell.Document{MakeMap: imap.Builder}
			str := strings.TrimLeftFunc(input, unicode.IsSpace)
			if got, e := doc.ReadDoc(strings.NewReader(str)); e != nil {
				res = e
			} else {
				res = got.Content
			}
			if e := compare(res, expect); e != nil {
				t.Fatal("ng:", name, e)
			} else {
				t.Log("ok:", name)
			}
		}
	}
}
