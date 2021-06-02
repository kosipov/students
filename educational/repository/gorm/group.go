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
	result := g.db.Preload("Subjects.SubjectObjects").Find(&groups)
	return &groups, result.Error
}

func (g *GroupRepository) GetGroupById(ctx context.Context, id int) (*models.Group, error) {
	var group models.Group
	result := g.db.First(&group, id)
	return &group, result.Error
}
