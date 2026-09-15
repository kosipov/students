package schedule

import (
	"github.com/kosipov/students/models"
	"time"
)

type PresenceStatus string

const (
	// StatusInClass: a lesson is going on, or starts within ArrivalMargin.
	StatusInClass PresenceStatus = "in_class"
	// StatusBreak: between two lessons of the day with no more than MaxBreak between them.
	StatusBreak PresenceStatus = "break"
	// StatusAway: no lesson now; there may be one later today.
	StatusAway PresenceStatus = "away"
	// StatusUnknown: the schedule has never been loaded.
	StatusUnknown PresenceStatus = "unknown"
)

const (
	// ArrivalMargin: the teacher is usually in the room a few minutes before the lesson.
	ArrivalMargin = 10 * time.Minute
	// MaxBreak: a gap up to the lunch break (an hour) still means the teacher is at the university.
	MaxBreak = time.Hour
	// StaleAfter: when the last successful sync is older, the schedule may have changed without us knowing.
	StaleAfter = 6 * time.Hour
)

type Presence struct {
	Status PresenceStatus
	// Current is the lesson going on (in class).
	Current *models.ScheduleLesson
	// Next is the next lesson today.
	Next     *models.ScheduleLesson
	SyncedAt *time.Time
	Stale    bool
}

// ComputePresence finds where the teacher is at now from the day's lessons ordered by start.
func ComputePresence(now time.Time, lessons []models.ScheduleLesson) (PresenceStatus, *models.ScheduleLesson, *models.ScheduleLesson) {
	var previous *models.ScheduleLesson
	for i := range lessons {
		lesson := &lessons[i]
		switch {
		case !now.Before(lesson.StartsAt.Add(-ArrivalMargin)) && now.Before(lesson.EndsAt):
			return StatusInClass, lesson, nextAfter(lessons, i)
		case !lesson.EndsAt.After(now):
			previous = lesson
		default:
			// The first lesson that hasn't started yet.
			if previous != nil && lesson.StartsAt.Sub(previous.EndsAt) <= MaxBreak {
				return StatusBreak, nil, lesson
			}
			return StatusAway, nil, lesson
		}
	}
	return StatusAway, nil, nil
}

// nextAfter returns the first lesson starting after lesson i ends; parallel lessons at the same time are skipped.
func nextAfter(lessons []models.ScheduleLesson, i int) *models.ScheduleLesson {
	for j := i + 1; j < len(lessons); j++ {
		if !lessons[j].StartsAt.Before(lessons[i].EndsAt) {
			return &lessons[j]
		}
	}
	return nil
}
