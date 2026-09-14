package educational

import (
	"errors"
	"strings"
	"testing"
)

func TestSubjectObjectInputNormalize(t *testing.T) {
	input := SubjectObjectInput{Name: "  Лабораторная  ", Href: " https://1drv.ms/t/c/1/abc ", Comment: " Пояснение "}
	if err := input.Normalize(); err != nil {
		t.Fatalf("Normalize() error = %v", err)
	}
	if input.Name != "Лабораторная" || input.Href != "https://1drv.ms/t/c/1/abc" || input.Comment != "Пояснение" {
		t.Errorf("Normalize() = %+v, want trimmed fields", input)
	}

	withoutLink := SubjectObjectInput{Name: "Задание"}
	if err := withoutLink.Normalize(); err != nil {
		t.Errorf("Normalize() without link error = %v, want the link to be optional", err)
	}
}

func TestSubjectObjectInputValidation(t *testing.T) {
	cases := map[string]struct {
		input SubjectObjectInput
		field string
	}{
		"empty name":        {SubjectObjectInput{Name: "  "}, "name"},
		"long name":         {SubjectObjectInput{Name: strings.Repeat("я", 256)}, "name"},
		"javascript link":   {SubjectObjectInput{Name: "Задание", Href: "javascript:alert(1)"}, "href"},
		"relative link":     {SubjectObjectInput{Name: "Задание", Href: "/tasks/1"}, "href"},
		"link without host": {SubjectObjectInput{Name: "Задание", Href: "https://"}, "href"},
		"long comment":      {SubjectObjectInput{Name: "Задание", Comment: strings.Repeat("я", 256)}, "comment"},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			err := tc.input.Normalize()
			var validationErr *ValidationError
			if !errors.As(err, &validationErr) || validationErr.Fields[tc.field] == "" {
				t.Fatalf("Normalize() error = %v, want %s error", err, tc.field)
			}
		})
	}

	maxLength := SubjectObjectInput{Name: strings.Repeat("я", 255)}
	if err := maxLength.Normalize(); err != nil {
		t.Errorf("Normalize() with 255 characters error = %v, want the limit to count characters, not bytes", err)
	}
}

func TestPatchNormalizeSkipsUnsetFields(t *testing.T) {
	var patch SubjectObjectPatch
	if err := patch.Normalize(); err != nil {
		t.Errorf("Normalize() of empty patch error = %v", err)
	}

	name := " Группа "
	groupPatch := GroupPatch{Name: &name}
	if err := groupPatch.Normalize(); err != nil || *groupPatch.Name != "Группа" {
		t.Errorf("GroupPatch.Normalize() = %q, %v; want trimmed name", *groupPatch.Name, err)
	}
}
