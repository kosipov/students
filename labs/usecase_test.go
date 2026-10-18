package labs

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/kosipov/students/campus"
)

func msk(day, clock string) time.Time {
	t, err := time.ParseInLocation("02.01.2006 15:04", day+" "+clock, campus.Location)
	if err != nil {
		panic(err)
	}
	return t
}

type fakeSource struct {
	mu    sync.Mutex
	calls []string
	err   error
	// lessons is what every room returns.
	lessons []campus.RoomLesson
}

func (s *fakeSource) FetchRoomWeek(_ context.Context, room string) (*campus.RoomWeek, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.calls = append(s.calls, room)
	if s.err != nil {
		return nil, s.err
	}
	return &campus.RoomWeek{Room: room, Start: msk("12.10.2026", "00:00"), Lessons: s.lessons}, nil
}

func (s *fakeSource) callCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.calls)
}

// newTestUseCase fixes "now" on Friday 16.10.2026, 12:50 MSK: the 3rd pair is going on.
func newTestUseCase(source Source) *UseCase {
	u := NewUseCase(source)
	u.now = func() time.Time { return msk("16.10.2026", "12:50") }
	return u
}

func TestOccupancy(t *testing.T) {
	source := &fakeSource{lessons: []campus.RoomLesson{
		{Number: 3, StartsAt: msk("16.10.2026", "12:40"), EndsAt: msk("16.10.2026", "14:10"),
			Title: "Базы данных", Kind: "лаб.", Teacher: "Шадрин Лев Сергеевич", Groups: "141-ПМо"},
	}}
	u := newTestUseCase(source)

	occupancy, err := u.Occupancy(context.Background(), 2, 3)
	if err != nil {
		t.Fatalf("Occupancy() error = %v", err)
	}

	if occupancy.Pair != 3 || !occupancy.StartsAt.Equal(msk("16.10.2026", "12:40")) || !occupancy.EndsAt.Equal(msk("16.10.2026", "14:10")) {
		t.Errorf("pair bounds = %d %v–%v, want 3 12:40–14:10", occupancy.Pair, occupancy.StartsAt, occupancy.EndsAt)
	}
	if len(occupancy.Rooms) != 7 {
		t.Fatalf("got %d rooms, want 7 of the 2nd floor", len(occupancy.Rooms))
	}
	for _, status := range occupancy.Rooms {
		if status.Failed || status.Stale {
			t.Errorf("room %s: Failed = %v, Stale = %v, want both false", status.Room.Name, status.Failed, status.Stale)
		}
		if status.Lesson == nil {
			t.Errorf("room %s is free, want busy: the fake source has a lesson for every room", status.Room.Name)
		} else if status.Lesson.Teacher != "Шадрин Лев Сергеевич" {
			t.Errorf("room %s lesson = %+v", status.Room.Name, status.Lesson)
		}
	}
	if source.callCount() != 7 {
		t.Errorf("campus calls = %d, want 7", source.callCount())
	}

	// The same query again is answered from the cache: the campus is not asked twice.
	if _, err := u.Occupancy(context.Background(), 2, 3); err != nil {
		t.Fatalf("second Occupancy() error = %v", err)
	}
	if source.callCount() != 7 {
		t.Errorf("campus calls after a repeat = %d, want still 7", source.callCount())
	}
}

func TestOccupancyFreeRoom(t *testing.T) {
	// The lesson is at another pair, so the room is free at the 3rd.
	source := &fakeSource{lessons: []campus.RoomLesson{
		{Number: 5, StartsAt: msk("16.10.2026", "16:00"), EndsAt: msk("16.10.2026", "17:30"), Title: "Правоведение"},
	}}
	u := newTestUseCase(source)

	occupancy, err := u.Occupancy(context.Background(), 5, 3)
	if err != nil {
		t.Fatalf("Occupancy() error = %v", err)
	}
	if len(occupancy.Rooms) != 7 || occupancy.Rooms[0].Room.Floor != 5 {
		t.Fatalf("got %d rooms, want 7 of the 5th floor", len(occupancy.Rooms))
	}
	for _, status := range occupancy.Rooms {
		if status.Lesson != nil {
			t.Errorf("room %s is busy at %+v, want free at the 3rd pair", status.Room.Name, status.Lesson)
		}
	}
}

func TestOccupancyAllFloors(t *testing.T) {
	source := &fakeSource{}
	u := newTestUseCase(source)

	occupancy, err := u.Occupancy(context.Background(), 0, 1)
	if err != nil {
		t.Fatalf("Occupancy() error = %v", err)
	}
	if len(occupancy.Rooms) != len(Rooms) {
		t.Errorf("got %d rooms, want all %d", len(occupancy.Rooms), len(Rooms))
	}
}

