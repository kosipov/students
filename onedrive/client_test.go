package onedrive

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/kosipov/students/educational"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

const testSharingURL = "https://1drv.ms/t/c/4127add944749bb9/IQBqpR55u0pRT6aOxRdTWyrAAY9qMa9Raj__V7osLDoFy44?e=oKBhrw"

type fakeOneDrive struct {
	item        driveItem
	body        string
	tokenCalls  int32
	rejectToken string
}

func (f *fakeOneDrive) server(t *testing.T) *httptest.Server {
	t.Helper()

	itemPath := "/api/shares/" + encodeSharingURL(testSharingURL) + "/driveitem"
	mux := http.NewServeMux()

	mux.HandleFunc("/token", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		n := atomic.AddInt32(&f.tokenCalls, 1)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"token":         "token-" + string(rune('0'+n)),
			"expiryTimeUtc": time.Now().Add(7 * 24 * time.Hour).UTC().Format("2006-01-02T15:04:05.0000000Z"),
		})
	})

	authorized := func(w http.ResponseWriter, r *http.Request) bool {
		auth := r.Header.Get("Authorization")
		if !strings.HasPrefix(auth, "Badger ") || auth == "Badger "+f.rejectToken {
			w.WriteHeader(http.StatusUnauthorized)
			return false
		}
		return true
	}

	mux.HandleFunc(itemPath, func(w http.ResponseWriter, r *http.Request) {
		if authorized(w, r) {
			_ = json.NewEncoder(w).Encode(f.item)
		}
	})
	mux.HandleFunc(itemPath+"/content", func(w http.ResponseWriter, r *http.Request) {
		if authorized(w, r) {
			_, _ = w.Write([]byte(f.body))
		}
	})

	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv
}

func newTestClient(srv *httptest.Server, opts ...Option) *Client {
	return NewClient(append([]Option{WithEndpoints(srv.URL+"/token", srv.URL+"/api")}, opts...)...)
}

func markdownItem(etag string, size int64) driveItem {
	return driveItem{
		Name: "Темы курсовых проектов.md",
		Size: size,
		ETag: etag,
		File: &struct {
			MimeType string `json:"mimeType"`
		}{MimeType: "text/markdown"},
	}
}

func TestSupports(t *testing.T) {
	client := NewClient()

	cases := map[string]bool{
		testSharingURL: true,
		"https://onedrive.live.com/:t:/g/personal/4127add944749bb9/IQBqpR55u0pRT6aOxRdTWyrAAY9qMa9Raj__V7osLDoFy44": true,
		"https://docs.google.com/document/d/1": false,
		"https://example.com/1drv.ms/file.md":  false,
		"javascript:alert(1)//1drv.ms":         false,
		"":                                     false,
	}
	for href, want := range cases {
		if got := client.Supports(href); got != want {
			t.Errorf("Supports(%q) = %v, want %v", href, got, want)
		}
	}
}

func TestFetch(t *testing.T) {
	fake := &fakeOneDrive{item: markdownItem("v1", 12), body: "# Задание\n"}
	client := newTestClient(fake.server(t))

	content, err := client.Fetch(context.Background(), testSharingURL, "")
	if err != nil {
		t.Fatalf("Fetch() error = %v", err)
	}
	if string(content.Body) != fake.body || content.ETag != "v1" {
		t.Fatalf("Fetch() = %q, %q", content.Body, content.ETag)
	}

	if _, err := client.Fetch(context.Background(), testSharingURL, ""); err != nil {
		t.Fatalf("second Fetch() error = %v", err)
	}
	if fake.tokenCalls != 1 {
		t.Errorf("token requested %d times, want it to be reused", fake.tokenCalls)
	}
}

func TestFetchNotModified(t *testing.T) {
	fake := &fakeOneDrive{item: markdownItem("v1", 12), body: "# Задание\n"}
	client := newTestClient(fake.server(t))

	_, err := client.Fetch(context.Background(), testSharingURL, "v1")
	if !errors.Is(err, educational.ErrContentNotModified) {
		t.Fatalf("Fetch() error = %v, want ErrContentNotModified", err)
	}
}

func TestFetchUnsupported(t *testing.T) {
	fake := &fakeOneDrive{item: driveItem{Name: "Методичка.pdf", Size: 10, ETag: "v1", File: &struct {
		MimeType string `json:"mimeType"`
	}{MimeType: "application/pdf"}}}
	client := newTestClient(fake.server(t))

	_, err := client.Fetch(context.Background(), testSharingURL, "")
	if !errors.Is(err, educational.ErrContentUnsupported) {
		t.Fatalf("Fetch() error = %v, want ErrContentUnsupported", err)
	}
}

func TestFetchTooLarge(t *testing.T) {
	fake := &fakeOneDrive{item: markdownItem("v1", 100), body: strings.Repeat("a", 100)}
	client := newTestClient(fake.server(t), WithMaxSize(50))

	_, err := client.Fetch(context.Background(), testSharingURL, "")
	if !errors.Is(err, ErrTooLarge) {
		t.Fatalf("Fetch() error = %v, want ErrTooLarge", err)
	}
}

func TestFetchBodyLargerThanReported(t *testing.T) {
	fake := &fakeOneDrive{item: markdownItem("v1", 10), body: strings.Repeat("a", 100)}
	client := newTestClient(fake.server(t), WithMaxSize(50))

	_, err := client.Fetch(context.Background(), testSharingURL, "")
	if !errors.Is(err, ErrTooLarge) {
		t.Fatalf("Fetch() error = %v, want ErrTooLarge", err)
	}
}

func TestFetchRenewsRejectedToken(t *testing.T) {
	fake := &fakeOneDrive{item: markdownItem("v1", 12), body: "# Задание\n", rejectToken: "token-1"}
	client := newTestClient(fake.server(t))

	if _, err := client.Fetch(context.Background(), testSharingURL, ""); err != nil {
		t.Fatalf("Fetch() error = %v", err)
	}
	if fake.tokenCalls != 2 {
		t.Errorf("token requested %d times, want 2 (initial and renewed)", fake.tokenCalls)
	}
}

func TestFetchNotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/token" {
			_, _ = w.Write([]byte(`{"token":"t","expiryTimeUtc":"2099-01-01T00:00:00Z"}`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	_, err := newTestClient(srv).Fetch(context.Background(), testSharingURL, "")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("Fetch() error = %v, want ErrNotFound", err)
	}
}
