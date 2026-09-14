package educational

import (
	"net/url"
	"strings"
	"unicode/utf8"
)

// maxFieldLength matches the varchar(255) columns the fields are stored in.
const maxFieldLength = 255

type GroupInput struct {
	Name   string
	Hidden bool
}

// GroupPatch changes only the fields that are set.
type GroupPatch struct {
	Name   *string
	Hidden *bool
}

type SubjectInput struct {
	Name string
}

type SubjectObjectInput struct {
	Name    string
	Href    string
	Comment string
	Hidden  bool
}

// SubjectObjectPatch changes only the fields that are set.
type SubjectObjectPatch struct {
	Name    *string
	Href    *string
	Comment *string
	Hidden  *bool
}

// Normalize trims the fields and validates them.
func (in *GroupInput) Normalize() error {
	v := validator{}
	in.Name = v.name("name", in.Name)
	return v.err()
}

func (p *GroupPatch) Normalize() error {
	v := validator{}
	if p.Name != nil {
		name := v.name("name", *p.Name)
		p.Name = &name
	}
	return v.err()
}

func (in *SubjectInput) Normalize() error {
	v := validator{}
	in.Name = v.name("name", in.Name)
	return v.err()
}

func (in *SubjectObjectInput) Normalize() error {
	v := validator{}
	in.Name = v.name("name", in.Name)
	in.Href = v.href("href", in.Href)
	in.Comment = v.text("comment", in.Comment)
	return v.err()
}

func (p *SubjectObjectPatch) Normalize() error {
	v := validator{}
	if p.Name != nil {
		name := v.name("name", *p.Name)
		p.Name = &name
	}
	if p.Href != nil {
		href := v.href("href", *p.Href)
		p.Href = &href
	}
	if p.Comment != nil {
		comment := v.text("comment", *p.Comment)
		p.Comment = &comment
	}
	return v.err()
}

// IsSafeHref reports whether href can be put into a link: an absolute http(s) URL.
func IsSafeHref(href string) bool {
	u, err := url.Parse(href)
	return err == nil && (u.Scheme == "http" || u.Scheme == "https") && u.Host != ""
}

type validator struct {
	fields map[string]string
}

func (v *validator) fail(field, message string) {
	if v.fields == nil {
		v.fields = map[string]string{}
	}
	if _, ok := v.fields[field]; !ok {
		v.fields[field] = message
	}
}

func (v *validator) name(field, value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		v.fail(field, "Укажите название")
	}
	return v.limit(field, value)
}

func (v *validator) text(field, value string) string {
	return v.limit(field, strings.TrimSpace(value))
}

// href accepts an empty value: a subject object may have no link yet.
func (v *validator) href(field, value string) string {
	value = strings.TrimSpace(value)
	if value != "" && !IsSafeHref(value) {
		v.fail(field, "Укажите ссылку, которая начинается с http:// или https://")
	}
	return v.limit(field, value)
}

func (v *validator) limit(field, value string) string {
	if utf8.RuneCountInString(value) > maxFieldLength {
		v.fail(field, "Не больше 255 символов")
	}
	return value
}

func (v *validator) err() error {
	if len(v.fields) == 0 {
		return nil
	}
	return &ValidationError{Fields: v.fields}
}
