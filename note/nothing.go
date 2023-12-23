package note

// takes no action in response to the Taker methods
type Nothing struct{}

func (Nothing) BeginCollection(*Context)       {}
func (Nothing) EndCollection()                 {}
func (Nothing) NextTerm()                      {}
func (Nothing) Comment(Type, string) (_ error) { return }
func (Nothing) Resolve() (_ string, _ bool)    { return }
