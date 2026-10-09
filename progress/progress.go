// Package progress tracks which lab works each student of a group has done.
package progress

import (
	"context"
	"errors"
	"sort"
	"strings"

	"github.com/kosipov/students/models"
)

// Mark statuses, in the order a lab work goes through them.
const (
	StatusInProgress = "in_progress"
	StatusSubmitted  = "submitted"
	StatusAccepted   = "accepted"
)

// ValidStatus reports whether the status can be stored. An empty cell has no mark at all.
func ValidStatus(status string) bool {
	switch status {
	case StatusInProgress, StatusSubmitted, StatusAccepted:
		return true
	}
	return false
}

var (
	ErrGroupNotFound   = errors.New("group not found")
	ErrStudentNotFound = errors.New("student not found")
	ErrTaskNotFound    = errors.New("subject object not found")
	// ErrRosterEmpty means the university returned no students for the group's name.
	ErrRosterEmpty = errors.New("roster is empty")
)

// RosterSource loads the roster of a group from the university (the campus site).
type RosterSource interface {
	FetchStudents(ctx context.Context, group string) ([]string, error)
}

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

// ImportResult summarizes a roster import: how many students were found in the source,
// how many were added and how many were skipped as duplicates.
type ImportResult struct {
	Found   int
	Added   int
	Skipped int
}

// SubjectTasks is a subject with the tasks the matrix shows as columns.
type SubjectTasks struct {
	Subject *models.Subject
	Tasks   []models.SubjectObject
}

// StudentMarks is a student with every mark they have in the group.
type StudentMarks struct {
	Student *models.Student
	Marks   []models.Mark
}

// Matrix is the grade sheet of a group: students as rows, tasks of each subject as columns.
type Matrix struct {
	Group    *models.Group
	Subjects []SubjectTasks
	Students []StudentMarks
}

// TaskProgress is one task with the student's mark, for the public page.
type TaskProgress struct {
	Task *models.SubjectObject
	// Mark is nil when the teacher hasn't marked anything yet.
	Mark *models.Mark
}

// SubjectProgress is a subject with the student's marks, for the public page.
type SubjectProgress struct {
	Subject *models.Subject
	Tasks   []TaskProgress
}

// StudentProgress is what a student sees on the public page.
type StudentProgress struct {
	Student  *models.Student
	Group    *models.Group
	Subjects []SubjectProgress
}

// Repository stores students and their marks.
type Repository interface {
	// GetGroup returns the group with subjects and tasks, newest content hidden when onlyVisible.
	GetGroup(ctx context.Context, id int, onlyVisible bool) (*models.Group, error)
	GetStudent(ctx context.Context, id int) (*models.Student, error)
	GetStudents(ctx context.Context, groupId int) ([]models.Student, error)
	CreateStudent(ctx context.Context, student *models.Student) error
	RenameStudent(ctx context.Context, student *models.Student) error
	DeleteStudent(ctx context.Context, student *models.Student) error
	GetMarks(ctx context.Context, studentIds []int) ([]models.Mark, error)
	GetSubjectObject(ctx context.Context, id int) (*models.SubjectObject, error)
	SetMark(ctx context.Context, mark *models.Mark) error
	DeleteMark(ctx context.Context, studentId, subjectObjectId int) error
}

// AdminUseCase manages the roster and the marks.
type AdminUseCase interface {
	// GetMatrix loads the whole grade sheet of a group.
	GetMatrix(ctx context.Context, groupId int) (*Matrix, error)
	AddStudent(ctx context.Context, groupId int, input StudentInput) (*models.Student, error)
	// ImportStudents adds students from a newline-separated list and skips duplicates.
	ImportStudents(ctx context.Context, groupId int, list string) (*ImportResult, error)
	// ImportFromCampus adds students from the university's roster of the group.
	ImportFromCampus(ctx context.Context, groupId int) (*ImportResult, error)
	RenameStudent(ctx context.Context, id int, input StudentInput) (*models.Student, error)
	RemoveStudent(ctx context.Context, id int) error
	// SetMark changes the status; an empty status removes the mark.
	SetMark(ctx context.Context, studentId, taskId int, input MarkInput) error
}

// PublicUseCase shows a student their own progress.
type PublicUseCase interface {
	// GetRoster lists the students of a visible group.
	GetRoster(ctx context.Context, groupId int) (*models.Group, []models.Student, error)
	// GetStudentProgress shows the marks of one student on visible subjects and tasks.
	GetStudentProgress(ctx context.Context, studentId int) (*StudentProgress, error)
}
