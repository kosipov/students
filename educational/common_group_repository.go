package educational

import (
	"context"
	"github.com/kosipov/students/models"
)

type CommonGroupRepository interface {
	// GetGroups returns all groups with their subjects and subject objects, without stored documents.
	GetGroups(ctx context.Context) (*[]models.Group, error)
	// GetVisibleGroups returns groups that are not hidden, with subjects and subject objects that are not hidden.
	GetVisibleGroups(ctx context.Context) (*[]models.Group, error)
	// GetGroupById returns the group with its subjects and subject objects, without stored documents.
	GetGroupById(ctx context.Context, id int) (*models.Group, error)
	CreateGroup(ctx context.Context, group *models.Group) error
	UpdateGroup(ctx context.Context, group *models.Group) error
	// DeleteGroup deletes the group with its subjects and subject objects.
	DeleteGroup(ctx context.Context, group *models.Group) error
}
