package educational

import (
	"context"
)

// ContentFetcher downloads documents that subject objects link to.
type ContentFetcher interface {
	// Supports reports whether href points to a storage the fetcher can download from.
	// It does not make network requests.
	Supports(href string) bool
	// Fetch downloads the document behind href. When etag is not empty and matches
	// the current version, ErrContentNotModified is returned instead of the body.
	// ErrContentUnsupported is returned when the document is not markdown.
	Fetch(ctx context.Context, href string, etag string) (*Content, error)
}

type Content struct {
	Body []byte
	ETag string
}
