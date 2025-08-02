package errorz

import "errors"

var (
	ErrTopicNotExists = errors.New("topic doesn't exist")
)
