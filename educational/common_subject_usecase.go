package educational

import (
	"context"
	"github.com/kosipov/students/models"
)

const CtxSubjectKey = "educational"

type CommonSubjectUseCase interface {
	GetSubjectByGroup(ctx context.Context, group *models.Group) (*[]models.Subject, error)
	SubjectObjectListFromSubject(ctx context.Context, subject *models.Subject) (*[]models.SubjectObject, error)
	GetSubjectById(ctx context.Context, id int) (*models.Subject, error)
	GetAllSubject(ctx context.Context) (*[]models.Subject, error)
}
