package labs

import (
	"context"
	"sync"
	"time"

	"github.com/kosipov/students/campus"
)

// CacheTTL is how long a fetched room schedule is served without asking the campus again:
// the admin looks a few times a day, and the site must not be bothered more often.
const CacheTTL = 12 * time.Hour

// fetchConcurrency limits how many room pages are requested at once, to be gentle with the site.
const fetchConcurrency = 4

// Source loads a room's weekly schedule; *campus.Client implements it.
type Source interface {
	FetchRoomWeek(ctx context.Context, room string, weekStart time.Time) (*campus.RoomWeek, error)
}

// RoomStatus is the occupancy of one lab at a pair.
type RoomStatus struct {
	Room Room
	// Lesson is what occupies the room; nil means the room is free.
	Lesson *campus.RoomLesson
	// FetchedAt is when the campus answered; zero when there is no data at all.
	FetchedAt time.Time
	// Stale means the campus didn't answer and an older copy is shown.
	Stale bool
	// Failed means the campus didn't answer and there is nothing to show.
	Failed bool
}

// Occupancy is the answer to "who sits in the labs at this pair".
type Occupancy struct {
	Date     time.Time
	Pair     int
	StartsAt time.Time
	EndsAt   time.Time
	Rooms    []RoomStatus
}

// cachedWeek is a room's weekly schedule kept in memory.
type cachedWeek struct {
	week      *campus.RoomWeek
	fetchedAt time.Time
}

// inFlight shares one campus request between concurrent queries of the same room and week.
type inFlight struct {
	done      chan struct{}
	week      *campus.RoomWeek
	fetchedAt time.Time
	stale     bool
	err       error
}

type UseCase struct {
	source Source
	// now and ttl are replaceable in tests.
	now func() time.Time
	ttl time.Duration

	mu       sync.Mutex
	cache    map[string]cachedWeek
	requests map[string]*inFlight
}

func NewUseCase(source Source) *UseCase {
	return &UseCase{
		source:   source,
		now:      time.Now,
		ttl:      CacheTTL,
		cache:    map[string]cachedWeek{},
		requests: map[string]*inFlight{},
	}
}

// CurrentPair is the pair to show for "now".
func (u *UseCase) CurrentPair() int {
	return PairAt(u.now())
}

// today returns the current day in the university's time zone.
func (u *UseCase) today() time.Time {
	now := u.now().In(campus.Location)
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, campus.Location)
}

// checkDate validates the day and returns it at midnight in the university's time zone:
// the campus site pages week by week, so only the current week and a few ahead can be shown.
func (u *UseCase) checkDate(date time.Time) (time.Time, error) {
	local := date.In(campus.Location)
	day := time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, campus.Location)

	current := weekStart(u.today())
	if weekStart(day).Before(current) || weekStart(day).After(current.AddDate(0, 0, 7*weeksAhead)) {
		return time.Time{}, ErrDateOutOfRange
	}
	return day, nil
}

// Occupancy finds who sits in the labs of the floor (0 means all floors) at the pair of the day.
// A room that failed to load doesn't fail the others: it is marked in its status.
func (u *UseCase) Occupancy(ctx context.Context, floor int, date time.Time, pair int) (*Occupancy, error) {
	if pair < 1 || pair > PairCount {
		return nil, ErrUnknownPair
	}
	day, err := u.checkDate(date)
	if err != nil {
		return nil, err
	}
	startsAt, endsAt := PairBounds(day, pair)

	rooms := make([]Room, 0, len(Rooms))
	for _, room := range Rooms {
		if floor == 0 || room.Floor == floor {
			rooms = append(rooms, room)
		}
	}

	statuses := make([]RoomStatus, len(rooms))
	sem := make(chan struct{}, fetchConcurrency)
	var wg sync.WaitGroup
	for i := range rooms {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			statuses[i] = u.roomStatus(ctx, rooms[i], day, pair)
		}(i)
	}
	wg.Wait()

	return &Occupancy{
		Date:     day,
		Pair:     pair,
		StartsAt: startsAt,
		EndsAt:   endsAt,
		Rooms:    statuses,
	}, nil
}

// RefreshRoom drops the cached week of the room and loads it again from the campus:
// the admin's explicit "fetch now" for one room.
func (u *UseCase) RefreshRoom(ctx context.Context, roomName string, date time.Time, pair int) (*RoomStatus, error) {
	room, ok := FindRoom(roomName)
	if !ok {
		return nil, ErrUnknownRoom
	}
	if pair < 1 || pair > PairCount {
		return nil, ErrUnknownPair
	}
	day, err := u.checkDate(date)
	if err != nil {
		return nil, err
	}

	u.mu.Lock()
	delete(u.cache, roomName+" "+weekStart(day).Format("2006-01-02"))
	u.mu.Unlock()

	status := u.roomStatus(ctx, room, day, pair)
	return &status, nil
}

// roomStatus loads the room's week (from the cache when fresh) and finds the lesson of the pair.
func (u *UseCase) roomStatus(ctx context.Context, room Room, day time.Time, pair int) RoomStatus {
	week, fetchedAt, stale, err := u.roomWeek(ctx, room.Name, weekStart(day))
	status := RoomStatus{Room: room, FetchedAt: fetchedAt, Stale: stale}
	if err != nil {
		status.Failed = true
	} else {
		status.Lesson = lessonAt(week, day, pair)
	}
	return status
}

// lessonAt finds the lesson occupying the room at the pair of the day; nil means free.
func lessonAt(week *campus.RoomWeek, day time.Time, pair int) *campus.RoomLesson {
	for i := range week.Lessons {
		if lesson := &week.Lessons[i]; lesson.Number == pair && sameDay(lesson.StartsAt, day) {
			return lesson
		}
	}
	return nil
}

// roomWeek returns the room's schedule of the week, from the cache when it's fresh enough.
func (u *UseCase) roomWeek(ctx context.Context, room string, start time.Time) (*campus.RoomWeek, time.Time, bool, error) {
	key := room + " " + start.Format("2006-01-02")

	u.mu.Lock()
	if entry, ok := u.cache[key]; ok && u.now().Sub(entry.fetchedAt) < u.ttl {
		u.mu.Unlock()
		return entry.week, entry.fetchedAt, false, nil
	}
	if flight, ok := u.requests[key]; ok {
		u.mu.Unlock()
		<-flight.done
		return flight.week, flight.fetchedAt, flight.stale, flight.err
	}
	flight := &inFlight{done: make(chan struct{})}
	u.requests[key] = flight
	u.mu.Unlock()

	flight.week, flight.err = u.source.FetchRoomWeek(ctx, room, start)
	flight.fetchedAt = u.now()

	u.mu.Lock()
	delete(u.requests, key)
	if flight.err == nil {
		u.cache[key] = cachedWeek{week: flight.week, fetchedAt: flight.fetchedAt}
	} else if entry, ok := u.cache[key]; ok {
		// The campus didn't answer: an outdated copy is better than nothing.
		flight.week, flight.fetchedAt, flight.stale, flight.err = entry.week, entry.fetchedAt, true, nil
	}
	u.mu.Unlock()

	close(flight.done)
	return flight.week, flight.fetchedAt, flight.stale, flight.err
}
