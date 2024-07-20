package charm

import (
	"errors"
	"fmt"
)

// returned from Parser
type UnhandledRune rune

func (u UnhandledRune) Error() string {
	return fmt.Sprintf("unhanded rune %v", rune(u))
}

// Used by states to wrap an error in a Terminal state.
// This is the only way for states to return an error.
// To stop processing, but return no error: see Finished()
func Error(e error) State {
	return Terminal{err: e}
}

// A next state indicating an expected termination.
// This is used to absorb runes.Eof and end a state gracefully because
// returning nil on Eof will trigger attempts by chained states to handle the Eof themselves.
func Finished() State {
	return Terminal{err: ErrFinished}
}

// this provided an alternative to the terminal state holding nil.
// ( to avoid testing its error when dereferencing )
var ErrFinished = errors.New("finished")

// acts as both an error and a state
type Terminal struct {
	err error
}

// true if this was an expected termination.
// see also: the package level Finished().
func (e Terminal) Finished() bool {
	return e.err == ErrFinished
}

// returns itself forever.
func (e Terminal) NewRune(r rune) (ret State) {
	return e
}

// implements string for printing states.
func (e Terminal) String() string {
	return fmt.Sprintf("terminal state: %s", e.Error())
}

// access the underlying error.
func (e Terminal) Unwrap() error {
	return e.err
}

// terminal implements error.
// returns the string of the wrapped error.
func (e Terminal) Error() string {
	return e.err.Error()
}
