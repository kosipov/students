package campus

import (
	"context"
	"errors"
	"fmt"
	"html"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

// DefaultClassroomsURL is the form the campus site itself uses to show a room's schedule.
const DefaultClassroomsURL = "https://campus.syktsu.ru/schedule/classroom/"

// RoomWeek is a room's schedule for one week.
type RoomWeek struct {
	// Room is the room's full name with the building, e.g. "245/1".
	Room string
	// Start is Monday of the week, midnight in Location.
	Start time.Time
	// Updated is when the university last changed any schedule, as the page says.
	Updated time.Time
	Lessons []RoomLesson
}

type RoomLesson struct {
	Number   int
	StartsAt time.Time
	EndsAt   time.Time
	// Title is the discipline, e.g. "Жилищное право".
	Title string
	// Kind is the lesson type as written on the page: "л.", "пр.", "лаб.".
	Kind string
	// Teacher is the lecturer as written on the page, e.g. "Брусиловский Дмитрий Александрович".
	Teacher string
	Groups  string
}

var roomNameRe = regexp.MustCompile(`для аудитории\s+([^\s<]+)`)

// ErrWeekTooFar means the requested week can't be reached from the current one:
// the site only pages week by week, and the client walks at most maxRoomWeekHops forward.
var ErrWeekTooFar = errors.New("campus: the week is too far")

// maxRoomWeekHops is how many "next week" steps FetchRoomWeek takes: the current week plus three ahead.
const maxRoomWeekHops = 3

// FetchRoomWeek returns the schedule of the week starting on the given Monday, as the site sees it.
// The room must be its full name with the building, e.g. "245/1": the site then answers
// with the schedule right away, without the search step. Weeks ahead are reached through
// the site's "next week" button, with the usual pause between requests.
func (c *Client) FetchRoomWeek(ctx context.Context, room string, weekStart time.Time) (*RoomWeek, error) {
	week, err := c.fetchRoomWeek(ctx, url.Values{"num_aud": {room}, "searchdata": {""}}, room)
	if err != nil {
		return nil, err
	}

	for hops := 0; week.Start.Before(weekStart) && hops < maxRoomWeekHops; hops++ {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(c.pause):
		}
		// The "next week" button sends Monday of the shown week and the room.
		week, err = c.fetchRoomWeek(ctx, url.Values{"next": {week.Start.Format("2006-01-02") + "_" + room}}, room)
		if err != nil {
			return nil, err
		}
	}
	if !week.Start.Equal(weekStart) {
		return nil, fmt.Errorf("%w: week of %s", ErrWeekTooFar, weekStart.Format("02.01.2006"))
	}
	return week, nil
}

func (c *Client) fetchRoomWeek(ctx context.Context, form url.Values, room string) (*RoomWeek, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.classroomsURL, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", userAgent)
	// The relay needs to tell this form from the schedule one; the campus itself ignores the header.
	req.Header.Set("X-Relay-Target", "classroom")
	for name, value := range c.headers {
		req.Header.Set(name, value)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("campus: request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("campus: unexpected status %d", resp.StatusCode)
	}
	page, err := io.ReadAll(io.LimitReader(resp.Body, maxPageSize))
	if err != nil {
		return nil, fmt.Errorf("campus: read page: %w", err)
	}

	return ParseRoomWeek(string(page), room)
}

// ParseRoomWeek reads a weekly schedule page of the room.
func ParseRoomWeek(page string, room string) (*RoomWeek, error) {
	week := weekRe.FindStringSubmatch(page)
	name := roomNameRe.FindStringSubmatch(page)
	table := tableRe.FindStringSubmatch(page)
	if week == nil || name == nil || table == nil {
		return nil, ErrUnexpectedPage
	}
	found := strings.TrimSpace(html.UnescapeString(name[1]))
	// The site echoes the room's name as entered in its database, punctuation included:
	// 236 lives there as "236/1." with a trailing dot.
	if strings.TrimSuffix(found, ".") != room {
		return nil, fmt.Errorf("%w: schedule of %q instead of %q", ErrUnexpectedPage, found, room)
	}

	start, err := time.ParseInLocation("02.01.2006", week[1], Location)
	if err != nil || start.Weekday() != time.Monday {
		return nil, fmt.Errorf("%w: week starts on %q", ErrUnexpectedPage, week[1])
	}

	result := &RoomWeek{Room: strings.TrimSuffix(found, "."), Start: start}
	if updated := updatedRe.FindStringSubmatch(page); updated != nil {
		if t, err := time.ParseInLocation("02.01.2006 15:04:05", updated[1], Location); err == nil {
			result.Updated = t
		}
	}

	var lessons []RoomLesson
	err = walkTable(table[1], start, func(number int, startsAt time.Time, cell string) {
		if lesson, ok := parseRoomLesson(cell); ok {
			lesson.Number = number
			lesson.StartsAt = startsAt
			lesson.EndsAt = startsAt.Add(LessonDuration)
			lessons = append(lessons, lesson)
		}
	})
	if err != nil {
		return nil, err
	}
	result.Lessons = lessons
	return result, nil
}

// parseRoomLesson reads "Жилищное право (пр.), Брусиловский Дмитрий Александрович<br> 1425а-ЮРо".
// An empty cell means the room is free. Anything that doesn't fit the pattern is kept in Title,
// so an unusual entry is still shown.
func parseRoomLesson(cell string) (RoomLesson, bool) {
	var lines []string
	for _, part := range brRe.Split(cell, -1) {
		line := strings.Join(strings.Fields(html.UnescapeString(tagRe.ReplaceAllString(part, " "))), " ")
		if line != "" {
			lines = append(lines, line)
		}
	}
	if len(lines) == 0 {
		return RoomLesson{}, false
	}

	lesson := RoomLesson{Title: lines[0], Groups: strings.Join(lines[1:], ", ")}
	if i := strings.LastIndex(lesson.Title, ", "); i >= 0 {
		lesson.Teacher = strings.TrimSpace(lesson.Title[i+2:])
		lesson.Title = strings.TrimSpace(lesson.Title[:i])
	}
	if m := kindRe.FindStringSubmatch(lesson.Title); m != nil {
		lesson.Title, lesson.Kind = m[1], m[2]
	}
	return lesson, true
}
