package notes

func KeepComments() *Builder {
	return new(Builder).init()
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
