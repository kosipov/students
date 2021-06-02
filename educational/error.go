package educational

import "errors"

var (
	ErrSubjectNotFound = errors.New("educational not found")
	ErrGroupNotFound   = errors.New("group not found")
)
