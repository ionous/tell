package notes

import (
	"fmt"
	"io"
	"strings"

	"github.com/ionous/tell/charm"
	"github.com/ionous/tell/runes"
)

func KeepComments() *Builder {
	panic("xxx")
	// b := &Builder{
	// 	ctx: context{
	// 		out: new(strings.Builder),
	// 	},
	// }
	// b.state = docStart{
	// 	ctx:           &b.ctx,
	// 	newCollection: nil,
	// 	inlineScalar:  nil,
	// }
	// return b
}

func DiscardComments() Commentator {
	return Nothing{}
}

func NewCommentator(keepComments bool) (ret Commentator) {
	if keepComments {
		ret = KeepComments()
	} else {
		ret = DiscardComments()
	}
	return
}

type Builder struct {
	state charm.State
}

func build(state charm.State) Builder {
	return Builder{state}
}

type context struct {
	out   *strings.Builder
	buf   strings.Builder
	stack stack
}

func newContext() *context {
	return &context{out: new(strings.Builder)}
}

func (ctx *context) GetComments() string {
	if ctx.buf.Len() > 0 {
		ctx.flush()
	}
	str := ctx.out.String()
	ctx.out = nil
	return str
}

// write passed runes, and then the buffer, to out
func (ctx *context) flush(qs ...rune) {
	writeRunes(ctx.out, qs...)
	if str := ctx.buf.String(); len(str) > 0 {
		io.WriteString(ctx.out, str)
		ctx.buf.Reset()
	}
}

func (b *Builder) OnParagraph() Commentator {
	b.send(runeParagraph)
	return b
}
func (b *Builder) OnKeyDecoded() Commentator {
	b.send(runeKey)
	return b
}
func (b *Builder) OnScalarValue() Commentator {
	b.send(runeValue)
	return b
}
func (b *Builder) GetComments() string {
	// b.state =b.send(runeEnd)
	// // pop
	// return ""
	panic("fix")
}
func (b *Builder) WriteRune(q rune) (_ int, _ error) {
	b.send(q)
	return
}

func (b *Builder) send(q rune) {
	if next := b.state.NewRune(q); next == nil {
		// no states left to parse remaining input
		err := fmt.Errorf("unknown handled rune %q in %q", q, charm.StateName(b.state))
		panic(err)
	} else if es, ok := next.(charm.Terminal); ok {
		err := fmt.Errorf("error for rune %q in %q %w", q, charm.StateName(b.state), es)
		panic(err)
	} else {
		b.state = next
	}
}

type pendingBlock struct {
	strings.Builder
}

type makeState func() charm.State

// a state which creates the passed state to handle a rune
func kickOff(m makeState) charm.State {
	return charm.Statement("kickOff", func(q rune) charm.State {
		return charm.RunState(q, m())
	})
}

func invalidRune(name string, q rune) error {
	return fmt.Errorf("unexpected rune %q during %s", q, name)
}

// these runes can be used by authors in comments
// includes htab because authors should be permitted to comment out literals
// and literals can include actual tabs.
// author escape sequences in a comment, ex. an escaped tab \t,
// are two separate and individually permitted runes.
func friendly(q rune) bool {
	return q == runes.HTab || q >= runes.Space
}

//
const (
	runeEnd       = '\f'
	runeParagraph = '\a'
	runeValue     = '\v'
	runeKey       = '\r'
)

func writeRunes(w RuneWriter, qs ...rune) {
	for _, q := range qs {
		w.WriteRune(q)
	}
}

func readLine(name string, w RuneWriter, eol func() charm.State) (ret charm.State) {
	return charm.Self(name, func(self charm.State, q rune) (ret charm.State) {
		switch {
		case q == runes.Newline:
			ret = eol()
		case friendly(q):
			w.WriteRune(q)
			ret = self
		default:
			ret = charm.Error(invalidRune(name, q))
		}
		return
	})
}

func nest(w RuneWriter) {
	writeRunes(w, runes.Newline, runes.HTab, runes.Hash)
}
