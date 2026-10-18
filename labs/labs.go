// Package labs tells which computer labs are busy according to the university's schedule.
package labs

import (
	"errors"
	"time"

	"github.com/kosipov/students/campus"
)

// ErrUnknownPair means the pair number is outside the university's timetable.
var ErrUnknownPair = errors.New("labs: unknown pair number")

// ErrUnknownRoom means the room is not one of the institute's labs.
var ErrUnknownRoom = errors.New("labs: unknown room")

// ErrDateOutOfRange means the date's week can't be reached on the campus site:
// only the current week and a few ahead can be paged to.
var ErrDateOutOfRange = errors.New("labs: the date is out of reach")

// weeksAhead is how many weeks past the current one the campus site is asked for.
const weeksAhead = 3

// Room is a computer lab; Name is its full campus name with the building, e.g. "245/1".
type Room struct {
	Name  string
	Floor int
}

// Rooms are the institute's computer labs, all in building 1.
var Rooms = []Room{
	{Name: "236/1", Floor: 2},
	{Name: "241/1", Floor: 2},
	{Name: "243/1", Floor: 2},
	{Name: "245/1", Floor: 2},
	{Name: "247/1", Floor: 2},
	{Name: "249/1", Floor: 2},
	{Name: "251/1", Floor: 2},
	{Name: "502/1", Floor: 5},
	{Name: "503/1", Floor: 5},
	{Name: "515/1", Floor: 5},
	{Name: "516/1", Floor: 5},
	{Name: "517/1", Floor: 5},
	{Name: "518/1", Floor: 5},
	{Name: "519/1", Floor: 5},
}

// pairStarts is the university's timetable: when each pair begins, MSK. A pair is 90 minutes.
var pairStarts = []string{"08:30", "10:10", "12:40", "14:20", "16:00", "17:40", "19:20", "21:00"}

// PairCount is how many pairs a day can have.
const PairCount = 8

// arrivalMargin mirrors schedule.ArrivalMargin: a room counts as taken a little before the pair starts.
const arrivalMargin = 10 * time.Minute

// PairBounds returns when the pair starts and ends on the given day.
func PairBounds(day time.Time, pair int) (time.Time, time.Time) {
	start, _ := time.ParseInLocation("15:04", pairStarts[pair-1], campus.Location)
	startsAt := time.Date(day.Year(), day.Month(), day.Day(), start.Hour(), start.Minute(), 0, 0, campus.Location)
	return startsAt, startsAt.Add(campus.LessonDuration)
}

// PairAt returns the pair to show for "now": the one going on, or the next one during a break.
// Before the first pair it is 1, after the last one it is the last pair.
func PairAt(now time.Time) int {
	local := now.In(campus.Location)
	day := time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, campus.Location)
	for pair := 1; pair <= PairCount; pair++ {
		startsAt, endsAt := PairBounds(day, pair)
		if local.Before(startsAt.Add(-arrivalMargin)) {
			return pair
		}
		if local.Before(endsAt) {
			return pair
		}
	}
	return PairCount
}

// FindRoom returns the lab by its full campus name, e.g. "245/1".
func FindRoom(name string) (Room, bool) {
	for _, room := range Rooms {
		if room.Name == name {
			return room, true
		}
	}
	return Room{}, false
}

// weekStart returns Monday of the week's day, midnight in the university's time zone.
func weekStart(day time.Time) time.Time {
	local := day.In(campus.Location)
	midnight := time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, campus.Location)
	return midnight.AddDate(0, 0, -(int(midnight.Weekday())+6)%7)
}

// sameDay tells whether the moment is on the given calendar day.
func sameDay(t time.Time, day time.Time) bool {
	local := t.In(campus.Location)
	return local.Year() == day.Year() && local.Month() == day.Month() && local.Day() == day.Day()
}
