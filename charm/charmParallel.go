package charm

// Run all of the passed states.
// If any return nil, they are dropped;
// If any return Terminal ( ex. via Error() or Finished() )
// the parallel state exits.
func Parallel(name string, rs ...State) State {
	return Self(name, func(self State, r rune) (ret State) {
		var cnt int
	Loop:
		for _, s := range rs {
			switch next := s.NewRune(r); next.(type) {
			case nil:
				// this drops the state from the update list
			case Terminal:
				ret = next
				break Loop
			default:
				rs[cnt] = next
				cnt++
			}
		}
		if cnt > 0 && ret == nil {
			rs = rs[:cnt]
			ret = self
		}
		return
	})
}
