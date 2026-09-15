package schedule

import (
	"context"
	"errors"
	"github.com/kosipov/students/models"
	"time"
)

// ErrSyncTooOften is returned when the schedule is refreshed by hand again right after the previous refresh.
var ErrSyncTooOften = errors.New("schedule: sync requested too often")

// Week is the teacher's lessons for one week.
type Week struct {
	// Start is Monday of the week, as a date in UTC.
	Start   time.Time
	Updated time.Time
	Lessons []models.ScheduleLesson
}

// Source reads the current and the next week of the teacher's schedule.
type Source interface {
	FetchWeeks(ctx context.Context) ([]Week, error)
}

type Repository interface {
	// ReplaceWeeks replaces the lessons of the given weeks and forgets lessons of weeks long past.
	ReplaceWeeks(ctx context.Context, weeks []Week) error
	// GetLessons returns lessons starting in [from, to), ordered by start.
	GetLessons(ctx context.Context, from, to time.Time) ([]models.ScheduleLesson, error)
	// GetSyncState returns the sync state, empty before the first sync.
	GetSyncState(ctx context.Context) (*models.ScheduleSync, error)
	SaveSyncState(ctx context.Context, state *models.ScheduleSync) error
}

type UseCase interface {
	// Sync copies the schedule from the source; a failure is stored in the sync state and keeps old lessons.
	Sync(ctx context.Context) error
	// SyncNow is Sync requested by the admin; it returns ErrSyncTooOften right after another manual sync.
	SyncNow(ctx context.Context) error
	GetPresence(ctx context.Context) (*Presence, error)
	GetSyncState(ctx context.Context) (*models.ScheduleSync, error)
}
