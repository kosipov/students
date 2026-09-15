package http

import (
	"sync"
	"time"
)

// attemptLimiter counts failed attempts per key within a sliding window.
type attemptLimiter struct {
	limit  int
	window time.Duration
	now    func() time.Time

	mu       sync.Mutex
	failures map[string][]time.Time
}

func newAttemptLimiter(limit int, window time.Duration) *attemptLimiter {
	return &attemptLimiter{limit: limit, window: window, now: time.Now, failures: map[string][]time.Time{}}
}

// Allow reports whether another attempt is allowed and, if not, when it will be.
func (l *attemptLimiter) Allow(key string) (bool, time.Duration) {
	l.mu.Lock()
	defer l.mu.Unlock()

	failures := l.recent(key)
	if len(failures) < l.limit {
		return true, 0
	}
	return false, failures[0].Add(l.window).Sub(l.now())
}

func (l *attemptLimiter) Fail(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.failures[key] = append(l.recent(key), l.now())
	if len(l.failures) > 10000 {
		l.prune()
	}
}

func (l *attemptLimiter) Reset(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.failures, key)
}

// recent drops failures older than the window. The caller holds the lock.
func (l *attemptLimiter) recent(key string) []time.Time {
	since := l.now().Add(-l.window)
	failures := l.failures[key]
	for len(failures) > 0 && !failures[0].After(since) {
		failures = failures[1:]
	}
	if len(failures) == 0 {
		delete(l.failures, key)
		return nil
	}
	l.failures[key] = failures
	return failures
}

// prune forgets keys without recent failures, so the map doesn't grow forever. The caller holds the lock.
func (l *attemptLimiter) prune() {
	for key := range l.failures {
		l.recent(key)
	}
}
