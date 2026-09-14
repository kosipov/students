package usecase

import (
	"context"
	"errors"
	"github.com/kosipov/students/educational"
	"github.com/kosipov/students/models"
	"testing"
	"time"
)

func TestGroupLifecycle(t *testing.T) {
	repo := newFakeRepo()
	uc := NewGroupUseCase(repo)
	ctx := context.Background()

	group, err := uc.CreateGroup(ctx, educational.GroupInput{Name: " ИСП-301 ", Hidden: true})
	if err != nil || group.GroupName != "ИСП-301" || !group.Hidden {
		t.Fatalf("CreateGroup() = %+v, %v; want trimmed hidden group", group, err)
	}

	show := false
	updated, err := uc.UpdateGroup(ctx, int(group.ID), educational.GroupPatch{Hidden: &show})
	if err != nil || updated.Hidden || updated.GroupName != "ИСП-301" {
		t.Fatalf("UpdateGroup(hidden) = %+v, %v; want only visibility changed", updated, err)
	}

	name := "ИСП-302"
	if updated, err = uc.UpdateGroup(ctx, int(group.ID), educational.GroupPatch{Name: &name}); err != nil || updated.GroupName != "ИСП-302" {
		t.Fatalf("UpdateGroup(name) = %+v, %v", updated, err)
	}

	empty := ""
	var validationErr *educational.ValidationError
	if _, err := uc.UpdateGroup(ctx, int(group.ID), educational.GroupPatch{Name: &empty}); !errors.As(err, &validationErr) {
		t.Errorf("UpdateGroup() with empty name error = %v, want ValidationError", err)
	}
	if _, err := uc.CreateGroup(ctx, educational.GroupInput{Name: " "}); !errors.As(err, &validationErr) {
		t.Errorf("CreateGroup() with empty name error = %v, want ValidationError", err)
	}
	if _, err := uc.UpdateGroup(ctx, 404, educational.GroupPatch{Name: &name}); !errors.Is(err, educational.ErrGroupNotFound) {
		t.Errorf("UpdateGroup() for missing group error = %v, want ErrGroupNotFound", err)
	}
}

func TestDeleteGroupDeletesSubjectsAndTasks(t *testing.T) {
	repo := newFakeRepo()
	repo.addGroup(models.Group{ID: 1, GroupName: "ИСП-301"})
	repo.addGroup(models.Group{ID: 2, GroupName: "ИСП-302"})
	repo.addSubject(models.Subject{ID: 10, GroupId: 1})
	repo.addSubject(models.Subject{ID: 20, GroupId: 2})
	repo.addSubjectObject(models.SubjectObject{ID: 11, SubjectId: 10})
	repo.addSubjectObject(models.SubjectObject{ID: 21, SubjectId: 20})
	uc := NewGroupUseCase(repo)
	ctx := context.Background()

	if err := uc.DeleteGroup(ctx, 1); err != nil {
		t.Fatalf("DeleteGroup() error = %v", err)
	}
	if _, err := repo.GetSubject(ctx, 10); err == nil {
		t.Error("subject of the deleted group is kept")
	}
	if _, err := repo.GetSubjectObject(ctx, 11); err == nil {
		t.Error("task of the deleted group is kept")
	}
	if _, err := repo.GetSubjectObject(ctx, 21); err != nil {
		t.Error("task of another group is deleted")
	}
	if err := uc.DeleteGroup(ctx, 1); !errors.Is(err, educational.ErrGroupNotFound) {
		t.Errorf("second DeleteGroup() error = %v, want ErrGroupNotFound", err)
	}
}

func TestGetOverview(t *testing.T) {
	repo := newFakeRepo()
	repo.addGroup(models.Group{ID: 1, GroupName: "ИСП-301"})
	repo.addGroup(models.Group{ID: 2, GroupName: "ИСП-302", Hidden: true})
	repo.addSubject(models.Subject{ID: 10, GroupId: 1, SubjectName: "Веб"})
	repo.addSubject(models.Subject{ID: 20, GroupId: 2, SubjectName: "БД"})
	repo.addSubjectObject(models.SubjectObject{ID: 11, SubjectId: 10, Name: "Старое", ContentFetchedAt: timeAgo(48 * time.Hour)})
	repo.addSubjectObject(models.SubjectObject{ID: 12, SubjectId: 10, Name: "Свежее", ContentFetchedAt: timeAgo(time.Hour), Hidden: true})
	repo.addSubjectObject(models.SubjectObject{ID: 21, SubjectId: 20, Name: "Сломанное", ContentError: "onedrive is down"})
	uc := NewGroupUseCase(repo)

	overview, err := uc.GetOverview(context.Background())
	if err != nil {
		t.Fatalf("GetOverview() error = %v", err)
	}
	if overview.Groups != 2 || overview.Subjects != 2 || overview.Tasks != 3 || overview.HiddenTasks != 1 {
		t.Errorf("overview counts = %+v, want 2 groups, 2 subjects, 3 tasks, 1 hidden", overview)
	}
	if len(overview.RecentlyUpdated) != 2 || overview.RecentlyUpdated[0].SubjectObject.Name != "Свежее" {
		t.Errorf("RecentlyUpdated = %+v, want downloaded tasks newest first", overview.RecentlyUpdated)
	}
	if len(overview.Issues) != 1 || overview.Issues[0].SubjectObject.Name != "Сломанное" ||
		overview.Issues[0].Group.GroupName != "ИСП-302" || overview.Issues[0].Subject.SubjectName != "БД" {
		t.Errorf("Issues = %+v, want the failed task with its group and subject", overview.Issues)
	}
}

func TestGetOverviewLimitsRecentlyUpdated(t *testing.T) {
	repo := newFakeRepo()
	repo.addGroup(models.Group{ID: 1})
	repo.addSubject(models.Subject{ID: 10, GroupId: 1})
	for i := 1; i <= recentlyUpdatedLimit+2; i++ {
		repo.addSubjectObject(models.SubjectObject{ID: 10 + i, SubjectId: 10, ContentFetchedAt: timeAgo(time.Duration(i) * time.Hour)})
	}

	overview, err := NewGroupUseCase(repo).GetOverview(context.Background())
	if err != nil || len(overview.RecentlyUpdated) != recentlyUpdatedLimit {
		t.Fatalf("GetOverview() = %d recently updated, %v; want %d", len(overview.RecentlyUpdated), err, recentlyUpdatedLimit)
	}
}
