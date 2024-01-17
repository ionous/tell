package decode_test

import (
	"errors"
	"testing"

	"github.com/ionous/tell/collect/imap"
)

func TestMap(t *testing.T) {
	test(t,
		// -----------
		"test keys with boolean names",
		imap.ItemMap{{
			Key:   "false:",
			Value: true,
		}, {
			Key:   "true:",
			Value: false,
		}}, `
false: true
true: false`,
		// the token parser is gready
		// so "true luck:" is a bool followed by a key
		// and therefore illegal. fix?
		"test bool with space",
		errors.New("unexpected"), `
true luck: false`,
		// -----------
		"test single value",
		imap.ItemMap{{
			Key:   "name:",
			Value: "Sammy Sosa",
		}}, `
name: "Sammy Sosa"`,
		// -----------
		"test split line",
		imap.ItemMap{{
			Key:   "name:",
			Value: "Sammy Sosa",
		}}, `
name:
  "Sammy Sosa"`,

		// -----------
		"test several values",
		imap.ItemMap{{
			Key:   "name:",
			Value: "Sammy Sosa",
		}, {
			Key:   "hr:",
			Value: 63,
		}, {
			Key:   "avg:",
			Value: true},
		}, `
name: "Sammy Sosa"
hr:   63
avg:  true`,
		// -----------------------
		"test map with nil value",
		imap.ItemMap{{
			Key:   "Field:",
			Value: nil,
		}, {
			Key:   "Next:",
			Value: 5,
		}},
		`
Field:
Next: 5`,
		// -----------------------
		"test nested maps",
		imap.ItemMap{{
			Key: "Field:",
			Value: imap.ItemMap{{
				Key:   "Next:",
				Value: 5,
			}}},
		}, `
Field:
  Next: 5`,
		// -----------------------
		// in yaml, inline nested maps are invalid
		// should they be here too?
		// to do, i think Value would need to examine history
		// either sniffing prior types or through a flag (ex. require newlines)
		// that it can send into NewMapping
		"test inline maps",
		imap.ItemMap{{
			Key: "Field:",
			Value: imap.ItemMap{{
				Key:   "Next:",
				Value: 5,
			}},
		}}, `
Field: Next: 5`,
	)
}
