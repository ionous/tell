package notes

// import (
// 	"github.com/ionous/tell/charm"
// 	"github.com/ionous/tell/runes"
// )

// // adds separators between records
// // meant to be run in parallel
// type prefixDecoder struct {
// 	*context
// }

// func keyWatcher(ctx *context) charm.State {
// 	d := prefixDecoder{ctx}
// 	return d.watch()
// }

// func (d *prefixDecoder) watch() charm.State {
// 	return charm.Step(d.watchTerm(), d.watchKey())
// }

// // immediately after a key:
// func (d *prefixDecoder) watchTerm() charm.State {
// 	return charm.Statement("watchTerm", func(q rune) (ret charm.State) {
// 		switch q {
// 		case runeValue: // got a value immediately after the key
// 			ret = d.dual()
// 		case runes.Hash:
// 			ret = charm.Jump("watchKey", d.needBothMarkers())
// 		case runes.Newline:
// 			ret = d.keyless()
// 		case runes.Hash:
// 			ret = invalidRune("watchTerm", q)
// 		}
// 		return
// 	})
// }

// func (d *prefixDecoder) watchKey() charm.State {
// 	return charm.Statement("watchKey", func(q rune) (ret charm.State) {
// 		switch q {
// 		case runeKey:
// 			d.out.terms++
// 			ret = d.watch()
// 		}
// 		return
// 	})
// }

// func (d *prefixDecoder) revisedWatch() charm.State {
// 	return charm.Statement("revisedWatch", func(q rune) (ret charm.State) {
// 		d.out.WriteRune(runes.CollectionMark)
// 		return charm.RunState(q, d.watch())
// 	})
// }

// // write the first mark, then as soon as there's a key or value, write the next one.
// func (d *prefixDecoder) needBothMarkers() charm.State {
// 	d.out.WriteRune(runes.CollectionMark)
// 	return charm.Self("one", func(self charm.State, q rune) (ret charm.State) {
// 		switch q {
// 		case runeKey:
// 			d.out.WriteRune(runes.CollectionMark)
// 			ret = d.watch()
// 		case runeValue:
// 			d.out.WriteRune(runes.CollectionMark)
// 			ret = d.waitForKey()
// 		default:
// 			ret = self // keep looping
// 		}
// 		return
// 	})
// }

// // everything's done, just waiting for the key to loop
// func (d *prefixDecoder) waitForKey() charm.State {
// 	return charm.Statement("waitForKey", func(q rune) (ret charm.State) {
// 		switch q {
// 		case runeKey:
// 			ret = d.watch()
// 		}
// 		return
// 	})
// }

// // there was explicitly no key comment
// // however if there are some element headers in the padding
// // and there is no sub-collection ( instead the doc specifies a scalar )
// // then the element headers will be revised into key comments.
// func (d *prefixDecoder) keyless() charm.State {
// 	return charm.Statement("keyless", func(q rune) (ret charm.State) {
// 		switch q {
// 		case runes.Hash:
// 		}
// 		return
// 	})
// }

// func (d *prefixDecoder) elComment() charm.State {
// 	return charm.Statement("elComment", func(q rune) (ret charm.State) {
// 		switch q {
// 		}
// 		return
// 	})
// }

// func (d *prefixDecoder) revised() charm.State {
// 	return charm.Statement("revised", func(q rune) (ret charm.State) {
// 		switch q {
// 		}
// 		return
// 	})
// }

// func (d *prefixDecoder) sandwich() charm.State {
// 	return charm.Statement("sandwich", func(q rune) (ret charm.State) {
// 		switch q {
// 		}
// 		return
// 	})
// }

// func (d *prefixDecoder) dual() charm.State {
// 	return charm.Statement("dual", func(q rune) (ret charm.State) {
// 		switch q {
// 		}
// 		return
// 	})
// }

// func (d *prefixDecoder) trailing() charm.State {
// 	return charm.Statement("trailing", func(q rune) (ret charm.State) {
// 		switch q {
// 		}
// 		return
// 	})
// }

// func (d *prefixDecoder) tail() charm.State {
// 	return charm.Statement("tail", func(q rune) (ret charm.State) {
// 		switch q {
// 		}
// 		return
// 	})
// }
