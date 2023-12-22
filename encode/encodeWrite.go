package encode

import "github.com/ionous/tell/runes"

// assumes a length of at least one
func fixedWrite(tab *TabWriter, suffix []string) {
	prevDepth := tab.depth
	first, rest := suffix[0], suffix[1:]
	if inline := len(first) > 0; !inline {
		tab.Softline()
		tab.depth += 2
	} else {
		tab.WriteRune(runes.Space)
		tab.depth = tab.xpos
		tab.writeLine(first)
	}
	tab.writeLines(rest)
	tab.depth = prevDepth
}
