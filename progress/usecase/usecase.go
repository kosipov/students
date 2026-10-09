package usecase

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jinzhu/gorm"
	"github.com/kosipov/students/models"
	"github.com/kosipov/students/progress"
)

type UseCase struct {
	repo   progress.Repository
	roster progress.RosterSource
}

func NewUseCase(repo progress.Repository, roster progress.RosterSource) *UseCase {
	return &UseCase{repo: repo, roster: roster}
}

// notFound replaces gorm's "record not found" with the domain error.
func notFound(err, domain error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return domain
	}
	return err
}

func (u *UseCase) GetMatrix(ctx context.Context, groupId int) (*progress.Matrix, error) {
	group, err := u.repo.GetGroup(ctx, groupId, false)
	if err != nil {
		return nil, notFound(err, progress.ErrGroupNotFound)
	}
	students, err := u.repo.GetStudents(ctx, groupId)
	if err != nil {
		return nil, err
	}
	marks, err := u.repo.GetMarks(ctx, studentIds(students))
	if err != nil {
		return nil, err
	}

	marksByStudent := make(map[int][]models.Mark)
	for i := range marks {
		marksByStudent[marks[i].StudentId] = append(marksByStudent[marks[i].StudentId], marks[i])
	}

	matrix := &progress.Matrix{
		Group:    group,
		Subjects: subjectTasks(group),
		Students: make([]progress.StudentMarks, 0, len(students)),
	}
	for i := range students {
		matrix.Students = append(matrix.Students, progress.StudentMarks{
			Student: &students[i],
			Marks:   marksByStudent[students[i].ID],
		})
	}
	return matrix, nil
}

func (u *UseCase) AddStudent(ctx context.Context, groupId int, input progress.StudentInput) (*models.Student, error) {
	if err := input.Normalize(); err != nil {
		return nil, err
	}
	if _, err := u.repo.GetGroup(ctx, groupId, false); err != nil {
		return nil, notFound(err, progress.ErrGroupNotFound)
	}
	student := &models.Student{Name: input.Name, GroupId: groupId}
	if err := u.repo.CreateStudent(ctx, student); err != nil {
		return nil, err
	}
	return student, nil
}

func (u *UseCase) ImportStudents(ctx context.Context, groupId int, list string) (*progress.ImportResult, error) {
	names, err := progress.ParseImportList(list)
	if err != nil {
		return nil, err
	}
	return u.importNames(ctx, groupId, names)
}

func (u *UseCase) ImportFromCampus(ctx context.Context, groupId int) (*progress.ImportResult, error) {
	group, err := u.repo.GetGroup(ctx, groupId, false)
	if err != nil {
		return nil, notFound(err, progress.ErrGroupNotFound)
	}
	names, err := u.roster.FetchStudents(ctx, group.GroupName)
	if err != nil {
		return nil, fmt.Errorf("campus roster: %w", err)
	}
	if len(names) == 0 {
		return nil, progress.ErrRosterEmpty
	}
	return u.importNames(ctx, groupId, names)
}

// importNames adds the students who are not in the group yet; duplicates match case-insensitively:
// "иванов иван" and "Иванов Иван" are the same student.
func (u *UseCase) importNames(ctx context.Context, groupId int, names []string) (*progress.ImportResult, error) {
	if _, err := u.repo.GetGroup(ctx, groupId, false); err != nil {
		return nil, notFound(err, progress.ErrGroupNotFound)
	}
	existing, err := u.repo.GetStudents(ctx, groupId)
	if err != nil {
		return nil, err
	}

	taken := make(map[string]bool, len(existing)+len(names))
	for i := range existing {
		taken[strings.ToLower(existing[i].Name)] = true
	}

	result := &progress.ImportResult{Found: len(names)}
	for _, name := range names {
		key := strings.ToLower(name)
		if taken[key] {
			result.Skipped++
			continue
		}
		taken[key] = true
		if err := u.repo.CreateStudent(ctx, &models.Student{Name: name, GroupId: groupId}); err != nil {
			return nil, err
		}
		result.Added++
	}
	return result, nil
}

