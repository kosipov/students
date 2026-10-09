package progress

import (
	"strings"
	"unicode/utf8"
)

const (
	// MaxNameLength matches the varchar(120) column the name is stored in.
	MaxNameLength     = 120
	MaxCommentLength  = 500
	maxImportStudents = 300
)

type StudentInput struct {
	Name string
}

type MarkInput struct {
	// Status is empty to remove the mark.
	Status  string
	Comment string
}

// Normalize trims the name and validates it.
func (in *StudentInput) Normalize() error {
	in.Name = strings.TrimSpace(in.Name)
	if in.Name == "" {
		return &ValidationError{Fields: map[string]string{"name": "Введите фамилию и имя"}}
	}
	if utf8.RuneCountInString(in.Name) > MaxNameLength {
		return &ValidationError{Fields: map[string]string{"name": "Слишком длинное имя"}}
	}
	return nil
}

// Normalize trims the comment and validates the status.
func (in *MarkInput) Normalize() error {
	in.Status = strings.TrimSpace(in.Status)
	in.Comment = strings.TrimSpace(in.Comment)
	if in.Status != "" && !ValidStatus(in.Status) {
		return &ValidationError{Fields: map[string]string{"status": "Неизвестный статус"}}
	}
	if utf8.RuneCountInString(in.Comment) > MaxCommentLength {
		return &ValidationError{Fields: map[string]string{"comment": "Слишком длинный комментарий"}}
	}
	return nil
}

// ParseImportList splits a pasted roster into names, skipping empty lines.
func ParseImportList(list string) ([]string, error) {
	names := make([]string, 0, 32)
	for _, line := range strings.Split(list, "\n") {
		name := strings.TrimSpace(line)
		if name != "" {
			names = append(names, name)
		}
	}
	if len(names) == 0 {
		return nil, &ValidationError{Fields: map[string]string{"names": "Вставьте список студентов, по одному на строку"}}
	}
	if len(names) > maxImportStudents {
		return nil, &ValidationError{Fields: map[string]string{"names": "Слишком много строк, разбейте список на части"}}
	}
	for _, name := range names {
		if utf8.RuneCountInString(name) > MaxNameLength {
			return nil, &ValidationError{Fields: map[string]string{"names": "В списке есть слишком длинная строка"}}
		}
	}
	return names, nil
}
