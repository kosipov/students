package educational

import (
	"context"
	"github.com/kosipov/students/models"
)

const CtxSubjectKey = "educational"

type CommonSubjectUseCase interface {
	// GetSubjectWithSubjectObjects returns the subject with its group and subject objects for the admin panel.
	GetSubjectWithSubjectObjects(ctx context.Context, id int) (*models.Subject, error)
	CreateSubject(ctx context.Context, groupId int, input SubjectInput) (*models.Subject, error)
	UpdateSubject(ctx context.Context, id int, input SubjectInput) (*models.Subject, error)
	DeleteSubject(ctx context.Context, id int) error

	CreateSubjectObject(ctx context.Context, subjectId int, input SubjectObjectInput) (*models.SubjectObject, error)
	UpdateSubjectObject(ctx context.Context, id int, patch SubjectObjectPatch) (*models.SubjectObject, error)
	DeleteSubjectObject(ctx context.Context, id int) error
	RefreshSubjectObjectContent(ctx context.Context, id int) (*models.SubjectObject, error)
	// GetCategoryNames returns the distinct names of categories in use, for suggestions in the admin panel.
	GetCategoryNames(ctx context.Context) ([]string, error)
	// IsDocument reports whether the subject object is shown as a page on the site.
	IsDocument(subjectObject *models.SubjectObject) bool

	// GetTask returns a task with its stored document, refreshing it from the source
	// when it is outdated. ErrSubjectObjectNotDocument is returned (along with the task)
	// when the link can't be shown on the site. Hidden subject objects and subject objects
	// of hidden groups are not found.
	GetTask(ctx context.Context, subjectObjectId int) (*Task, error)
	// PreviewTask works like GetTask, but shows hidden tasks too.
	PreviewTask(ctx context.Context, subjectObjectId int) (*Task, error)
}
