package http

import (
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/kosipov/students/models"
	"github.com/kosipov/students/schedule"
	"log"
	"net/http"
	"time"
)

// manualSyncTimeout fits into the server's write timeout: two page requests with a pause between them.
const manualSyncTimeout = 25 * time.Second

type Handler struct {
	useCase schedule.UseCase
}

type lessonResponse struct {
	Number   int       `json:"number"`
	Title    string    `json:"title"`
	Kind     string    `json:"kind"`
	Location string    `json:"location"`
	Room     string    `json:"room"`
	Building string    `json:"building"`
	Groups   string    `json:"groups"`
	StartsAt time.Time `json:"startsAt"`
	EndsAt   time.Time `json:"endsAt"`
}

type presenceResponse struct {
	Status   schedule.PresenceStatus `json:"status"`
	Current  *lessonResponse         `json:"current"`
	Next     *lessonResponse         `json:"next"`
	SyncedAt *time.Time              `json:"syncedAt"`
	Stale    bool                    `json:"stale"`
}

type syncStateResponse struct {
	CheckedAt       *time.Time `json:"checkedAt"`
	SucceededAt     *time.Time `json:"succeededAt"`
	CampusUpdatedAt *time.Time `json:"campusUpdatedAt"`
	Error           string     `json:"error"`
	Stale           bool       `json:"stale"`
}

func toLesson(lesson *models.ScheduleLesson) *lessonResponse {
	if lesson == nil {
		return nil
	}
	return &lessonResponse{
		Number:   lesson.Number,
		Title:    lesson.Title,
		Kind:     lesson.Kind,
		Location: lesson.Location,
		Room:     lesson.Room,
		Building: lesson.Building,
		Groups:   lesson.Groups,
		StartsAt: lesson.StartsAt.UTC(),
		EndsAt:   lesson.EndsAt.UTC(),
	}
}

func toSyncState(state *models.ScheduleSync) syncStateResponse {
	return syncStateResponse{
		CheckedAt:       state.CheckedAt,
		SucceededAt:     state.SucceededAt,
		CampusUpdatedAt: state.CampusUpdatedAt,
		Error:           state.Error,
		Stale:           state.SucceededAt == nil || time.Since(*state.SucceededAt) > schedule.StaleAfter,
	}
}

// Presence tells where the teacher can be found right now according to the university's schedule.
func (h *Handler) Presence(c *gin.Context) {
	presence, err := h.useCase.GetPresence(c.Request.Context())
	if err != nil {
		log.Printf("Failed to get presence: %s", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Что-то пошло не так, попробуйте ещё раз"})
		return
	}

	// Answers change every minute; a browser or proxy must not keep an old one.
	c.Header("Cache-Control", "no-store")
	c.JSON(http.StatusOK, presenceResponse{
		Status:   presence.Status,
		Current:  toLesson(presence.Current),
		Next:     toLesson(presence.Next),
		SyncedAt: presence.SyncedAt,
		Stale:    presence.Stale,
	})
}

func (h *Handler) SyncState(c *gin.Context) {
	state, err := h.useCase.GetSyncState(c.Request.Context())
	if err != nil {
		log.Printf("Failed to get schedule sync state: %s", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Что-то пошло не так, попробуйте ещё раз"})
		return
	}
	c.JSON(http.StatusOK, toSyncState(state))
}

// SyncNow refreshes the schedule by the admin's request. A failed sync is not an error of the request:
// it is returned in the state, like the result of the periodic sync.
func (h *Handler) SyncNow(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), manualSyncTimeout)
	defer cancel()

	if err := h.useCase.SyncNow(ctx); errors.Is(err, schedule.ErrSyncTooOften) {
		c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "Расписание только что обновлялось, попробуйте через минуту"})
		return
	}
	h.SyncState(c)
}
