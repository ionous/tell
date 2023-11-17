package tell_test

import (
	"fmt"

	"github.com/ionous/tell"
)

func ExampleUnmarshal() {
	// Unmarshal is the simplest interface
	// using package decode gives more control
	var b bool
	if e := tell.Unmarshal([]byte(`true`), &b); e != nil {
		panic(e)
	} else {
		fmt.Println(b)
	}
	// Output: true
}

func ExampleMarshal() {
	b := true
	if out, e := tell.Marshal(b); e != nil {
		panic(e)
	} else {
		fmt.Println(string(out))
	}
	// Output: true
}
