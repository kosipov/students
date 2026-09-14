package educational

import (
	"context"
	"github.com/kosipov/students/models"
)

const CtxSubjectKey = "educational"

type CommonSubjectUseCase interface {
	GetSubjectsByGroup(ctx context.Context, groupId int) (*[]models.Subject, error)
	// GetVisibleSubjectsByGroup returns ErrGroupNotFound for a hidden group.
	GetVisibleSubjectsByGroup(ctx context.Context, groupId int) (*[]models.Subject, error)
	SubjectObjectListFromSubject(ctx context.Context, subjectId int) (*[]models.SubjectObject, error)
	// VisibleSubjectObjectListFromSubject skips hidden subject objects and returns
	// ErrSubjectNotFound when the group of the subject is hidden.
	VisibleSubjectObjectListFromSubject(ctx context.Context, subjectId int) (*[]models.SubjectObject, error)
	GetSubjectById(ctx context.Context, id int) (*models.Subject, error)
	GetAllSubject(ctx context.Context) (*[]models.Subject, error)
	CreateSubject(ctx context.Context, name string, groupId int) error
	CreateSubjectObject(ctx context.Context, name string, subjectId int, href string) (*models.SubjectObject, error)
	GetSubjectObject(ctx context.Context, subjectId int, subjectObjectId int) (*models.SubjectObject, error)
	UpdateSubjectObject(ctx context.Context, subjectId int, subjectObjectId int, name string, href string) (*models.SubjectObject, error)
	DeleteSubjectObject(ctx context.Context, subjectId int, subjectObjectId int) error
	SetSubjectObjectHidden(ctx context.Context, subjectId int, subjectObjectId int, hidden bool) error
	// GetTask returns a subject object with its stored document, refreshing it from the source
	// when it is outdated. ErrSubjectObjectNotDocument is returned (along with the subject object)
	// when the link can't be shown on the site. Hidden subject objects and subject objects
	// of hidden groups are not found.
	GetTask(ctx context.Context, subjectObjectId int) (*models.SubjectObject, error)
	// PreviewTask works like GetTask, but shows hidden subject objects too.
	PreviewTask(ctx context.Context, subjectId int, subjectObjectId int) (*models.SubjectObject, error)
	RefreshSubjectObjectContent(ctx context.Context, subjectId int, subjectObjectId int) (*models.SubjectObject, error)
	IsDocument(subjectObject *models.SubjectObject) bool
}
