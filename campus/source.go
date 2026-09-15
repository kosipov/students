package campus

import (
	"context"
	"github.com/kosipov/students/models"
	"github.com/kosipov/students/schedule"
	"time"
	"unicode/utf8"
)

// Source adapts the client to the schedule domain.
type Source struct {
	Client *Client
}

func (s Source) FetchWeeks(ctx context.Context) ([]schedule.Week, error) {
	weeks, err := s.Client.FetchWeeks(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]schedule.Week, 0, len(weeks))
	for _, week := range weeks {
		converted := schedule.Week{
			// A date column: Monday's calendar date, not the instant of midnight in Moscow.
			Start:   time.Date(week.Start.Year(), week.Start.Month(), week.Start.Day(), 0, 0, 0, 0, time.UTC),
			Updated: week.Updated,
			Lessons: make([]models.ScheduleLesson, 0, len(week.Lessons)),
		}
		for _, lesson := range week.Lessons {
			converted.Lessons = append(converted.Lessons, models.ScheduleLesson{
				Number:   lesson.Number,
				StartsAt: lesson.StartsAt.UTC(),
				EndsAt:   lesson.EndsAt.UTC(),
				Title:    truncate(lesson.Title, 255),
				Kind:     truncate(lesson.Kind, 50),
				Location: truncate(lesson.Location, 100),
				Room:     truncate(lesson.Room, 50),
				Building: truncate(lesson.Building, 50),
				Groups:   truncate(lesson.Groups, 255),
			})
		}
		result = append(result, converted)
	}
	return result, nil
}

// truncate keeps an unexpectedly long value from failing the whole sync on the column limit.
func truncate(value string, limit int) string {
	if utf8.RuneCountInString(value) <= limit {
		return value
	}
	return string([]rune(value)[:limit])
}
