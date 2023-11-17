package encode

import (
	"strings"

	"github.com/ionous/tell/runes"
)

// implement comment iter for a standard comment block
func StandardComments(str string) (ret CommentIter) {
	return &cit{next: str}
}

type cit struct {
	curr, next string
}

func (c *cit) Next() (okay bool) {
	if okay = len(c.next) > 0; okay {
		c.curr = c.next
		if i := strings.IndexRune(c.next, runes.Record); i < 0 {
			c.next = ""
		} else {
			c.next = c.next[i+1:]
		}
	}
	return
}

func (c *cit) Entry() Comment {
	// curr := c.curr
	// return Comment{
	// 	//
	// }
	// if i := strings.IndexRune(c.next, runes.CollectionMark); i < 0 {

	// runes.CollectionMark
	panic("split c.curr into parts")
}
