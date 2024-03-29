package encode

import (
	"fmt"
	r "reflect"
	"strings"

	"github.com/ionous/tell/runes"
)

// turns a value representing one or more comments into a comment iterator.
// the encoder uses the iterator to generate comments for collections.
type Commenting func(r.Value) (Comments, error)

// an implementation of Commenting that walks the passed slice.
func CommentSlice(els []Comment) Comments {
	return &commentSlice{next: els}
}

// a nil commenting implementation for readability
var NoComments Commenting = nil

// implies that there are comments;
// the encoder should simply skip them.
func DiscardComments(r.Value) (Comments, error) {
	return nil, nil
}

// expects that the value is a kind of string
// containing a standard tell comment block
func CommentBlock(v r.Value) (ret Comments, err error) {
	if str, e := ExtractString(v); e != nil {
		err = fmt.Errorf("comment factory %s", e)
	} else {
		ret = &commentBlock{rest: str}
	}
	return
}

// a helper which, given a reflected string value returns that string.
func ExtractString(el r.Value) (ret string, err error) {
	switch k := el.Kind(); k {
	case r.String:
		ret = el.String()
	// when GetReflectedValue isnt implemented;
	// strings can be wrapped by interfaces due to GetValue()
	// which returns an `any`
	case r.Interface:
		ret, err = ExtractString(el.Elem())
	default:
		err = fmt.Errorf("expected an underlying string; got %s(%s)", k, el.Type())
	}
	return
}

// forever. nothing.
type noComments struct{}

func (s noComments) Next() (_ bool)          { return }
func (s noComments) GetComment() (_ Comment) { return }

// iterate over pre-built comments
type commentSlice struct {
	curr Comment
	next []Comment
}

func (s *commentSlice) Next() (okay bool) {
	if okay = len(s.next) > 0; !okay {
		s.curr = Comment{}
	} else {
		s.curr, s.next = s.next[0], s.next[1:]
	}
	return
}

func (s *commentSlice) GetComment() Comment {
	return s.curr
}

// walk a comment block string
type commentBlock struct {
	curr, rest string
}

func (c *commentBlock) Next() (okay bool) {
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

func (c *commentBlock) GetComment() (ret Comment) {
	if curr := c.curr; len(curr) > 0 {
		parts := strings.Split(curr, string(runes.KeyValue))
		for i, p := range parts {
			if len(p) > 0 {
				lines := strings.Split(p, string(runes.Newline))
				switch i {
				case 0:
					ret.Header = lines
				case 1:
					ret.Prefix = lines
				case 2:
					ret.Suffix = lines
					// error?
				}
			}
		}
	}
	return
}
