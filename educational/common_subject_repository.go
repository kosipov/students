package educational

import (
	"context"
	"github.com/kosipov/students/models"
)

type CommonSubjectRepository interface {
	GetSubjectsByGroup(ctx context.Context, group *models.Group) (*[]models.Subject, error)
	GetSubjectObjectsBySubject(ctx context.Context, subject *models.Subject) (*[]models.SubjectObject, error)
	GetSubject(ctx context.Context, id int) (*models.Subject, error)
	GetSubjects(ctx context.Context) (*[]models.Subject, error)
}
