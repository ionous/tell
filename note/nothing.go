package note

import "strings"

// takes no action in response to the Taker methods
type Nothing struct{}

func (Nothing) BeginCollection(buf *strings.Builder) {}
func (Nothing) EndCollection()                       {}
func (Nothing) NextKey()                             {}
func (Nothing) Comment(Type, string)                 {}
func (Nothing) Resolve() (_ string, _ bool)          { return }
