package charm

import (
	"errors"
	"fmt"
)

// a next state indicating an error that caused the state machine to exit
func Error(e error) State {
	return Terminal{err: e}
}

// a next state indicating an expected termination.
// ( for example, to absorb runes.Eof and end a state gracefully.
// | whereas returning nil from a state would consider the Eof unhandled
// | and trigger attempts by chained states to handle the Eof themselves. )
func Finished() State {
	return Terminal{err: errFinished}
}

var errFinished = errors.New("finished")

// acts as both an error and a state
type Terminal struct {
	err error
}

func (e Terminal) Finished() bool {
	return e.err == errFinished
}

// returns itself forever
func (e Terminal) NewRune(r rune) (ret State) {
	return e
}

// implements string for printing states
func (e Terminal) String() string {
	return fmt.Sprintf("terminal state: %s", e.Error())
}

// access the underlying error
func (e Terminal) Unwrap() error {
	return e.err
}

// terminal implements error
func (e Terminal) Error() string {
	return e.err.Error()
}
