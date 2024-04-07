package decode_test

import (
	"errors"
	"testing"

	"github.com/ionous/tell/collect/imap"
)

// -------------------------------------------------------------
// ..- Mapping:   # this key is assigned "value"
// ...."value"
//
// ..- Mapping:   # this key is assigned nil because
// ....Same Map:  # same indent with a new key uses the same map
//
// ..- Mapping:        # here, the value is a sequence.
// ....- "new sequence" # in 0.7 this generated an error
// -------------------------------------------------------------
func TestBug(t *testing.T) {
	test(t,
		// --------------
		"single value",
		[]any{imap.ItemMap{{
			Key:   "First:",
			Value: []any{"one"},
		}, {
			Key:   "Second:",
			Value: []any{"other"},
		}}}, `
- First:
  - "one"
  Second:
  - "other"`,
	)
}

// verify that an erroneous tab is caught
func TestTabCatching(t *testing.T) {
	var badTab = "- - \"First\"\n\t- \"Second\""
	test(t,
		"bad tab",
		errors.New("tabs are invalid"),
		badTab)
}
