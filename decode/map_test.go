package decode_test

import (
	"errors"
	"testing"

	"github.com/ionous/tell/collect/imap"
)

func TestMap(t *testing.T) {
	test(t,
		// -----------
		"test keys with boolean names", `
false: true
true: false`,
		imap.ItemMap{
			{"false:", true},
			{"true:", false},
		},
		// the token parser is gready
		// so "true luck:" is a bool followed by a key
		// and therefore illegal. fix?
		"test bool with space", `
true luck: false`,
		errors.New("unexpected"),

		// -----------
		"test single value", `
name: "Sammy Sosa"`,
		imap.ItemMap{
			{"name:", "Sammy Sosa"},
		},
		// -----------
		"test split line", `
name:
  "Sammy Sosa"`,
		imap.ItemMap{
			{"name:", "Sammy Sosa"},
		},

		// -----------
		"test several values", `
name: "Sammy Sosa"
hr:   63
avg:  true`,
		imap.ItemMap{
			{"name:", "Sammy Sosa"},
			{"hr:", 63},
			{"avg:", true},
		},
		// -----------------------
		"test map with nil value", `
Field:
Next: 5`,
		imap.ItemMap{
			{"Field:", nil},
			{"Next:", 5},
		},

		// -----------------------
		"test nested maps", `
Field:
  Next: 5`,
		imap.ItemMap{
			{"Field:", imap.ItemMap{
				{"Next:", 5},
			}},
		},

		// -----------------------
		// in yaml, inline nested maps are invalid
		// should they be here too?
		// to do, i think Value would need to examine history
		// either sniffing prior types or through a flag (ex. require newlines)
		// that it can send into NewMapping
		"test inline maps", `
Field: Next: 5`,
		imap.ItemMap{{
			"Field:", imap.ItemMap{{
				"Next:", 5,
			}},
		}},
	)
}
