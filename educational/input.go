package educational

import (
	"net/url"
	"strings"
	"unicode/utf8"
)

const (
	// maxFieldLength matches the varchar(255) columns the fields are stored in.
	maxFieldLength = 255
	// MaxCategoryLength fits the varchar(100) column with room to spare for the unique index.
	MaxCategoryLength = 50
	MaxCategories     = 10
	MaxPasswordLength = 100
)

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
	Name       string
	Href       string
	Comment    string
	Hidden     bool
	Categories []string
	// Password protects the task; empty means no password.
	Password string
}

// SubjectObjectPatch changes only the fields that are set.
type SubjectObjectPatch struct {
	Name    *string
	Href    *string
	Comment *string
	Hidden  *bool
	// Categories replaces all categories of the subject object when set.
	Categories *[]string
	// Password replaces the password when set; an empty one removes it.
	Password *string
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
	in.Categories = v.categories("categories", in.Categories)
	in.Password = v.password("password", in.Password)
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
	if p.Categories != nil {
		categories := v.categories("categories", *p.Categories)
		p.Categories = &categories
	}
	if p.Password != nil {
		password := v.password("password", *p.Password)
		p.Password = &password
	}
	return v.err()
}

// CategoryKey identifies a category regardless of letter case and "ё", so "Курсовые" and "курсовые"
// are one category.
func CategoryKey(name string) string {
	return strings.ReplaceAll(strings.ToLower(name), "ё", "е")
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

// categories collapses spaces, drops empty names and repeats (in any letter case) and keeps the order.
func (v *validator) categories(field string, values []string) []string {
	result := make([]string, 0, len(values))
	seen := make(map[string]bool, len(values))
	for _, value := range values {
		name := strings.Join(strings.Fields(value), " ")
		if name == "" {
			continue
		}
		if utf8.RuneCountInString(name) > MaxCategoryLength {
			v.fail(field, "Название категории — не больше 50 символов")
			continue
		}
		key := CategoryKey(name)
		if seen[key] {
			continue
		}
		seen[key] = true
		result = append(result, name)
	}
	if len(result) > MaxCategories {
		v.fail(field, "Не больше 10 категорий у одного задания")
	}
	return result
}

// password is trimmed: students type it by hand, and a space nobody sees must not lock them out.
func (v *validator) password(field, value string) string {
	value = strings.TrimSpace(value)
	if utf8.RuneCountInString(value) > MaxPasswordLength {
		v.fail(field, "Пароль — не больше 100 символов")
	}
	return value
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
