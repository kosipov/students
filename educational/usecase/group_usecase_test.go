package usecase

import (
	"context"
	"errors"
	"github.com/jinzhu/gorm"
	"github.com/kosipov/students/educational"
	"github.com/kosipov/students/models"
	"testing"
)

type fakeGroupRepo struct {
	educational.CommonGroupRepository

	groups map[int]models.Group
}

func (r *fakeGroupRepo) GetGroupById(ctx context.Context, id int) (*models.Group, error) {
	group, ok := r.groups[id]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	return &group, nil
}

func (r *fakeGroupRepo) UpdateGroupHidden(ctx context.Context, group *models.Group) error {
	stored := r.groups[int(group.ID)]
	stored.Hidden = group.Hidden
	r.groups[int(group.ID)] = stored
	return nil
}

func TestSetGroupHidden(t *testing.T) {
	repo := &fakeGroupRepo{groups: map[int]models.Group{1: {ID: 1, GroupName: "ИСП-301"}}}
	uc := NewGroupUseCase(repo)

	if err := uc.SetGroupHidden(context.Background(), 1, true); err != nil {
		t.Fatalf("SetGroupHidden() error = %v", err)
	}
	if !repo.groups[1].Hidden {
		t.Error("group is not hidden")
	}

	if err := uc.SetGroupHidden(context.Background(), 1, false); err != nil || repo.groups[1].Hidden {
		t.Errorf("SetGroupHidden(false) error = %v, hidden = %v", err, repo.groups[1].Hidden)
	}

	if err := uc.SetGroupHidden(context.Background(), 2, true); !errors.Is(err, educational.ErrGroupNotFound) {
		t.Errorf("SetGroupHidden() for missing group error = %v, want ErrGroupNotFound", err)
	}
}
