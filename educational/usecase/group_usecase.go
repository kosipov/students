package usecase

import (
	"context"
	"github.com/kosipov/students/educational"
	"github.com/kosipov/students/models"
	"sort"
)

// recentlyUpdatedLimit is how many recently updated tasks the overview shows.
const recentlyUpdatedLimit = 5

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
	group, err := g.groupRepo.GetGroupById(ctx, id)
	if err != nil {
		return nil, notFound(err, educational.ErrGroupNotFound)
	}
	return group, nil
}

func (g *GroupUseCase) CreateGroup(ctx context.Context, input educational.GroupInput) (*models.Group, error) {
	if err := input.Normalize(); err != nil {
		return nil, err
	}

	group := &models.Group{GroupName: input.Name, Hidden: input.Hidden}
	if err := g.groupRepo.CreateGroup(ctx, group); err != nil {
		return nil, err
	}
	return group, nil
}

func (g *GroupUseCase) UpdateGroup(ctx context.Context, id int, patch educational.GroupPatch) (*models.Group, error) {
	if err := patch.Normalize(); err != nil {
		return nil, err
	}
	group, err := g.GetGroupById(ctx, id)
	if err != nil {
		return nil, err
	}

	if patch.Name != nil {
		group.GroupName = *patch.Name
	}
	if patch.Hidden != nil {
		group.Hidden = *patch.Hidden
	}

	if err := g.groupRepo.UpdateGroup(ctx, group); err != nil {
		return nil, err
	}
	return group, nil
}

func (g *GroupUseCase) DeleteGroup(ctx context.Context, id int) error {
	group, err := g.GetGroupById(ctx, id)
	if err != nil {
		return err
	}
	return g.groupRepo.DeleteGroup(ctx, group)
}

func (g *GroupUseCase) GetOverview(ctx context.Context) (*educational.Overview, error) {
	groups, err := g.groupRepo.GetGroups(ctx)
	if err != nil {
		return nil, err
	}

	overview := &educational.Overview{Groups: len(*groups)}
	for gi := range *groups {
		group := &(*groups)[gi]
		overview.Subjects += len(group.Subjects)

		for si := range group.Subjects {
			subject := &group.Subjects[si]
			overview.Tasks += len(subject.SubjectObjects)

			for oi := range subject.SubjectObjects {
				subjectObject := &subject.SubjectObjects[oi]
				task := educational.Task{SubjectObject: subjectObject, Subject: subject, Group: group}

				if subjectObject.Hidden {
					overview.HiddenTasks++
				}
				if subjectObject.ContentError != "" {
					overview.Issues = append(overview.Issues, task)
				}
				if subjectObject.ContentFetchedAt != nil {
					overview.RecentlyUpdated = append(overview.RecentlyUpdated, task)
				}
			}
		}
	}

	sort.SliceStable(overview.RecentlyUpdated, func(i, j int) bool {
		return overview.RecentlyUpdated[i].SubjectObject.ContentFetchedAt.After(*overview.RecentlyUpdated[j].SubjectObject.ContentFetchedAt)
	})
	if len(overview.RecentlyUpdated) > recentlyUpdatedLimit {
		overview.RecentlyUpdated = overview.RecentlyUpdated[:recentlyUpdatedLimit]
	}

	return overview, nil
}