func TestOccupancyCacheExpires(t *testing.T) {
	source := &fakeSource{}
	u := newTestUseCase(source)
	u.ttl = time.Hour

	if _, err := u.Occupancy(context.Background(), 2, 3); err != nil {
		t.Fatalf("Occupancy() error = %v", err)
	}
	// Fifty minutes later the cache is still fresh.
	u.now = func() time.Time { return msk("16.10.2026", "13:40") }
	if _, err := u.Occupancy(context.Background(), 2, 3); err != nil {
		t.Fatalf("Occupancy() error = %v", err)
	}
	if source.callCount() != 7 {
		t.Errorf("campus calls = %d, want 7: the cache is fresh", source.callCount())
	}

	// Two hours after the fetch the cache is outdated: the campus is asked again.
	u.now = func() time.Time { return msk("16.10.2026", "15:00") }
	if _, err := u.Occupancy(context.Background(), 2, 4); err != nil {
		t.Fatalf("Occupancy() error = %v", err)
	}
	if source.callCount() != 14 {
		t.Errorf("campus calls = %d, want 14: the cache expired", source.callCount())
	}
}

func TestOccupancyStaleOnCampusError(t *testing.T) {
	source := &fakeSource{lessons: []campus.RoomLesson{
		{Number: 3, StartsAt: msk("16.10.2026", "12:40"), Title: "Базы данных"},
	}}
	u := newTestUseCase(source)
	u.ttl = time.Hour

	if _, err := u.Occupancy(context.Background(), 2, 3); err != nil {
		t.Fatalf("Occupancy() error = %v", err)
	}

	// The cache expired and the campus doesn't answer: the outdated copy is shown, marked stale.
	source.err = errors.New("campus is down")
	u.now = func() time.Time { return msk("16.10.2026", "15:00") }
	occupancy, err := u.Occupancy(context.Background(), 2, 4)
	if err != nil {
		t.Fatalf("Occupancy() error = %v", err)
	}
	for _, status := range occupancy.Rooms {
		if !status.Stale || status.Failed {
			t.Errorf("room %s: Stale = %v, Failed = %v; want stale data", status.Room.Name, status.Stale, status.Failed)
		}
	}
}

func TestOccupancyFailedWithoutCache(t *testing.T) {
	source := &fakeSource{err: errors.New("campus is down")}
	u := newTestUseCase(source)

	occupancy, err := u.Occupancy(context.Background(), 2, 3)
	if err != nil {
		t.Fatalf("Occupancy() error = %v: one room's failure must not fail the others", err)
	}
	for _, status := range occupancy.Rooms {
		if !status.Failed || status.Lesson != nil {
			t.Errorf("room %s: Failed = %v, Lesson = %+v; want failed without data", status.Room.Name, status.Failed, status.Lesson)
		}
	}
}

func TestOccupancyRejectsUnknownPair(t *testing.T) {
	u := newTestUseCase(&fakeSource{})
	for _, pair := range []int{0, 9, -1} {
		if _, err := u.Occupancy(context.Background(), 2, pair); !errors.Is(err, ErrUnknownPair) {
			t.Errorf("Occupancy(pair %d) error = %v, want ErrUnknownPair", pair, err)
		}
	}
}

func TestPairAt(t *testing.T) {
	cases := []struct {
		now  time.Time
		want int
	}{
		{msk("16.10.2026", "07:00"), 1}, // before the first pair
		{msk("16.10.2026", "08:21"), 1}, // the teacher arrives a little earlier
		{msk("16.10.2026", "09:30"), 1}, // the first pair is going on
		{msk("16.10.2026", "10:05"), 2}, // the break: the next pair is shown
		{msk("16.10.2026", "12:50"), 3}, // the third pair is going on
		{msk("16.10.2026", "21:05"), 8}, // the last pair
		{msk("16.10.2026", "23:40"), 8}, // after the last pair
	}
	for _, tc := range cases {
		if got := PairAt(tc.now); got != tc.want {
			t.Errorf("PairAt(%v) = %d, want %d", tc.now.Format("15:04"), got, tc.want)
		}
	}
}

func TestWeekStart(t *testing.T) {
	if got := weekStart(msk("18.10.2026", "15:00")); !got.Equal(msk("12.10.2026", "00:00")) {
		t.Errorf("weekStart(Sunday) = %v, want Monday 12.10.2026", got)
	}
	if got := weekStart(msk("12.10.2026", "15:00")); !got.Equal(msk("12.10.2026", "00:00")) {
		t.Errorf("weekStart(Monday) = %v, want the same Monday", got)
	}
}
