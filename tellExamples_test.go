package tell_test

import (
	"fmt"

	"github.com/ionous/tell"
)

// Read a tell document.
func ExampleUnmarshal() {
	var out any
	const msg = `- Hello: "\U0001F30F"`
	if e := tell.Unmarshal([]byte(msg), &out); e != nil {
		panic(e)
	} else {
		fmt.Printf("%#v", out)
	}
	// Output:
	// []interface {}{map[string]interface {}{"Hello:":"🌏"}}
}

// Write a tell document.
func ExampleMarshal() {
	m := map[string]any{
		"Tell":           "A yaml-like text format.",
		"What It Is":     "A way of describing data...",
		"What It Is Not": "A subset of yaml.",
	}
	if out, e := tell.Marshal(m); e != nil {
		panic(e)
	} else {
		fmt.Println(string(out))
	}
	// Output:
	// Tell: "A yaml-like text format."
	// What It Is: "A way of describing data..."
	// What It Is Not: "A subset of yaml."
}
