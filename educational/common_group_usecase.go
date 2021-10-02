package educational

import (
	"context"
	"github.com/kosipov/students/models"
)

const CtxGroupKey = "group"

type CommonGroupUseCase interface {
	GetAllGroups(ctx context.Context) (*[]models.Group, error)
	GetGroupById(ctx context.Context, id int) (*models.Group, error)
	CreateGroup(ctx context.Context, groupName string) error
}
