package campus

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"
	"time"
)

const teacher = "Осипов Константин Сергеевич"

// The testdata pages were saved from campus.syktsu.ru in September 2026.
func readPage(t *testing.T, name string) string {
	t.Helper()
	page, err := os.ReadFile("testdata/" + name)
	if err != nil {
		t.Fatal(err)
	}
	return string(page)
}

func msk(day, clock string) time.Time {
	t, err := time.ParseInLocation("02.01.2006 15:04", day+" "+clock, Location)
	if err != nil {
		panic(err)
	}
	return t
}

func TestParseWeekWithLessons(t *testing.T) {
	week, err := ParseWeek(readPage(t, "week_with_lessons.html"), teacher)
	if err != nil {
		t.Fatalf("ParseWeek() error = %v", err)
	}

	if !week.Start.Equal(msk("14.09.2026", "00:00")) {
		t.Errorf("Start = %v, want Monday 14.09.2026", week.Start)
	}
	if !week.Updated.Equal(time.Date(2026, 9, 15, 18, 28, 12, 0, Location)) {
		t.Errorf("Updated = %v, want 15.09.2026 18:28:12", week.Updated)
	}
	if len(week.Lessons) != 10 {
		t.Fatalf("got %d lessons, want 10", len(week.Lessons))
	}

	first := week.Lessons[0]
	want := Lesson{
		Number:   6,
		StartsAt: msk("15.09.2026", "17:40"),
		EndsAt:   msk("15.09.2026", "19:10"),
		Title:    "Проектирование и разработка веб-приложений",
		Kind:     "л.",
		Location: "245/1",
		Room:     "245",
		Building: "1",
		Groups:   "1435-ИРо",
	}
	if first != want {
		t.Errorf("first lesson = %+v, want %+v", first, want)
	}

	saturday := week.Lessons[6]
	if saturday.Title != "Параллельное программирование" || saturday.Kind != "л." || !saturday.StartsAt.Equal(msk("19.09.2026", "08:30")) ||
		saturday.Location != "516/1" || saturday.Groups != "122-МКо" {
		t.Errorf("Saturday lesson = %+v", saturday)
	}
}

func TestParseEmptyWeek(t *testing.T) {
	week, err := ParseWeek(readPage(t, "empty_week.html"), teacher)
	if err != nil {
		t.Fatalf("ParseWeek() error = %v", err)
	}
	if len(week.Lessons) != 0 || !week.Start.Equal(msk("31.08.2026", "00:00")) {
		t.Errorf("week = %+v, want no lessons from 31.08.2026", week)
	}
}

func TestParseWeekRejectsOtherPages(t *testing.T) {
	cases := map[string]string{
		"search results without schedule": readPage(t, "search_results.html"),
		"empty page":                      "",
		"broken table": strings.Replace(readPage(t, "week_with_lessons.html"),
			`<td>6</td><td>17:40</td>`, `<td>шесть</td><td>17:40</td>`, 1),
	}
	for name, page := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := ParseWeek(page, teacher); !errors.Is(err, ErrUnexpectedPage) {
				t.Errorf("ParseWeek() error = %v, want ErrUnexpectedPage", err)
			}
		})
	}

	if _, err := ParseWeek(readPage(t, "week_with_lessons.html"), "Осипов Дмитрий Анатольевич"); !errors.Is(err, ErrUnexpectedPage) {
		t.Errorf("ParseWeek() of another teacher error = %v, want ErrUnexpectedPage", err)
	}
}

func TestParseLesson(t *testing.T) {
	cases := []struct {
		cell string
		want Lesson
	}{
		{"", Lesson{}},
		{"Физкультура (пр.), спортзал/2<br> 1435-ИРо<br>1425-ИРо", Lesson{Title: "Физкультура", Kind: "пр.", Location: "спортзал/2", Room: "спортзал", Building: "2", Groups: "1435-ИРо, 1425-ИРо"}},
		{"Консультация, дистанционно", Lesson{Title: "Консультация", Location: "дистанционно"}},
		{"Защита &quot;курсовых&quot;", Lesson{Title: `Защита "курсовых"`}},
	}
	for _, tc := range cases {
		got, ok := parseLesson(tc.cell)
		if ok != (tc.want != Lesson{}) || got != tc.want {
			t.Errorf("parseLesson(%q) = %+v, %v; want %+v", tc.cell, got, ok, tc.want)
		}
	}
}

