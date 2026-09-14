package educational

import (
	"context"
	"github.com/kosipov/students/models"
)

type CommonGroupRepository interface {
	GetGroups(ctx context.Context) (*[]models.Group, error)
	// GetVisibleGroups returns groups that are not hidden, with subjects and subject objects that are not hidden.
	GetVisibleGroups(ctx context.Context) (*[]models.Group, error)
	GetGroupById(ctx context.Context, id int) (*models.Group, error)
	CreateGroup(ctx context.Context, group *models.Group) error
	UpdateGroupHidden(ctx context.Context, group *models.Group) error
}
