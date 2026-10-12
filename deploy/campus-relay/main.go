// Command campus-relay forwards the site's requests to campus.syktsu.ru from a Russian address.
//
// The university doesn't answer requests from the site's server, so the site asks this relay
// instead. It replaces the Yandex Cloud Function (deploy/campus-relay/index.py) and is built
// to run as a tiny container, e.g. on Cloud.ru Evolution Container Apps free tier.
//
// It is not an open proxy: it only posts the given form to the two campus addresses,
// and only for callers that know RELAY_SECRET.
//
// Environment variables:
//
//	RELAY_SECRET         - the same value as CAMPUS_RELAY_SECRET of the site; required.
//	CAMPUS_URL           - the schedule page; defaults to the teacher schedule of campus.syktsu.ru.
//	CAMPUS_STUDENTS_URL  - the group roster form; defaults to get_student.php of campus.syktsu.ru.
//	PORT                 - the port to listen on; defaults to 8080.
//
// The caller picks the address with the X-Relay-Target header: "students" for the roster,
// anything else (or none) for the schedule.
package main

import (
	"crypto/subtle"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

const (
	defaultScheduleURL = "https://campus.syktsu.ru/schedule/teacher/"
	defaultStudentsURL = "https://campus.syktsu.ru/php_files/get_student.php"
	// userAgent is a regular desktop browser's, so requests don't point to the site that makes them.
	userAgent      = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/139.0.0.0 Safari/537.36"
	maxRequestBody = 64 << 10
	maxPage        = 4 << 20
	campusTimeout  = 20 * time.Second
)

func main() {
	secret := os.Getenv("RELAY_SECRET")
	if secret == "" {
		log.Fatal("RELAY_SECRET is required")
	}
	targets := map[string]string{
		"schedule": envOr("CAMPUS_URL", defaultScheduleURL),
		"students": envOr("CAMPUS_STUDENTS_URL", defaultStudentsURL),
	}
	client := &http.Client{Timeout: campusTimeout}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Only POST is allowed", http.StatusMethodNotAllowed)
			return
		}
		if subtle.ConstantTimeCompare([]byte(r.Header.Get("X-Relay-Secret")), []byte(secret)) != 1 {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		target, ok := targets[r.Header.Get("X-Relay-Target")]
		if !ok {
			target = targets["schedule"]
		}

		body := http.MaxBytesReader(w, r.Body, maxRequestBody)
		defer body.Close()

		req, err := http.NewRequestWithContext(r.Context(), http.MethodPost, target, body)
		if err != nil {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Set("User-Agent", userAgent)
		req.Header.Set("Accept-Language", "ru-RU,ru;q=0.9")

		resp, err := client.Do(req)
		if err != nil {
			log.Printf("campus did not answer: %s", err)
			http.Error(w, "Campus did not answer", http.StatusGatewayTimeout)
			return
		}
		defer resp.Body.Close()

		page, err := io.ReadAll(io.LimitReader(resp.Body, maxPage))
		if err != nil {
			http.Error(w, "Failed to read the campus answer", http.StatusBadGateway)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(resp.StatusCode)
		w.Write(page)
	})

	port := envOr("PORT", "8080")
	log.Printf("campus-relay is listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func envOr(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
