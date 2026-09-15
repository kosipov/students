package http

import (
	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
	gsessions "github.com/gorilla/sessions"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"
)

type testStore struct{ *gsessions.CookieStore }

func (s testStore) Options(sessions.Options) {}

// newSessionContext returns a gin context with the sessions middleware applied and its recorder.
func newSessionContext(t *testing.T, cookie string) (*gin.Context, *httptest.ResponseRecorder) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	if cookie != "" {
		c.Request.Header.Set("Cookie", cookie)
	}
	sessions.Sessions("test", testStore{gsessions.NewCookieStore([]byte("test-secret"))})(c)
	return c, recorder
}

func TestUnlockedTasksInSession(t *testing.T) {
	c, recorder := newSessionContext(t, "")
	if readUnlockedTasks(c).IsUnlocked(1, 0) {
		t.Fatal("a new session must have no unlocked tasks")
	}

	if err := saveUnlockedTask(c, 7, 2); err != nil {
		t.Fatalf("saveUnlockedTask() error = %v", err)
	}
	if err := saveUnlockedTask(c, 7, 3); err != nil {
		t.Fatalf("saveUnlockedTask() error = %v", err)
	}

	// Every Save writes the cookie again; the browser keeps the last one.
	cookies := recorder.Header().Values("Set-Cookie")
	next, _ := newSessionContext(t, cookies[len(cookies)-1])
	unlocked := readUnlockedTasks(next)
	if !unlocked.IsUnlocked(7, 3) || unlocked.IsUnlocked(7, 2) || unlocked.IsUnlocked(8, 3) || len(unlocked) != 1 {
		t.Errorf("unlocked = %+v, want only task 7 with the latest password version", unlocked)
	}
}

func TestUnlockedTasksKeepTheLatest(t *testing.T) {
	c, _ := newSessionContext(t, "")
	for id := 1; id <= maxUnlockedTasks+5; id++ {
		if err := saveUnlockedTask(c, id, 0); err != nil {
			t.Fatalf("saveUnlockedTask(%d) error = %v", id, err)
		}
	}

	unlocked := readUnlockedTasks(c)
	if len(unlocked) != maxUnlockedTasks || unlocked.IsUnlocked(5, 0) || !unlocked.IsUnlocked(6, 0) || !unlocked.IsUnlocked(maxUnlockedTasks+5, 0) {
		t.Errorf("kept %d tasks from %d, want the latest %d", len(unlocked), unlocked[0].id, maxUnlockedTasks)
	}
}

func TestClientIP(t *testing.T) {
	cases := []struct {
		name, remote, forwarded, want string
	}{
		{"direct request", "203.0.113.7:5000", "", "203.0.113.7"},
		{"forged header without proxy", "203.0.113.7:5000", "1.2.3.4", "203.0.113.7"},
		{"behind proxy", "10.0.1.5:4000", "198.51.100.9", "198.51.100.9"},
		{"forged entry before proxy's", "10.0.1.5:4000", "1.2.3.4, 198.51.100.9", "198.51.100.9"},
		{"garbage from proxy", "10.0.1.5:4000", "not-an-ip", "10.0.1.5"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
			c.Request.RemoteAddr = tc.remote
			if tc.forwarded != "" {
				c.Request.Header.Set("X-Forwarded-For", tc.forwarded)
			}
			if got := clientIP(c); got != tc.want {
				t.Errorf("clientIP() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestAttemptLimiter(t *testing.T) {
	now := time.Date(2026, 9, 15, 12, 0, 0, 0, time.UTC)
	limiter := newAttemptLimiter(3, 15*time.Minute)
	limiter.now = func() time.Time { return now }

	for i := 0; i < 3; i++ {
		if allowed, _ := limiter.Allow("a"); !allowed {
			t.Fatalf("attempt %d denied, want allowed", i+1)
		}
		limiter.Fail("a")
		now = now.Add(time.Minute)
	}

	allowed, retryAfter := limiter.Allow("a")
	if allowed || retryAfter != 12*time.Minute {
		t.Errorf("Allow() after 3 failures = %v, %v; want denied for 12 minutes", allowed, retryAfter)
	}
	if allowed, _ := limiter.Allow("b"); !allowed {
		t.Error("other keys must not be limited")
	}

	now = now.Add(12 * time.Minute)
	if allowed, _ := limiter.Allow("a"); !allowed {
		t.Error("the oldest failure left the window, want allowed")
	}

	limiter.Fail("a")
	limiter.Reset("a")
	if allowed, _ := limiter.Allow("a"); !allowed {
		t.Error("Reset() must forget failures")
	}
}

func TestAttemptLimiterForgetsOldKeys(t *testing.T) {
	now := time.Date(2026, 9, 15, 12, 0, 0, 0, time.UTC)
	limiter := newAttemptLimiter(3, time.Minute)
	limiter.now = func() time.Time { return now }

	for i := 0; i < 10001; i++ {
		limiter.Fail(strconv.Itoa(i))
	}
	now = now.Add(2 * time.Minute)
	limiter.Fail("fresh")

	if len(limiter.failures) != 1 {
		t.Errorf("limiter keeps %d keys, want only the fresh one", len(limiter.failures))
	}
}
