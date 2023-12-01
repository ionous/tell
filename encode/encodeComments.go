package encode

import (
	"strings"

	"github.com/ionous/tell/runes"
)

// // implement comment iter for a standard comment block
func makeComments(str string) (ret cit) {
	return cit{rest: str}
}

type cit struct {
	curr, rest string
}

func (c *cit) Next() (okay bool) {
	if okay = len(c.rest) > 0; okay {
		if i := strings.IndexRune(c.rest, runes.NextRecord); i < 0 {
			c.curr = c.rest
			c.rest = ""
		} else {
			c.curr, c.rest = c.rest[:i], c.rest[i+1:]
		}
	}
	return
}

func (c *cit) GetComment() Comment {
	if curr := c.curr; len(curr) > 0 {


	}
	return Comment{}
}
