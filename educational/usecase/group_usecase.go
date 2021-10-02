package usecase

import (
	"context"
	"github.com/kosipov/students/educational"
	"github.com/kosipov/students/models"
)

type GroupUseCase struct {
	groupRepo educational.CommonGroupRepository
}

func NewGroupUseCase(groupRepo educational.CommonGroupRepository) *GroupUseCase {
	return &GroupUseCase{groupRepo: groupRepo}
}

func (g *GroupUseCase) GetAllGroups(ctx context.Context) (*[]models.Group, error) {
	return g.groupRepo.GetGroups(ctx)
}

func (g *GroupUseCase) GetGroupById(ctx context.Context, id int) (*models.Group, error) {
	return g.groupRepo.GetGroupById(ctx, id)
}

func (g *GroupUseCase) CreateGroup(ctx context.Context, groupName string) error {
	group := &models.Group{GroupName: groupName}
	return g.groupRepo.CreateGroup(ctx, group)
}
