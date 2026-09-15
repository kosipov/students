package campus

import (
	"errors"
	"fmt"
	"html"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// LessonDuration is the length of a lesson. The page shows only start times; the university's
// timetable (08:30, 10:10, 12:40, 14:20, 16:00, 17:40, 19:20) is 90-minute lessons with breaks.
const LessonDuration = 90 * time.Minute

// Location is the university's time zone: Syktyvkar lives in Moscow time, UTC+3 all year.
// A fixed zone avoids depending on tzdata, which the Alpine image doesn't have.
var Location = time.FixedZone("MSK", 3*60*60)

// ErrUnexpectedPage means the page doesn't look like a teacher's weekly schedule, e.g. the teacher
// wasn't found or the site changed. Stored lessons must not be replaced by such a page.
var ErrUnexpectedPage = errors.New("campus: unexpected schedule page")

// Week is a teacher's schedule for one week.
type Week struct {
	// Start is Monday of the week, midnight in Location.
	Start time.Time
	// Updated is when the university last changed any schedule, as the page says.
	Updated time.Time
	Lessons []Lesson
}

type Lesson struct {
	Number   int
	StartsAt time.Time
	EndsAt   time.Time
	// Title is the discipline, e.g. "Проектирование и разработка веб-приложений".
	Title string
	// Kind is the lesson type as written on the page: "л.", "пр.", "лаб.".
	Kind string
	// Location is the room as written on the page, e.g. "245/1"; Room and Building are its parts when it has that form.
	Location string
	Room     string
	Building string
	Groups   string
}

var (
	weekRe    = regexp.MustCompile(`на неделю c (\d{2}\.\d{2}\.\d{4}) по (\d{2}\.\d{2}\.\d{4})`)
	teacherRe = regexp.MustCompile(`для преподавателя\s+([^<]+?)\s*</h3>`)
	updatedRe = regexp.MustCompile(`Расписание обновлено (\d{2}\.\d{2}\.\d{4} \d{2}:\d{2}:\d{2})`)
	tableRe   = regexp.MustCompile(`(?s)<table[^>]*class="schedule"[^>]*>(.*?)</table>`)
	cellRe    = regexp.MustCompile(`(?s)<td([^>]*)>(.*?)</td>`)
	dayDateRe = regexp.MustCompile(`\((\d{2}\.\d{2}\.\d{4})\)`)
	timeRe    = regexp.MustCompile(`^(\d{1,2}):(\d{2})$`)
	brRe      = regexp.MustCompile(`(?i)<br\s*/?>`)
	tagRe     = regexp.MustCompile(`<[^>]*>`)
	kindRe    = regexp.MustCompile(`^(.*\S)\s*\(([^()]+)\)$`)
	roomRe    = regexp.MustCompile(`^(\S+)/(\S+)$`)
)

// ParseWeek reads a weekly schedule page of the teacher.
func ParseWeek(page string, teacher string) (*Week, error) {
	week := weekRe.FindStringSubmatch(page)
	name := teacherRe.FindStringSubmatch(page)
	table := tableRe.FindStringSubmatch(page)
	if week == nil || name == nil || table == nil {
		return nil, ErrUnexpectedPage
	}
	if strings.TrimSpace(html.UnescapeString(name[1])) != teacher {
		return nil, fmt.Errorf("%w: schedule of %q instead of %q", ErrUnexpectedPage, name[1], teacher)
	}

	start, err := time.ParseInLocation("02.01.2006", week[1], Location)
	if err != nil || start.Weekday() != time.Monday {
		return nil, fmt.Errorf("%w: week starts on %q", ErrUnexpectedPage, week[1])
	}

	result := &Week{Start: start}
	if updated := updatedRe.FindStringSubmatch(page); updated != nil {
		if t, err := time.ParseInLocation("02.01.2006 15:04:05", updated[1], Location); err == nil {
			result.Updated = t
		}
	}

	lessons, err := parseTable(table[1], start)
	if err != nil {
		return nil, err
	}
	result.Lessons = lessons
	return result, nil
}

// parseTable walks the cells in order: the markup has no opening <tr> for most rows, so rows can't be trusted.
// A day starts with a "dayofweek" cell holding the date, followed by triples of cells: number, start time, lesson.
func parseTable(table string, weekStart time.Time) ([]Lesson, error) {
	weekEnd := weekStart.AddDate(0, 0, 7)
	var (
		lessons []Lesson
		day     time.Time
		cells   = cellRe.FindAllStringSubmatch(table, -1)
	)

	for i := 0; i < len(cells); {
		attrs, content := cells[i][1], cells[i][2]
		if strings.Contains(attrs, "dayofweek") {
			date := dayDateRe.FindStringSubmatch(content)
			if date == nil {
				return nil, fmt.Errorf("%w: day without date %q", ErrUnexpectedPage, content)
			}
			parsed, err := time.ParseInLocation("02.01.2006", date[1], Location)
			if err != nil || parsed.Before(weekStart) || !parsed.Before(weekEnd) {
				return nil, fmt.Errorf("%w: day %q outside the week", ErrUnexpectedPage, date[1])
			}
			day = parsed
			i++
			continue
		}

		if day.IsZero() || i+2 >= len(cells) {
			return nil, fmt.Errorf("%w: lesson cells without a day", ErrUnexpectedPage)
		}
		number, err := strconv.Atoi(strings.TrimSpace(cells[i][2]))
		if err != nil {
			return nil, fmt.Errorf("%w: lesson number %q", ErrUnexpectedPage, cells[i][2])
		}
		clock := timeRe.FindStringSubmatch(strings.TrimSpace(cells[i+1][2]))
		if clock == nil {
			return nil, fmt.Errorf("%w: lesson time %q", ErrUnexpectedPage, cells[i+1][2])
		}
		hour, _ := strconv.Atoi(clock[1])
		minute, _ := strconv.Atoi(clock[2])

		if lesson, ok := parseLesson(cells[i+2][2]); ok {
			lesson.Number = number
			lesson.StartsAt = time.Date(day.Year(), day.Month(), day.Day(), hour, minute, 0, 0, Location)
			lesson.EndsAt = lesson.StartsAt.Add(LessonDuration)
			lessons = append(lessons, lesson)
		}
		i += 3
	}
	return lessons, nil
}

// parseLesson reads "Title (kind), 245/1<br> 1435-ИРо". An empty cell means no lesson.
// Anything that doesn't fit the pattern is kept in Title, so an unusual lesson is still shown.
func parseLesson(cell string) (Lesson, bool) {
	var lines []string
	for _, part := range brRe.Split(cell, -1) {
		line := strings.Join(strings.Fields(html.UnescapeString(tagRe.ReplaceAllString(part, " "))), " ")
		if line != "" {
			lines = append(lines, line)
		}
	}
	if len(lines) == 0 {
		return Lesson{}, false
	}

	lesson := Lesson{Title: lines[0], Groups: strings.Join(lines[1:], ", ")}
	if i := strings.LastIndex(lesson.Title, ", "); i >= 0 {
		lesson.Location = strings.TrimSpace(lesson.Title[i+2:])
		lesson.Title = strings.TrimSpace(lesson.Title[:i])
	}
	if m := kindRe.FindStringSubmatch(lesson.Title); m != nil {
		lesson.Title, lesson.Kind = m[1], m[2]
	}
	if m := roomRe.FindStringSubmatch(lesson.Location); m != nil {
		lesson.Room, lesson.Building = m[1], m[2]
	}
	return lesson, true
}
