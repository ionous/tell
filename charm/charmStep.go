package charm

// RunStep - run a sequence of of two states
// sending the current rune to the first state immediately
// see also: Step()
func RunStep(r rune, child, parent State) State {
	return Step(child, parent).NewRune(r)
}

// Step - construct a sequence of two states.
// If the next rune is not handled by the first state or any of its returned states,
// the rune is handed to the second state.
// this acts similar to a parent-child statechart.
func Step(child, parent State) State {
	container := StateName(parent)
	return &chainParser{container, child, parent}
}

// use the first of whichever of the passed states respond to the next rune.
// step puts states into a child/parent aggregation
// first is more like sibling states.
func FirstOf(name string, several ...State) State {
	return Statement(name, func(q rune) (ret State) {
		for _, n := range several {
			if next := n.NewRune(q); next != nil {
				ret = next
				break
			}
		}
		return
	})
}

func Jump(container string, target State) State {
	return jumpState{container, target}
}

type jumpState struct {
	container string
	State
}

// For use in Step() to run an action after the first step completes.
func OnExit(name string, onExit func()) State {
	return Statement("on exit", func(rune) (none State) {
		onExit()
		return
	})
}

type chainParser struct {
	container  string
	next, last State
}

func (p *chainParser) String() string {
	return StateName(p.next) + "(chain: " + StateName(p.last) + ")"
}

// runs the each state ( and any of their returned states ) to completion
func (p *chainParser) NewRune(r rune) (ret State) {
	if next := p.next.NewRune(r); next == nil {
		// out of next states, run the original last state
		ret = p.last.NewRune(r)
	} else if err, ok := next.(Terminal); ok {
		// if the next state is an error state, return it now.
		ret = err
	} else if jump, ok := next.(jumpState); ok {
		// if we see a jump state, our chain is dead:
		// keep ripping off chains until we've found the targeted container
		// any chains above us will see a normal state after that
		if jump.container == p.container {
			ret = jump.State
		} else {
			ret = next
		}
	} else {
		// remember the new next state, and
		// return *this* to keep stepping towards last.
		ret, p.next = p, next
	}
	return
}
