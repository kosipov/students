package group

import (
	"context"
	"github.com/kosipov/students/models"
)

const CtxGroupKey = "group"

type UseCase interface {
	GetAllGroups(ctx context.Context) (*[]models.Group, error)
}
