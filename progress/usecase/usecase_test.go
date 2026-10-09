package usecase

import (
	"context"
	"testing"

	"github.com/jinzhu/gorm"
	"github.com/kosipov/students/models"
	"github.com/kosipov/students/progress"
)

type fakeRepo struct {
	group    *models.Group
	students []models.Student
	marks    []models.Mark
	tasks    map[int]*models.SubjectObject

	created []models.Student
	setMark *models.Mark
	deleted bool
}

func (f *fakeRepo) GetGroup(ctx context.Context, id int, onlyVisible bool) (*models.Group, error) {
	if f.group == nil {
		return nil, gorm.ErrRecordNotFound
	}
	group := *f.group
	if onlyVisible {
		if group.Hidden {
			return nil, gorm.ErrRecordNotFound
		}
		subjects := make([]models.Subject, 0, len(group.Subjects))
		for _, subject := range group.Subjects {
			tasks := make([]models.SubjectObject, 0, len(subject.SubjectObjects))
			for _, task := range subject.SubjectObjects {
				if !task.Hidden {
					tasks = append(tasks, task)
				}
			}
			subject.SubjectObjects = tasks
			subjects = append(subjects, subject)
		}
		group.Subjects = subjects
	}
	return &group, nil
}

func (f *fakeRepo) GetStudent(ctx context.Context, id int) (*models.Student, error) {
	for i := range f.students {
		if f.students[i].ID == id {
			return &f.students[i], nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}

func (f *fakeRepo) GetStudents(ctx context.Context, groupId int) ([]models.Student, error) {
	return f.students, nil
}

func (f *fakeRepo) CreateStudent(ctx context.Context, student *models.Student) error {
	f.created = append(f.created, *student)
	return nil
}

func (f *fakeRepo) RenameStudent(ctx context.Context, student *models.Student) error {
	return nil
}

func (f *fakeRepo) DeleteStudent(ctx context.Context, student *models.Student) error {
	f.deleted = true
	return nil
}

func (f *fakeRepo) GetMarks(ctx context.Context, studentIds []int) ([]models.Mark, error) {
	return f.marks, nil
}

func (f *fakeRepo) GetSubjectObject(ctx context.Context, id int) (*models.SubjectObject, error) {
	task, ok := f.tasks[id]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	return task, nil
}

func (f *fakeRepo) SetMark(ctx context.Context, mark *models.Mark) error {
	f.setMark = mark
	return nil
}

func (f *fakeRepo) DeleteMark(ctx context.Context, studentId, subjectObjectId int) error {
	f.setMark = nil
	f.deleted = true
	return nil
}

func TestImportStudentsSkipsDuplicates(t *testing.T) {
	repo := &fakeRepo{
		group:    &models.Group{ID: 1},
		students: []models.Student{{ID: 1, Name: "Иванов Иван", GroupId: 1}},
	}
	useCase := NewUseCase(repo, &fakeRoster{})

	result, err := useCase.ImportStudents(context.Background(), 1, "Петров Пётр\nиванов иван\n\nПетров Пётр\nСидорова Анна")
	if err != nil {
		t.Fatalf("ImportStudents: %s", err)
	}
	if result.Added != 2 || result.Skipped != 2 {
		t.Fatalf("expected 2 added and 2 skipped, got %d and %d", result.Added, result.Skipped)
	}
	if len(repo.created) != 2 {
		t.Fatalf("expected 2 created students, got %d", len(repo.created))
	}
}

func TestSetMarkRequiresTaskOfOwnGroup(t *testing.T) {
	repo := &fakeRepo{
		students: []models.Student{{ID: 1, Name: "Иванов Иван", GroupId: 1}},
		tasks: map[int]*models.SubjectObject{
			7: {ID: 7, SubjectId: 3, Subject: models.Subject{ID: 3, GroupId: 2}},
		},
	}
	useCase := NewUseCase(repo, &fakeRoster{})

	err := useCase.SetMark(context.Background(), 1, 7, progress.MarkInput{Status: progress.StatusAccepted})
	if err != progress.ErrTaskNotFound {
		t.Fatalf("expected ErrTaskNotFound, got %v", err)
	}
}

func TestSetMarkEmptyStatusDeletes(t *testing.T) {
	repo := &fakeRepo{
		students: []models.Student{{ID: 1, Name: "Иванов Иван", GroupId: 1}},
		tasks: map[int]*models.SubjectObject{
			7: {ID: 7, SubjectId: 3, Subject: models.Subject{ID: 3, GroupId: 1}},
		},
	}
	useCase := NewUseCase(repo, &fakeRoster{})

	if err := useCase.SetMark(context.Background(), 1, 7, progress.MarkInput{Status: ""}); err != nil {
		t.Fatalf("SetMark: %s", err)
	}
	if !repo.deleted {
		t.Fatal("expected the mark to be deleted")
	}
}

func TestStudentProgressHidesHiddenTasks(t *testing.T) {
	repo := &fakeRepo{
		group: &models.Group{
			ID: 1,
			Subjects: []models.Subject{{
				ID: 3, GroupId: 1,
				SubjectObjects: []models.SubjectObject{
					{ID: 7, Name: "ЛР 1", SubjectId: 3},
					{ID: 8, Name: "ЛР 2", SubjectId: 3, Hidden: true},
				},
			}},
		},
		students: []models.Student{{ID: 1, Name: "Иванов Иван", GroupId: 1}},
		marks: []models.Mark{
			{StudentId: 1, SubjectObjectId: 7, Status: progress.StatusAccepted},
			{StudentId: 1, SubjectObjectId: 8, Status: progress.StatusSubmitted},
		},
	}
	useCase := NewUseCase(repo, &fakeRoster{})

	result, err := useCase.GetStudentProgress(context.Background(), 1)
	if err != nil {
		t.Fatalf("GetStudentProgress: %s", err)
	}
	if len(result.Subjects) != 1 || len(result.Subjects[0].Tasks) != 1 {
		t.Fatalf("expected 1 subject with 1 visible task, got %+v", result.Subjects)
	}
	task := result.Subjects[0].Tasks[0]
	if task.Task.ID != 7 || task.Mark == nil || task.Mark.Status != progress.StatusAccepted {
		t.Fatalf("unexpected task progress: %+v", task)
	}
}

type fakeRoster struct {
	names []string
	err   error
	got   string
}

func (f *fakeRoster) FetchStudents(ctx context.Context, group string) ([]string, error) {
	f.got = group
	return f.names, f.err
}

func TestImportFromCampus(t *testing.T) {
	repo := &fakeRepo{
		group:    &models.Group{ID: 1, GroupName: "121б-пио"},
		students: []models.Student{{ID: 1, Name: "Иванов Иван", GroupId: 1}},
	}
	roster := &fakeRoster{names: []string{"Иванов Иван", "Петрова Мария"}}
	useCase := NewUseCase(repo, roster)

	result, err := useCase.ImportFromCampus(context.Background(), 1)
	if err != nil {
		t.Fatalf("ImportFromCampus: %s", err)
	}
	if roster.got != "121б-пио" {
		t.Fatalf("expected the campus to be asked for 121б-пио, got %q", roster.got)
	}
	if result.Found != 2 || result.Added != 1 || result.Skipped != 1 {
		t.Fatalf("unexpected result: %+v", result)
	}
}

func TestImportFromCampusUnknownGroup(t *testing.T) {
	repo := &fakeRepo{group: &models.Group{ID: 1, GroupName: "нет-такой"}}
	useCase := NewUseCase(repo, &fakeRoster{names: []string{}})

	_, err := useCase.ImportFromCampus(context.Background(), 1)
	if err != progress.ErrRosterEmpty {
		t.Fatalf("expected ErrRosterEmpty, got %v", err)
	}
}
