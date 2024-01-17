package decode_test

import (
	"errors"
	"testing"
)

func TestSeq(t *testing.T) {
	test(t,
		// --------------
		"single value",
		[]any{5}, `
- 5`,
		// --------------
		"fail without dash",
		errors.New("unknown number"), `
-false`,
		// --------------
		"value with newline",
		[]any{5}, `
- 5`,
		// --------------
		"split line",
		[]any{5}, `
-
  5`,
		// --------------
		"several values",
		[]any{5, 10, 12}, `
- 5
- 10
- 12`,
		// --------------
		"nested sub sequence",
		[]any{[]any{5}}, `
- - 5`,
		// --------------
		"new line sub sequence",
		[]any{[]any{5}}, `
-
  - 5`,
		// --------------
		"multiple sub values",
		[]any{[]any{nil, 5}}, `
- -
  - 5`,
		// --------------
		"nil values",
		[]any{nil, nil, nil}, `
-
-
-`,
		// --------------
		"nil value trailing newline",
		[]any{nil, nil, nil}, `
-
-
-
`,
		// --------------
		"continuing sub sequence ",
		[]any{[]any{5}, 6}, `
- - 5
- 6`)
}
