package encode

import (
	"fmt"
	r "reflect"
	"strings"

	"github.com/ionous/tell/runes"
)

// an implementation of CommentFactory that walks the passed comments
func Comments(els []Comment) CommentIter {
	return &comments{next: els}
}

// an implementation of CommentFactory which generates no comments.
func DiscardComments(r.Value) (CommentIter, error) {
	return nil, nil
}

// implement CommentFactory.
// expects that the value  expecting an interface type with an underlying string
// containing a standard tell comment block
func CommentBlock(v r.Value) (ret CommentIter, err error) {
	if str, e := ExtractString(v); e != nil {
		err = fmt.Errorf("comment factory %s", e)
	} else {
		ret = &cit{rest: str}
	}
	return
}

// a helper which, given a reflected value with an underlying string value
// returns that string. ( for example, from `var comment any = "string"` )
func ExtractString(v r.Value) (ret string, err error) {
	if k := v.Kind(); k != r.Interface {
		err = fmt.Errorf("expected an interface value; got %s(%s)", k, v.Type())
	} else if el := v.Elem(); el.Kind() != r.String {
		err = fmt.Errorf("expected an underlying string; got %s(%s)", el.Kind(), el.Type())
	} else {
		ret = el.String()
	}
	return
}

type comments struct {
	curr Comment
	next []Comment
}

func (s *comments) Next() (okay bool) {
	if okay = len(s.next) > 0; !okay {
		s.curr = Comment{}
	} else {
		s.curr, s.next = s.next[0], s.next[1:]
	}
	return
}

func (s *comments) GetComment() Comment {
	return s.curr
}

type cit struct {
	curr, rest string
}

func (c *cit) Next() (okay bool) {
	if okay = len(c.rest) > 0; okay {
		if i := strings.IndexRune(c.rest, runes.NextTerm); i < 0 {
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
