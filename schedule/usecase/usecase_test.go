package usecase

import (
	"context"
	"errors"
	"github.com/kosipov/students/models"
	"github.com/kosipov/students/schedule"
	"sort"
	"sync"
	"testing"
	"time"
)

type fakeRepo struct {
	mu      sync.Mutex
	weeks   map[time.Time][]models.ScheduleLesson
	state   models.ScheduleSync
	saveErr error
}

func (r *fakeRepo) ReplaceWeeks(ctx context.Context, weeks []schedule.Week) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.saveErr != nil {
		return r.saveErr
	}
	for _, week := range weeks {
		r.weeks[week.Start] = week.Lessons
	}
	return nil
}

func (r *fakeRepo) GetLessons(ctx context.Context, from, to time.Time) ([]models.ScheduleLesson, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var lessons []models.ScheduleLesson
	for _, week := range r.weeks {
		for _, lesson := range week {
			if !lesson.StartsAt.Before(from) && lesson.StartsAt.Before(to) {
				lessons = append(lessons, lesson)
			}
		}
	}
	sort.Slice(lessons, func(i, j int) bool { return lessons[i].StartsAt.Before(lessons[j].StartsAt) })
	return lessons, nil
}

func (r *fakeRepo) GetSyncState(ctx context.Context) (*models.ScheduleSync, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	state := r.state
	return &state, nil
}

func (r *fakeRepo) SaveSyncState(ctx context.Context, state *models.ScheduleSync) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.state = *state
	return nil
}

type fakeSource struct {
	weeks []schedule.Week
	err   error
	calls int
}

func (s *fakeSource) FetchWeeks(ctx context.Context) ([]schedule.Week, error) {
	s.calls++
	return s.weeks, s.err
}

var monday = time.Date(2026, 9, 14, 0, 0, 0, 0, time.UTC)

// Tuesday 15.09.2026 17:40–19:10 Moscow time in 245/1.
var tuesdayLesson = models.ScheduleLesson{
	Number:   6,
	StartsAt: time.Date(2026, 9, 15, 14, 40, 0, 0, time.UTC),
	EndsAt:   time.Date(2026, 9, 15, 16, 10, 0, 0, time.UTC),
	Title:    "Проектирование и разработка веб-приложений",
	Location: "245/1",
}

func newUseCase(source *fakeSource, now time.Time) (*UseCase, *fakeRepo) {
	repo := &fakeRepo{weeks: map[time.Time][]models.ScheduleLesson{}}
	uc := NewUseCase(repo, source)
	uc.now = func() time.Time { return now }
	return uc, repo
}

func TestSyncStoresWeeksAndState(t *testing.T) {
	updated := time.Date(2026, 9, 15, 18, 28, 12, 0, time.FixedZone("MSK", 3*60*60))
	source := &fakeSource{weeks: []schedule.Week{
		{Start: monday, Updated: updated, Lessons: []models.ScheduleLesson{tuesdayLesson}},
		{Start: monday.AddDate(0, 0, 7), Updated: updated},
	}}
	now := time.Date(2026, 9, 15, 10, 0, 0, 0, time.UTC)
	uc, repo := newUseCase(source, now)

	if err := uc.Sync(context.Background()); err != nil {
		t.Fatalf("Sync() error = %v", err)
	}
	if len(repo.weeks[monday]) != 1 {
		t.Errorf("stored weeks = %v, want the lesson of the current week", repo.weeks)
	}
	if repo.state.SucceededAt == nil || !repo.state.SucceededAt.Equal(now) || repo.state.Error != "" ||
		repo.state.CampusUpdatedAt == nil || !repo.state.CampusUpdatedAt.Equal(updated) {
		t.Errorf("state = %+v, want success at %v with the campus update time", repo.state, now)
	}
}

