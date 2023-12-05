package decode_test

import (
	"errors"
	"testing"
)

func TestSeq(t *testing.T) {
	test(t,
		// --------------
		"test single value", `
- 5`,
		[]any{5},

		// --------------
		"test fail without dash", `
-false`,
		errors.New("unknown number"),

		// --------------
		"test value with newline", `
- 5
`, []any{5},

		// --------------
		"test split line", `
-
  5
`, []any{5},

		// --------------
		"test several values", `
- 5
- 10
- 12`,
		[]any{5, 10, 12},

		// --------------
		"test nested sub sequence", `
- - 5`,
		[]any{[]any{5}},

		// --------------
		"test new line sub sequence", `
-
  - 5
`, []any{[]any{5}},
		// --------------
		"test multiple sub values", `
- -
  - 5
`, []any{[]any{nil, 5}},

		// --------------
		"test nil values", `
-
-
-`,
		[]any{nil, nil, nil},

		// --------------
		"test nil value trailing newline", `
-
-
-
`,
		[]any{nil, nil, nil},

		// --------------
		"test continuing sub sequence ", `
- - 5
- 6`,
		[]any{[]any{5}, 6})
}
