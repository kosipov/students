package educational

import (
	"context"
	"github.com/kosipov/students/models"
)

type CommonGroupRepository interface {
	GetGroups(ctx context.Context) (*[]models.Group, error)
	GetGroupById(ctx context.Context, id int) (*models.Group, error)
}
