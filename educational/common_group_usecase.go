package educational

import (
	"context"
	"github.com/kosipov/students/models"
)

const CtxGroupKey = "group"

type CommonGroupUseCase interface {
	// GetAllGroups returns all groups with subjects and subject objects for the admin panel.
	GetAllGroups(ctx context.Context) (*[]models.Group, error)
	// GetVisibleGroups returns what students can see: groups and subject objects that are not hidden.
	GetVisibleGroups(ctx context.Context) (*[]models.Group, error)
	GetGroupById(ctx context.Context, id int) (*models.Group, error)
	CreateGroup(ctx context.Context, input GroupInput) (*models.Group, error)
	UpdateGroup(ctx context.Context, id int, patch GroupPatch) (*models.Group, error)
	DeleteGroup(ctx context.Context, id int) error
	GetOverview(ctx context.Context) (*Overview, error)
}
