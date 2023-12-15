package note

import "strings"

// a comment block
type Taker interface {
	BeginCollection(*strings.Builder)
	Comment(Type, string)
	NextKey()
	EndCollection()
	// return true
	Resolve() (ret string, okay bool)
}
