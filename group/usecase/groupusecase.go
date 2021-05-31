package usecase

import (
	"context"
	"github.com/kosipov/students/group"
	"github.com/kosipov/students/models"
)

type GroupUseCase struct {
	groupRepo group.GroupRepository
}

func NewGroupUseCase(groupRepo group.GroupRepository) *GroupUseCase {
	return &GroupUseCase{groupRepo: groupRepo}
}

func (g *GroupUseCase) GetAllGroups(ctx context.Context) (*[]models.Group, error) {
	return g.groupRepo.GetGroups(ctx)
}
