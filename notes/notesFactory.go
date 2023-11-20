package notes

func KeepComments() Commentator {
	return new(Builder)
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
