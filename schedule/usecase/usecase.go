package usecase

import (
	"context"
	"github.com/kosipov/students/models"
	"github.com/kosipov/students/schedule"
	"log"
	"sync"
	"time"
)

const (
	// manualSyncInterval keeps the admin's "refresh now" button from hammering the university's site.
	manualSyncInterval = time.Minute
	// retryInterval is used after a failed sync instead of the regular interval.
	retryInterval = 5 * time.Minute
)

// location is the university's time zone, Moscow time, which defines what "today" is.
var location = time.FixedZone("MSK", 3*60*60)

type UseCase struct {
	repo   schedule.Repository
	source schedule.Source
	now    func() time.Time

	syncMu     sync.Mutex
	manualMu   sync.Mutex
	lastManual time.Time
}

func NewUseCase(repo schedule.Repository, source schedule.Source) *UseCase {
	return &UseCase{repo: repo, source: source, now: time.Now}
}

func (uc *UseCase) Sync(ctx context.Context) error {
	// The periodic and the manual sync must not replace lessons at the same time.
	uc.syncMu.Lock()
	defer uc.syncMu.Unlock()

	state, err := uc.repo.GetSyncState(ctx)
	if err != nil {
		return err
	}
	checkedAt := uc.now().UTC()
	state.CheckedAt = &checkedAt

	weeks, err := uc.source.FetchWeeks(ctx)
	if err == nil {
		err = uc.repo.ReplaceWeeks(ctx, weeks)
	}
	if err != nil {
		log.Printf("Failed to sync schedule: %s", err)
		state.Error = err.Error()
		if saveErr := uc.repo.SaveSyncState(ctx, state); saveErr != nil {
			log.Printf("Failed to save schedule sync state: %s", saveErr)
		}
		return err
	}

	state.SucceededAt = &checkedAt
	state.Error = ""
	for _, week := range weeks {
		if updated := week.Updated.UTC(); !week.Updated.IsZero() && (state.CampusUpdatedAt == nil || updated.After(*state.CampusUpdatedAt)) {
			state.CampusUpdatedAt = &updated
		}
	}
	return uc.repo.SaveSyncState(ctx, state)
}

func (uc *UseCase) SyncNow(ctx context.Context) error {
	uc.manualMu.Lock()
	now := uc.now()
	if !uc.lastManual.IsZero() && now.Sub(uc.lastManual) < manualSyncInterval {
		uc.manualMu.Unlock()
		return schedule.ErrSyncTooOften
	}
	uc.lastManual = now
	uc.manualMu.Unlock()

	return uc.Sync(ctx)
}

func (uc *UseCase) GetSyncState(ctx context.Context) (*models.ScheduleSync, error) {
	return uc.repo.GetSyncState(ctx)
}

func (uc *UseCase) GetPresence(ctx context.Context) (*schedule.Presence, error) {
	state, err := uc.repo.GetSyncState(ctx)
	if err != nil {
		return nil, err
	}
	if state.SucceededAt == nil {
		return &schedule.Presence{Status: schedule.StatusUnknown}, nil
	}

	now := uc.now()
	local := now.In(location)
	dayStart := time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, location)
	lessons, err := uc.repo.GetLessons(ctx, dayStart.UTC(), dayStart.AddDate(0, 0, 1).UTC())
	if err != nil {
		return nil, err
	}

	status, current, next := schedule.ComputePresence(now, lessons)
	return &schedule.Presence{
		Status:   status,
		Current:  current,
		Next:     next,
		SyncedAt: state.SucceededAt,
		Stale:    now.Sub(*state.SucceededAt) > schedule.StaleAfter,
	}, nil
}

// Run syncs right away and then every interval, or sooner after a failure, until ctx is done.
func (uc *UseCase) Run(ctx context.Context, interval time.Duration) {
	for {
		wait := interval
		syncCtx, cancel := context.WithTimeout(ctx, time.Minute)
		if err := uc.Sync(syncCtx); err != nil && retryInterval < interval {
			wait = retryInterval
		}
		cancel()

		select {
		case <-ctx.Done():
			return
		case <-time.After(wait):
		}
	}
}
