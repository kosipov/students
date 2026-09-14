package usecase

import (
	"context"
	"errors"
	"github.com/jinzhu/gorm"
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

func (g *GroupUseCase) GetVisibleGroups(ctx context.Context) (*[]models.Group, error) {
	return g.groupRepo.GetVisibleGroups(ctx)
}

func (g *GroupUseCase) GetGroupById(ctx context.Context, id int) (*models.Group, error) {
	return g.groupRepo.GetGroupById(ctx, id)
}

func (g *GroupUseCase) SetGroupHidden(ctx context.Context, id int, hidden bool) error {
	group, err := g.groupRepo.GetGroupById(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return educational.ErrGroupNotFound
		}
		return err
	}

	group.Hidden = hidden
	return g.groupRepo.UpdateGroupHidden(ctx, group)
}

func (g *GroupUseCase) CreateGroup(ctx context.Context, groupName string) error {
	group := &models.Group{GroupName: groupName}
	return g.groupRepo.CreateGroup(ctx, group)
}
