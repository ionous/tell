package notes

func KeepComments() CommentResolver {
	return newNotes()
}

func DiscardComments() CommentResolver {
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

func newNotes() *commentResolver {
	ctx := newContext()
	b := build(newDocument(ctx))
	return &commentResolver{ctx, b}
}

type commentResolver struct {
	ctx *context
	Builder
}

func (p *commentResolver) GetComments() string {
	return p.ctx.res
}

func (p *commentResolver) GetAllComments() []string {
	return p.Builder.GetAllComments(p.ctx)
}
