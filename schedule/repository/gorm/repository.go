package gorm

import (
	"context"
	"github.com/jinzhu/gorm"
	"github.com/kosipov/students/models"
	"github.com/kosipov/students/schedule"
	"time"
)

// keepPastWeeks is how long lessons of past weeks are kept; the site never needs them.
const keepPastWeeks = 2

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) ReplaceWeeks(ctx context.Context, weeks []schedule.Week) error {
	if len(weeks) == 0 {
		return nil
	}

	return r.db.Transaction(func(tx *gorm.DB) error {
		oldest := weeks[0].Start
		for _, week := range weeks {
			if week.Start.Before(oldest) {
				oldest = week.Start
			}
			if err := tx.Where("week_start = ?", week.Start).Delete(&models.ScheduleLesson{}).Error; err != nil {
				return err
			}
			for i := range week.Lessons {
				lesson := week.Lessons[i]
				lesson.ID = 0
				lesson.WeekStart = week.Start
				if err := tx.Create(&lesson).Error; err != nil {
					return err
				}
			}
		}
		return tx.Where("week_start < ?", oldest.AddDate(0, 0, -7*keepPastWeeks)).Delete(&models.ScheduleLesson{}).Error
	})
}

func (r *Repository) GetLessons(ctx context.Context, from, to time.Time) ([]models.ScheduleLesson, error) {
	var lessons []models.ScheduleLesson
	err := r.db.Where("starts_at >= ? AND starts_at < ?", from, to).Order("starts_at, number, id").Find(&lessons).Error
	return lessons, err
}

func (r *Repository) GetSyncState(ctx context.Context) (*models.ScheduleSync, error) {
	var state models.ScheduleSync
	err := r.db.Order("id").First(&state).Error
	if gorm.IsRecordNotFoundError(err) {
		return &models.ScheduleSync{}, nil
	}
	return &state, err
}

func (r *Repository) SaveSyncState(ctx context.Context, state *models.ScheduleSync) error {
	if state.ID == 0 {
		return r.db.Create(state).Error
	}
	return r.db.Model(&models.ScheduleSync{ID: state.ID}).Updates(map[string]interface{}{
		"checked_at":        state.CheckedAt,
		"succeeded_at":      state.SucceededAt,
		"campus_updated_at": state.CampusUpdatedAt,
		"error":             state.Error,
	}).Error
}
