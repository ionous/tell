package charm

// a next state indicating an terminal error
func Error(e error) State {
	return Terminal{err: e}
}

// represents both an error state an non-error final stat
type Terminal struct {
	err error
}

// returns itself forever
func (e Terminal) NewRune(r rune) (ret State) {
	return e
}

func (e Terminal) Unwrap() error {
	return e.err
}

func (e Terminal) Error() string {
	return e.err.Error()
}

func (e Terminal) String() (ret string) {
	if e.err != nil {
		ret = "error state"
	} else {
		ret = "terminal state"
	}
	return
}
