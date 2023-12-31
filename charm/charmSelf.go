package charm

// Self are States which pass a function pointer as the first argument to the state callback.
// This allows closures to return themselves.
// For example, the following state returns itself forever:
//   var recursive  Self = func(self State, r rune) State { return self }
func Self(name string, closure func(State, rune) State) State {
	return &selfState{name, closure}
}

type selfState struct {
	name    string
	closure func(State, rune) State
}

func (s *selfState) String() string {
	return s.name
}

// NewRune implements State by calling the underlying function.
func (s *selfState) NewRune(r rune) State {
	return s.closure(s, r)
}
