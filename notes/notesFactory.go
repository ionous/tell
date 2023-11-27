package notes

func KeepComments() *Builder {
	panic("xxx")
	// b := &Builder{
	// 	ctx: context{
	// 		out: new(strings.Builder),
	// 	},
	// }
	// b.state = docStart{
	// 	ctx:           b.OnNestedComment().ctx,
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
