package http

import (
	"context"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/kosipov/students/campus"
	"github.com/kosipov/students/labs"
	"log"
	"net/http"
	"strconv"
	"time"
)

// occupancyTimeout fits into the server's write timeout: on a cold cache the rooms
// are fetched a few at a time, and a slow campus must not hang the request forever.
const occupancyTimeout = 25 * time.Second

type Handler struct {
	useCase *labs.UseCase
}

type roomLessonResponse struct {
	Number   int       `json:"number"`
	Title    string    `json:"title"`
	Kind     string    `json:"kind"`
	Teacher  string    `json:"teacher"`
	Groups   string    `json:"groups"`
	StartsAt time.Time `json:"startsAt"`
	EndsAt   time.Time `json:"endsAt"`
}

type roomStatusResponse struct {
	Room      string              `json:"room"`
	Floor     int                 `json:"floor"`
	Busy      bool                `json:"busy"`
	Lesson    *roomLessonResponse `json:"lesson"`
	FetchedAt *time.Time          `json:"fetchedAt"`
	Stale     bool                `json:"stale"`
	Failed    bool                `json:"failed"`
}

type occupancyResponse struct {
	Date      time.Time            `json:"date"`
	Pair      int                  `json:"pair"`
	StartsAt  time.Time            `json:"startsAt"`
	EndsAt    time.Time            `json:"endsAt"`
	FetchedAt *time.Time           `json:"fetchedAt"`
	Rooms     []roomStatusResponse `json:"rooms"`
}

func toRoomLesson(lesson *campus.RoomLesson) *roomLessonResponse {
	if lesson == nil {
		return nil
	}
	return &roomLessonResponse{
		Number:   lesson.Number,
		Title:    lesson.Title,
		Kind:     lesson.Kind,
		Teacher:  lesson.Teacher,
		Groups:   lesson.Groups,
		StartsAt: lesson.StartsAt,
		EndsAt:   lesson.EndsAt,
	}
}

func toRoomStatus(status labs.RoomStatus) roomStatusResponse {
	var fetchedAt *time.Time
	if !status.FetchedAt.IsZero() {
		fetchedAt = &status.FetchedAt
	}
	return roomStatusResponse{
		Room:      status.Room.Name,
		Floor:     status.Room.Floor,
		Busy:      status.Lesson != nil,
		Lesson:    toRoomLesson(status.Lesson),
		FetchedAt: fetchedAt,
		Stale:     status.Stale,
		Failed:    status.Failed,
	}
}

func parseFloor(raw string) (int, error) {
	switch raw {
	case "2":
		return 2, nil
	case "5":
		return 5, nil
	case "all":
		return 0, nil
	default:
		return 0, fmt.Errorf("unknown floor %q", raw)
	}
}

// parseDate reads a "2006-01-02" day; an empty value means today in the university's time zone.
func parseDate(raw string, today time.Time) (time.Time, error) {
	if raw == "" {
		return today, nil
	}
	day, err := time.ParseInLocation("2006-01-02", raw, campus.Location)
	if err != nil {
		return time.Time{}, fmt.Errorf("unknown date %q", raw)
	}
	return day, nil
}

func occupancyError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, labs.ErrUnknownPair):
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Пара должна быть now или числом от 1 до 8"})
	case errors.Is(err, labs.ErrUnknownRoom):
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Неизвестная аудитория"})
	case errors.Is(err, labs.ErrDateOutOfRange):
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Кампус показывает только текущую неделю и три следующие"})
	case errors.Is(err, campus.ErrWeekTooFar):
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Кампус показывает только текущую неделю и три следующие"})
	default:
		log.Printf("Failed to get labs occupancy: %s", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Что-то пошло не так, попробуйте ещё раз"})
	}
}

// Occupancy answers who sits in the labs: GET /api/admin/labs?floor=2|5|all&pair=now|1..8&date=2006-01-02.
func (h *Handler) Occupancy(c *gin.Context) {
	floor, err := parseFloor(c.Query("floor"))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Этаж должен быть 2, 5 или all"})
		return
	}

	pairRaw := c.Query("pair")
	var pair int
	if pairRaw == "now" {
		pair = h.useCase.CurrentPair()
	} else {
		pair, err = strconv.Atoi(pairRaw)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Пара должна быть now или числом от 1 до 8"})
			return
		}
	}

	date, err := parseDate(c.Query("date"), time.Now())
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Дата должна быть вида 2006-01-02"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), occupancyTimeout)
	defer cancel()

	occupancy, err := h.useCase.Occupancy(ctx, floor, date, pair)
	if err != nil {
		occupancyError(c, err)
		return
	}

	rooms := make([]roomStatusResponse, 0, len(occupancy.Rooms))
	var oldest *time.Time
	for _, status := range occupancy.Rooms {
		rooms = append(rooms, toRoomStatus(status))
		if !status.FetchedAt.IsZero() && (oldest == nil || status.FetchedAt.Before(*oldest)) {
			t := status.FetchedAt
			oldest = &t
		}
	}

	c.JSON(http.StatusOK, occupancyResponse{
		Date:      occupancy.Date,
		Pair:      occupancy.Pair,
		StartsAt:  occupancy.StartsAt,
		EndsAt:    occupancy.EndsAt,
		FetchedAt: oldest,
		Rooms:     rooms,
	})
}

type refreshRequest struct {
	Room string `json:"room"`
	Date string `json:"date"`
	Pair int    `json:"pair"`
}

// RefreshRoom drops the cached schedule of the room and loads it again from the campus:
// POST /api/admin/labs/refresh {"room": "245/1", "date": "2006-01-02", "pair": 3}.
func (h *Handler) RefreshRoom(c *gin.Context) {
	var request refreshRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Нужны room, date и pair"})
		return
	}

	date, err := parseDate(request.Date, time.Now())
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Дата должна быть вида 2006-01-02"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), occupancyTimeout)
	defer cancel()

	status, err := h.useCase.RefreshRoom(ctx, request.Room, date, request.Pair)
	if err != nil {
		occupancyError(c, err)
		return
	}
	c.JSON(http.StatusOK, toRoomStatus(*status))
}
