package educational

import (
	"context"
	"github.com/kosipov/students/models"
)

type CommonSubjectRepository interface {
	GetGroup(ctx context.Context, groupId int) (*models.Group, error)
	GetSubject(ctx context.Context, id int) (*models.Subject, error)
	// GetSubjectWithSubjectObjects returns the subject with its group and subject objects, without stored documents.
	GetSubjectWithSubjectObjects(ctx context.Context, id int) (*models.Subject, error)
	CreateSubject(ctx context.Context, subject *models.Subject) error
	UpdateSubject(ctx context.Context, subject *models.Subject) error
	// DeleteSubject deletes the subject with its subject objects.
	DeleteSubject(ctx context.Context, subject *models.Subject) error

	GetSubjectObject(ctx context.Context, subjectObjectId int) (*models.SubjectObject, error)
	CreateSubjectObject(ctx context.Context, subjectObject *models.SubjectObject) error
	// UpdateSubjectObject stores the fields edited in the admin panel: name, href, comment and hidden.
	UpdateSubjectObject(ctx context.Context, subjectObject *models.SubjectObject) error
	UpdateSubjectObjectContent(ctx context.Context, subjectObject *models.SubjectObject) error
	DeleteSubjectObject(ctx context.Context, subjectObject *models.SubjectObject) error
}
