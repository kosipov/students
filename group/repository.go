package group

import (
	"context"
	"github.com/kosipov/students/models"
)

type GroupRepository interface {
	GetGroups(ctx context.Context) (*[]models.Group, error)
}
