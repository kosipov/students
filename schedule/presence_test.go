package schedule

import (
	"github.com/kosipov/students/models"
	"testing"
	"time"
)

var msk = time.FixedZone("MSK", 3*60*60)

func at(clock string) time.Time {
	t, err := time.ParseInLocation("2006-01-02 15:04", "2026-09-19 "+clock, msk)
	if err != nil {
		panic(err)
	}
	return t
}

func lesson(number int, start string, room string) models.ScheduleLesson {
	return models.ScheduleLesson{Number: number, StartsAt: at(start), EndsAt: at(start).Add(90 * time.Minute), Location: room}
}

// Saturday 19.09.2026 from the real schedule: 08:30 and 10:10 in 516/1, then 12:40 and 14:20 in 241/1.
var saturday = []models.ScheduleLesson{
	lesson(1, "08:30", "516/1"),
	lesson(2, "10:10", "516/1"),
	lesson(3, "12:40", "241/1"),
	lesson(4, "14:20", "241/1"),
}

func TestComputePresence(t *testing.T) {
	cases := []struct {
		now         string
		lessons     []models.ScheduleLesson
		wantStatus  PresenceStatus
		wantCurrent int // lesson number, 0 for none
		wantNext    int
	}{
		{"07:00", saturday, StatusAway, 0, 1},
		{"08:19", saturday, StatusAway, 0, 1},
		{"08:20", saturday, StatusInClass, 1, 2},
		{"09:59", saturday, StatusInClass, 1, 2},
		{"10:00", saturday, StatusInClass, 2, 3}, // the 10-minute break falls into the arrival margin
		{"11:45", saturday, StatusBreak, 0, 3},   // lunch: 11:40–12:40 is exactly MaxBreak
		{"14:15", saturday, StatusInClass, 4, 0},
		{"15:50", saturday, StatusAway, 0, 0},
		{"12:00", nil, StatusAway, 0, 0},
		// A long gap between lessons means the teacher isn't at the university.
		{"12:00", []models.ScheduleLesson{lesson(1, "08:30", "516/1"), lesson(6, "17:40", "245/1")}, StatusAway, 0, 6},
		// Two groups in one slot: the next lesson is the one after the slot, not the parallel one.
		{"08:40", []models.ScheduleLesson{lesson(1, "08:30", "516/1"), lesson(1, "08:30", "517/1"), lesson(2, "10:10", "516/1")}, StatusInClass, 1, 2},
	}

	for _, tc := range cases {
		status, current, next := ComputePresence(at(tc.now), tc.lessons)
		number := func(l *models.ScheduleLesson) int {
			if l == nil {
				return 0
			}
			return l.Number
		}
		if status != tc.wantStatus || number(current) != tc.wantCurrent || number(next) != tc.wantNext {
			t.Errorf("at %s: %s, current %d, next %d; want %s, %d, %d",
				tc.now, status, number(current), number(next), tc.wantStatus, tc.wantCurrent, tc.wantNext)
		}
	}
}
