// Package markdown renders markdown documents into HTML for pages.
package markdown

import (
	"bytes"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"html/template"
)

// renderer keeps goldmark's safe defaults: raw HTML from the source is omitted and links
// with dangerous schemes (javascript:, vbscript:, ...) get an empty href.
var renderer = goldmark.New(
	goldmark.WithExtensions(extension.GFM),
)

// Render converts markdown (GitHub Flavored) into HTML that is safe to embed into a page.
func Render(source []byte) (template.HTML, error) {
	var buf bytes.Buffer
	if err := renderer.Convert(source, &buf); err != nil {
		return "", err
	}
	// Safe to mark as trusted HTML: the renderer doesn't pass raw HTML through.
	return template.HTML(buf.String()), nil
}
