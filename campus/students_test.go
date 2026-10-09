package campus

import (
	"errors"
	"os"
	"testing"
)

func TestParseStudents(t *testing.T) {
	answer, err := os.ReadFile("testdata/students.json")
	if err != nil {
		t.Fatalf("read fixture: %s", err)
	}

	names, err := ParseStudents(string(answer))
	if err != nil {
		t.Fatalf("ParseStudents: %s", err)
	}
	if len(names) != 23 {
		t.Fatalf("expected 23 students, got %d", len(names))
	}
	if names[0] != "Галигузова Ирина Юрьевна" {
		t.Fatalf("unexpected first student: %q", names[0])
	}
}

func TestParseStudentsUnknownGroup(t *testing.T) {
	names, err := ParseStudents(`{"student":""}`)
	if err != nil {
		t.Fatalf("ParseStudents: %s", err)
	}
	if len(names) != 0 {
		t.Fatalf("expected no students, got %v", names)
	}
}

func TestParseStudentsNotJSON(t *testing.T) {
	_, err := ParseStudents("<html>404</html>")
	if !errors.Is(err, ErrUnexpectedRoster) {
		t.Fatalf("expected ErrUnexpectedRoster, got %v", err)
	}
}
