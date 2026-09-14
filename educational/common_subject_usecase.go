package educational

import (
	"context"
	"github.com/kosipov/students/models"
)

const CtxSubjectKey = "educational"

type CommonSubjectUseCase interface {
	GetSubjectsByGroup(ctx context.Context, groupId int) (*[]models.Subject, error)
	SubjectObjectListFromSubject(ctx context.Context, subjectId int) (*[]models.SubjectObject, error)
	GetSubjectById(ctx context.Context, id int) (*models.Subject, error)
	GetAllSubject(ctx context.Context) (*[]models.Subject, error)
	CreateSubject(ctx context.Context, name string, groupId int) error
	CreateSubjectObject(ctx context.Context, name string, subjectId int, href string) (*models.SubjectObject, error)
	GetSubjectObject(ctx context.Context, subjectId int, subjectObjectId int) (*models.SubjectObject, error)
	UpdateSubjectObject(ctx context.Context, subjectId int, subjectObjectId int, name string, href string) (*models.SubjectObject, error)
	DeleteSubjectObject(ctx context.Context, subjectId int, subjectObjectId int) error
}
