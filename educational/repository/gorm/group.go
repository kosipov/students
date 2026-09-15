package gorm

import (
	"context"
	"errors"
	"github.com/jinzhu/gorm"
	"github.com/kosipov/students/models"
)

type GroupRepository struct {
	db *gorm.DB
}

func NewGroupRepository(db *gorm.DB) *GroupRepository {
	return &GroupRepository{db: db}
}

func (g *GroupRepository) GetGroups(ctx context.Context) (*[]models.Group, error) {
	var groups []models.Group
	result := preloadSubjects(g.db, false).Order("id").Find(&groups)
	return &groups, result.Error
}

func (g *GroupRepository) GetVisibleGroups(ctx context.Context) (*[]models.Group, error) {
	var groups []models.Group
	result := preloadSubjects(g.db, true).
		Where("hidden = ?", false).
		Order("id").
		Find(&groups)
	return &groups, result.Error
}

func (g *GroupRepository) GetGroupById(ctx context.Context, id int) (*models.Group, error) {
	var group models.Group
	result := preloadSubjects(g.db, false).First(&group, id)
	return &group, result.Error
}

func (g *GroupRepository) CreateGroup(ctx context.Context, group *models.Group) error {
	return g.db.Create(group).Error
}

func (g *GroupRepository) UpdateGroup(ctx context.Context, group *models.Group) error {
	return g.db.Model(&models.Group{ID: group.ID}).Updates(map[string]interface{}{
		"group_name": group.GroupName,
		"hidden":     group.Hidden,
	}).Error
}

func (g *GroupRepository) DeleteGroup(ctx context.Context, group *models.Group) error {
	if group.ID == 0 {
		// gorm deletes every row when the primary key is blank.
		return errors.New("delete group: empty id")
	}

	return g.db.Transaction(func(tx *gorm.DB) error {
		var subjectIds []int
		if err := tx.Model(&models.Subject{}).Where("group_id = ?", group.ID).Pluck("id", &subjectIds).Error; err != nil {
			return err
		}
		if len(subjectIds) > 0 {
			if err := deleteSubjectObjectsOfSubjects(tx, subjectIds); err != nil {
				return err
			}
			if err := tx.Where("id IN (?)", subjectIds).Delete(&models.Subject{}).Error; err != nil {
				return err
			}
		}
		return tx.Where("id = ?", group.ID).Delete(&models.Group{}).Error
	})
}

// subjectObjectListColumns are the subject object columns needed for lists: everything but the stored document.
const subjectObjectListColumns = "id, name, comment, href, subject_id, hidden, password, password_version, " +
	"content_fetched_at, content_checked_at, content_error, content_unsupported"

// preloadSubjects loads subjects with their subject objects for lists.
func preloadSubjects(db *gorm.DB, onlyVisible bool) *gorm.DB {
	return db.
		Preload("Subjects", func(db *gorm.DB) *gorm.DB {
			return db.Order("id")
		}).
		Preload("Subjects.SubjectObjects", func(db *gorm.DB) *gorm.DB {
			db = db.Select(subjectObjectListColumns).Order("id")
			if onlyVisible {
				db = db.Where("hidden = ?", false)
			}
			return db
		}).
		Preload("Subjects.SubjectObjects.Categories", orderById)
}