func (u *UseCase) RenameStudent(ctx context.Context, id int, input progress.StudentInput) (*models.Student, error) {
	if err := input.Normalize(); err != nil {
		return nil, err
	}
	student, err := u.repo.GetStudent(ctx, id)
	if err != nil {
		return nil, notFound(err, progress.ErrStudentNotFound)
	}
	student.Name = input.Name
	if err := u.repo.RenameStudent(ctx, student); err != nil {
		return nil, err
	}
	return student, nil
}

func (u *UseCase) RemoveStudent(ctx context.Context, id int) error {
	student, err := u.repo.GetStudent(ctx, id)
	if err != nil {
		return notFound(err, progress.ErrStudentNotFound)
	}
	return u.repo.DeleteStudent(ctx, student)
}

func (u *UseCase) SetMark(ctx context.Context, studentId, taskId int, input progress.MarkInput) error {
	if err := input.Normalize(); err != nil {
		return err
	}
	student, err := u.repo.GetStudent(ctx, studentId)
	if err != nil {
		return notFound(err, progress.ErrStudentNotFound)
	}
	task, err := u.repo.GetSubjectObject(ctx, taskId)
	if err != nil {
		return notFound(err, progress.ErrTaskNotFound)
	}
	// A mark makes sense only for a task of the student's own group.
	if task.Subject.GroupId != student.GroupId {
		return progress.ErrTaskNotFound
	}

	if input.Status == "" {
		return u.repo.DeleteMark(ctx, studentId, taskId)
	}
	return u.repo.SetMark(ctx, &models.Mark{
		StudentId:       studentId,
		SubjectObjectId: taskId,
		Status:          input.Status,
		Comment:         input.Comment,
	})
}

func (u *UseCase) GetRoster(ctx context.Context, groupId int) (*models.Group, []models.Student, error) {
	// A hidden group is not found for students.
	group, err := u.repo.GetGroup(ctx, groupId, true)
	if err != nil {
		return nil, nil, notFound(err, progress.ErrGroupNotFound)
	}
	students, err := u.repo.GetStudents(ctx, groupId)
	if err != nil {
		return nil, nil, err
	}
	return group, students, nil
}

func (u *UseCase) GetStudentProgress(ctx context.Context, studentId int) (*progress.StudentProgress, error) {
	student, err := u.repo.GetStudent(ctx, studentId)
	if err != nil {
		return nil, notFound(err, progress.ErrStudentNotFound)
	}
	// Students of a hidden group must not see anything either.
	group, err := u.repo.GetGroup(ctx, student.GroupId, true)
	if err != nil {
		return nil, notFound(err, progress.ErrGroupNotFound)
	}
	marks, err := u.repo.GetMarks(ctx, []int{studentId})
	if err != nil {
		return nil, err
	}

	markByTask := make(map[int]*models.Mark, len(marks))
	for i := range marks {
		markByTask[marks[i].SubjectObjectId] = &marks[i]
	}

	result := &progress.StudentProgress{Student: student, Group: group}
	for _, st := range subjectTasks(group) {
		subjectProgress := progress.SubjectProgress{Subject: st.Subject, Tasks: make([]progress.TaskProgress, 0, len(st.Tasks))}
		for i := range st.Tasks {
			subjectProgress.Tasks = append(subjectProgress.Tasks, progress.TaskProgress{
				Task: &st.Tasks[i],
				Mark: markByTask[st.Tasks[i].ID],
			})
		}
		result.Subjects = append(result.Subjects, subjectProgress)
	}
	return result, nil
}

func studentIds(students []models.Student) []int {
	ids := make([]int, 0, len(students))
	for i := range students {
		ids = append(ids, students[i].ID)
	}
	return ids
}

// subjectTasks flattens the preloaded group into subjects with their tasks.
func subjectTasks(group *models.Group) []progress.SubjectTasks {
	subjects := make([]progress.SubjectTasks, 0, len(group.Subjects))
	for i := range group.Subjects {
		subjects = append(subjects, progress.SubjectTasks{
			Subject: &group.Subjects[i],
			Tasks:   group.Subjects[i].SubjectObjects,
		})
	}
	return subjects
}