func TestFetchWeeks(t *testing.T) {
	var (
		mu    sync.Mutex
		forms []string
	)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.Header.Get("User-Agent") != userAgent {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		_ = r.ParseForm()
		mu.Lock()
		forms = append(forms, r.PostForm.Encode())
		mu.Unlock()

		switch {
		case r.PostForm.Get("name") == teacher:
			_, _ = w.Write([]byte(readPage(t, "week_with_lessons.html")))
		case r.PostForm.Get("next") == "2026-09-14_"+teacher:
			_, _ = w.Write([]byte(readPage(t, "next_week.html")))
		default:
			_, _ = w.Write([]byte(readPage(t, "search_results.html")))
		}
	}))
	defer srv.Close()

	weeks, err := NewClient(teacher, WithScheduleURL(srv.URL), WithPause(0)).FetchWeeks(context.Background())
	if err != nil {
		t.Fatalf("FetchWeeks() error = %v", err)
	}
	if len(weeks) != 2 || !weeks[0].Start.Equal(msk("14.09.2026", "00:00")) || !weeks[1].Start.Equal(msk("21.09.2026", "00:00")) {
		t.Fatalf("weeks = %+v, want 14.09 and 21.09", weeks)
	}
	if len(weeks[1].Lessons) == 0 {
		t.Error("next week has no lessons")
	}
	if len(forms) != 2 {
		t.Errorf("requests = %v, want 2", forms)
	}
}

func TestFetchWeeksFailsOnUnexpectedPage(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(readPage(t, "search_results.html")))
	}))
	defer srv.Close()

	_, err := NewClient(teacher, WithScheduleURL(srv.URL), WithPause(0)).FetchWeeks(context.Background())
	if !errors.Is(err, ErrUnexpectedPage) {
		t.Errorf("FetchWeeks() error = %v, want ErrUnexpectedPage", err)
	}
}

func TestFetchWeeksFailsOnServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
	}))
	defer srv.Close()

	if _, err := NewClient(teacher, WithScheduleURL(srv.URL)).FetchWeeks(context.Background()); err == nil || !strings.Contains(err.Error(), "502") {
		t.Errorf("FetchWeeks() error = %v, want status 502", err)
	}
}

func TestFetchWeeksThroughProxy(t *testing.T) {
	var (
		mu      sync.Mutex
		proxied []string
	)
	// The campus URL points to a host that doesn't exist: only the proxy can answer.
	proxy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		proxied = append(proxied, r.URL.String())
		mu.Unlock()
		if r.Header.Get("Proxy-Authorization") == "" {
			w.WriteHeader(http.StatusProxyAuthRequired)
			return
		}
		_ = r.ParseForm()
		if r.PostForm.Get("name") == teacher {
			_, _ = w.Write([]byte(readPage(t, "week_with_lessons.html")))
			return
		}
		_, _ = w.Write([]byte(readPage(t, "next_week.html")))
	}))
	defer proxy.Close()

	proxyURL, err := ParseProxyURL(strings.Replace(proxy.URL, "http://", "http://user:secret@", 1))
	if err != nil {
		t.Fatalf("ParseProxyURL() error = %v", err)
	}
	client := NewClient(teacher, WithScheduleURL("http://campus.invalid/schedule/teacher/"), WithPause(0), WithProxy(proxyURL))

	weeks, err := client.FetchWeeks(context.Background())
	if err != nil {
		t.Fatalf("FetchWeeks() through proxy error = %v", err)
	}
	if len(weeks) != 2 || len(proxied) != 2 || proxied[0] != "http://campus.invalid/schedule/teacher/" {
		t.Errorf("weeks = %d, proxied requests = %v; want both requests to the campus URL through the proxy", len(weeks), proxied)
	}
}

func TestParseProxyURL(t *testing.T) {
	for _, valid := range []string{"http://127.0.0.1:3128", "https://user:p%40ss@proxy.example:443", "socks5://proxy.example:1080", " http://proxy.example:8080 "} {
		if _, err := ParseProxyURL(valid); err != nil {
			t.Errorf("ParseProxyURL(%q) error = %v", valid, err)
		}
	}
	for _, invalid := range []string{"proxy.example:3128", "ftp://proxy.example:21", "socks5h://proxy.example:1080", "http://proxy.example", "http://:3128"} {
		if _, err := ParseProxyURL(invalid); err == nil {
			t.Errorf("ParseProxyURL(%q) = nil error, want an error", invalid)
		}
	}
}

func TestFetchWeeksThroughRelay(t *testing.T) {
	var (
		mu      sync.Mutex
		secrets []string
		bodies  []string
	)
	relay := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		mu.Lock()
		secrets = append(secrets, r.Header.Get("X-Relay-Secret"))
		bodies = append(bodies, string(body))
		mu.Unlock()

		if strings.Contains(string(body), "next=") {
			_, _ = w.Write([]byte(readPage(t, "next_week.html")))
			return
		}
		_, _ = w.Write([]byte(readPage(t, "week_with_lessons.html")))
	}))
	defer relay.Close()

	client := NewClient(teacher, WithRelay(relay.URL, "s3cret"), WithPause(0))
	weeks, err := client.FetchWeeks(context.Background())
	if err != nil {
		t.Fatalf("FetchWeeks() through relay error = %v", err)
	}
	if len(weeks) != 2 {
		t.Fatalf("weeks = %d, want 2", len(weeks))
	}
	if secrets[0] != "s3cret" || secrets[1] != "s3cret" {
		t.Errorf("secrets = %v, want the relay secret with every request", secrets)
	}
	if !strings.Contains(bodies[0], "name=") || !strings.Contains(bodies[1], "next=2026-09-14") {
		t.Errorf("bodies = %v, want the same form as the site sends", bodies)
	}
}
