package campus

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// The room page was saved from campus.syktsu.ru in October 2026: the schedule of 245/1.
func TestParseRoomWeek(t *testing.T) {
	week, err := ParseRoomWeek(readPage(t, "room_week.html"), "245/1")
	if err != nil {
		t.Fatalf("ParseRoomWeek() error = %v", err)
	}

	if week.Room != "245/1" {
		t.Errorf("Room = %q, want 245/1", week.Room)
	}
	if !week.Start.Equal(msk("12.10.2026", "00:00")) {
		t.Errorf("Start = %v, want Monday 12.10.2026", week.Start)
	}
	if !week.Updated.Equal(time.Date(2026, 10, 17, 12, 2, 57, 0, Location)) {
		t.Errorf("Updated = %v, want 17.10.2026 12:02:57", week.Updated)
	}
	if len(week.Lessons) != 33 {
		t.Fatalf("got %d lessons, want 33", len(week.Lessons))
	}

	first := week.Lessons[0]
	want := RoomLesson{
		Number:   1,
		StartsAt: msk("12.10.2026", "08:30"),
		EndsAt:   msk("12.10.2026", "10:00"),
		Title:    "Жилищное право",
		Kind:     "пр.",
		Teacher:  "Брусиловский Дмитрий Александрович",
		Groups:   "1425а-ЮРо",
	}
	if first != want {
		t.Errorf("first lesson = %+v, want %+v", first, want)
	}

	// Two groups share the room, and the teacher follows a comma inside the discipline's name.
	for _, lesson := range week.Lessons {
		if lesson.Number == 6 && lesson.StartsAt.Equal(msk("14.10.2026", "17:40")) {
			if lesson.Title != "Основы алгоритмизации и программирования" || lesson.Kind != "л." ||
				lesson.Teacher != "Канов Иван Дмитриевич" || lesson.Groups != "1415-ИТо, 1415-САо" {
				t.Errorf("Wednesday lesson = %+v", lesson)
			}
			return
		}
	}
	t.Error("Wednesday lesson of the 6th pair not found")
}

func TestParseRoomWeekRejectsOtherPages(t *testing.T) {
	cases := map[string]string{
		"search results without schedule": readPage(t, "search_results.html"),
		"empty page":                      "",
	}
	for name, page := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := ParseRoomWeek(page, "245/1"); !errors.Is(err, ErrUnexpectedPage) {
				t.Errorf("ParseRoomWeek() error = %v, want ErrUnexpectedPage", err)
			}
		})
	}

	if _, err := ParseRoomWeek(readPage(t, "room_week.html"), "245/8"); !errors.Is(err, ErrUnexpectedPage) {
		t.Errorf("ParseRoomWeek() of another room error = %v, want ErrUnexpectedPage", err)
	}
}

func TestParseRoomLesson(t *testing.T) {
	cases := []struct {
		cell string
		want RoomLesson
	}{
		{"", RoomLesson{}},
		{"Информационные системы (пр.), Захаров Иван Дмитриевич<br>131б-ПИо (подгруппа 2)",
			RoomLesson{Title: "Информационные системы", Kind: "пр.", Teacher: "Захаров Иван Дмитриевич", Groups: "131б-ПИо (подгруппа 2)"}},
		{"Пересдача Основы российской государственности (пр.), Пинаевский Дмитрий Иванович",
			RoomLesson{Title: "Пересдача Основы российской государственности", Kind: "пр.", Teacher: "Пинаевский Дмитрий Иванович"}},
		{"Стандартизация, сертификация и техническое документоведение (пр.), Миронов Михаил Валериевич<br>1435-САо",
			RoomLesson{Title: "Стандартизация, сертификация и техническое документоведение", Kind: "пр.", Teacher: "Миронов Михаил Валериевич", Groups: "1435-САо"}},
		{"Звонок &quot;другу&quot;", RoomLesson{Title: `Звонок "другу"`}},
	}
	for _, tc := range cases {
		got, ok := parseRoomLesson(tc.cell)
		if ok != (tc.want != RoomLesson{}) || got != tc.want {
			t.Errorf("parseRoomLesson(%q) = %+v, %v; want %+v", tc.cell, got, ok, tc.want)
		}
	}
}

func TestFetchRoomWeek(t *testing.T) {
	var gotForm, gotTarget string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.Header.Get("User-Agent") != userAgent {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		_ = r.ParseForm()
		gotForm = r.PostForm.Encode()
		gotTarget = r.Header.Get("X-Relay-Target")
		_, _ = w.Write([]byte(readPage(t, "room_week.html")))
	}))
	defer srv.Close()

	week, err := NewClient(teacher, WithClassroomsURL(srv.URL)).FetchRoomWeek(context.Background(), "245/1")
	if err != nil {
		t.Fatalf("FetchRoomWeek() error = %v", err)
	}
	if len(week.Lessons) != 33 {
		t.Errorf("got %d lessons, want 33", len(week.Lessons))
	}
	if gotForm != "num_aud=245%2F1&searchdata=" {
		t.Errorf("form = %q, want the room search form", gotForm)
	}
	if gotTarget != "classroom" {
		t.Errorf("X-Relay-Target = %q, want classroom", gotTarget)
	}
}

func TestFetchRoomWeekFailsOnServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
	}))
	defer srv.Close()

	if _, err := NewClient(teacher, WithClassroomsURL(srv.URL)).FetchRoomWeek(context.Background(), "245/1"); err == nil {
		t.Error("FetchRoomWeek() error = nil, want status 502")
	}
}
