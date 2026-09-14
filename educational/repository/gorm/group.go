package gorm

import (
	"context"
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
	result := preloadSubjectObjects(g.db, false).Find(&groups)
	return &groups, result.Error
}

func (g *GroupRepository) GetVisibleGroups(ctx context.Context) (*[]models.Group, error) {
	var groups []models.Group
	result := preloadSubjectObjects(g.db, true).
		Where("hidden = ?", false).
		Find(&groups)
	return &groups, result.Error
}

func (g *GroupRepository) GetGroupById(ctx context.Context, id int) (*models.Group, error) {
	var group models.Group
	result := g.db.First(&group, id)
	return &group, result.Error
}

func (g *GroupRepository) CreateGroup(ctx context.Context, group *models.Group) error {
	newGroup := g.db.Create(group)
	return newGroup.Error
}

func (g *GroupRepository) UpdateGroupHidden(ctx context.Context, group *models.Group) error {
	return g.db.Model(group).Update("hidden", group.Hidden).Error
}

// preloadSubjectObjects loads subjects with their subject objects for lists.
// Stored documents are not needed for lists, so they are not loaded.
func preloadSubjectObjects(db *gorm.DB, onlyVisible bool) *gorm.DB {
	return db.
		Preload("Subjects").
		Preload("Subjects.SubjectObjects", func(db *gorm.DB) *gorm.DB {
			db = db.Select("id, name, comment, href, subject_id, hidden")
			if onlyVisible {
				db = db.Where("hidden = ?", false)
			}
			return db
		})
}
