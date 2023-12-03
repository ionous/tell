package encode

import (
	"fmt"
	r "reflect"
	"strings"

	"github.com/ionous/tell/runes"
)

func DiscardComments(r.Value) (CommentIter, error) {
	return nil, nil
}

// implement CommentFactory expecting an interface type with an underlying string
// containing a standard tell comment block
func CommentBlock(v r.Value) (ret CommentIter, err error) {
	if str, e := extractString(v); e != nil {
		err = fmt.Errorf("comment factory %s", e)
	} else {
		ret = &cit{rest: str}
	}
	return
}

func extractString(v r.Value) (ret string, err error) {
	if k := v.Kind(); k != r.Interface {
		err = fmt.Errorf("expected an interface value; got %s(%s)", k, v.Type())
	} else if el := v.Elem(); el.Kind() != r.String {
		err = fmt.Errorf("expected an underlying string; got %s(%s)", el.Kind(), el.Type())
	} else {
		ret = el.String()
	}
	return
}

type emptyComments struct{}

func (emptyComments) Next() (_ bool)          { return }
func (emptyComments) GetComment() (_ Comment) { return }

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
		// FIX: not implemented!
	}
	return Comment{}
}
