package educational

import "errors"

var (
	ErrSubjectNotFound = errors.New("educational not found")
	ErrGroupNotFound   = errors.New("group not found")

	ErrSubjectObjectNotFound = errors.New("subject object not found")
	// ErrSubjectObjectNotDocument means the subject object links to something
	// that can't be shown on the site, so the link should be opened as is.
	ErrSubjectObjectNotDocument = errors.New("subject object does not link to a document")

	ErrContentNotModified = errors.New("content not modified")
	ErrContentUnsupported = errors.New("content type is not supported")
)
