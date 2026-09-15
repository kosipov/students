package educational

import (
	"errors"
	"sort"
	"strings"
)

var (
	ErrSubjectNotFound = errors.New("subject not found")
	ErrGroupNotFound   = errors.New("group not found")

	ErrSubjectObjectNotFound = errors.New("subject object not found")
	// ErrSubjectObjectNotDocument means the subject object links to something
	// that can't be shown on the site, so the link should be opened as is.
	ErrSubjectObjectNotDocument = errors.New("subject object does not link to a document")

	// ErrTaskLocked means the task has a password the student hasn't entered yet.
	ErrTaskLocked    = errors.New("task is locked")
	ErrWrongPassword = errors.New("wrong task password")

	ErrContentNotModified = errors.New("content not modified")
	ErrContentUnsupported = errors.New("content type is not supported")
)

// ValidationError lists invalid input fields with messages that can be shown to the user.
type ValidationError struct {
	Fields map[string]string
}

func (e *ValidationError) Error() string {
	fields := make([]string, 0, len(e.Fields))
	for field := range e.Fields {
		fields = append(fields, field)
	}
	sort.Strings(fields)
	return "invalid fields: " + strings.Join(fields, ", ")
}
