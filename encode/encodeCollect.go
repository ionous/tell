package encode

type Mapper interface {
	TellMapping() MappingIter
	TellComments() CommentIter
}

type Sequencer interface {
	TellSequence() SequenceIter
	TellComments() CommentIter
}

type MappingIter interface {
	Next() bool
	GetKey() string
	GetValue() any
}

type SequenceIter interface {
	Next() bool
	GetValue() any
	// tbd: maybe an optional GetReflectedValue
	// so we dont have to unpack/repack values otherwise
}

type Comment struct {
	Header []string
	Inline []string
	Footer []string
}

// comment access for collections
type CommentIter interface {
	Next() bool
	Entry() Comment
}