func TestSyncFailureKeepsLessonsAndRecordsError(t *testing.T) {
	source := &fakeSource{weeks: []schedule.Week{{Start: monday, Lessons: []models.ScheduleLesson{tuesdayLesson}}}}
	first := time.Date(2026, 9, 15, 10, 0, 0, 0, time.UTC)
	uc, repo := newUseCase(source, first)
	if err := uc.Sync(context.Background()); err != nil {
		t.Fatalf("first Sync() error = %v", err)
	}

	source.err = errors.New("campus is down")
	later := first.Add(time.Hour)
	uc.now = func() time.Time { return later }
	if err := uc.Sync(context.Background()); err == nil {
		t.Fatal("Sync() with a failing source returned no error")
	}

	if len(repo.weeks[monday]) != 1 {
		t.Error("lessons were lost after a failed sync")
	}
	if !repo.state.CheckedAt.Equal(later) || !repo.state.SucceededAt.Equal(first) || repo.state.Error != "campus is down" {
		t.Errorf("state = %+v, want checked later, last success kept, error stored", repo.state)
	}

	source.err = nil
	uc.now = func() time.Time { return later.Add(time.Hour) }
	if err := uc.Sync(context.Background()); err != nil || repo.state.Error != "" {
		t.Errorf("recovered Sync() = %v, state error %q; want the error cleared", err, repo.state.Error)
	}
}

func TestSyncNowIsLimited(t *testing.T) {
	source := &fakeSource{}
	now := time.Date(2026, 9, 15, 10, 0, 0, 0, time.UTC)
	uc, _ := newUseCase(source, now)

	if err := uc.SyncNow(context.Background()); err != nil {
		t.Fatalf("SyncNow() error = %v", err)
	}
	if err := uc.SyncNow(context.Background()); !errors.Is(err, schedule.ErrSyncTooOften) {
		t.Errorf("second SyncNow() error = %v, want ErrSyncTooOften", err)
	}
	uc.now = func() time.Time { return now.Add(time.Minute) }
	if err := uc.SyncNow(context.Background()); err != nil {
		t.Errorf("SyncNow() a minute later error = %v", err)
	}
	if source.calls != 2 {
		t.Errorf("source calls = %d, want 2", source.calls)
	}
}

func TestGetPresence(t *testing.T) {
	source := &fakeSource{weeks: []schedule.Week{{Start: monday, Lessons: []models.ScheduleLesson{tuesdayLesson}}}}
	syncedAt := time.Date(2026, 9, 15, 12, 0, 0, 0, time.UTC)
	uc, _ := newUseCase(source, syncedAt)

	presence, err := uc.GetPresence(context.Background())
	if err != nil || presence.Status != schedule.StatusUnknown {
		t.Fatalf("GetPresence() before any sync = %+v, %v; want unknown", presence, err)
	}

	if err := uc.Sync(context.Background()); err != nil {
		t.Fatal(err)
	}

	// 18:00 Moscow time, during the lesson.
	uc.now = func() time.Time { return time.Date(2026, 9, 15, 15, 0, 0, 0, time.UTC) }
	presence, _ = uc.GetPresence(context.Background())
	if presence.Status != schedule.StatusInClass || presence.Current == nil || presence.Current.Location != "245/1" || presence.Stale {
		t.Errorf("presence during the lesson = %+v, want in class in 245/1", presence)
	}

	// 23:30 UTC on Tuesday is already Wednesday in Moscow: Tuesday's lesson is not "today".
	uc.now = func() time.Time { return time.Date(2026, 9, 15, 23, 30, 0, 0, time.UTC) }
	presence, _ = uc.GetPresence(context.Background())
	if presence.Status != schedule.StatusAway || presence.Next != nil {
		t.Errorf("presence after midnight in Moscow = %+v, want away with nothing today", presence)
	}

	// 02:00 Moscow time on Tuesday: the lesson is later today.
	uc.now = func() time.Time { return time.Date(2026, 9, 14, 23, 0, 0, 0, time.UTC) }
	presence, _ = uc.GetPresence(context.Background())
	if presence.Status != schedule.StatusAway || presence.Next == nil {
		t.Errorf("presence early on Tuesday = %+v, want away with the lesson next", presence)
	}

	uc.now = func() time.Time { return syncedAt.Add(schedule.StaleAfter + time.Minute) }
	if presence, _ = uc.GetPresence(context.Background()); !presence.Stale {
		t.Error("presence long after the last sync is not marked stale")
	}
}

func TestRunRetriesSoonerAfterFailure(t *testing.T) {
	source := &fakeSource{err: errors.New("campus is down")}
	uc, _ := newUseCase(source, time.Now())
	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})
	go func() {
		uc.Run(ctx, time.Hour)
		close(done)
	}()

	time.Sleep(50 * time.Millisecond)
	cancel()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("Run() doesn't stop when the context is done")
	}
	if source.calls != 1 {
		t.Errorf("source calls = %d, want the first sync right away", source.calls)
	}
}
