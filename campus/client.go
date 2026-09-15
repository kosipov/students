// Package campus reads a teacher's schedule from campus.syktsu.ru. The site has no API, so the client
// submits the same forms as the schedule page in a browser and parses the HTML.
package campus

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	DefaultScheduleURL = "https://campus.syktsu.ru/schedule/teacher/"
	// userAgent is a regular desktop browser's, so requests don't point to the site that makes them.
	userAgent      = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/139.0.0.0 Safari/537.36"
	requestTimeout = 20 * time.Second
	maxPageSize    = 2 << 20
	// requestPause spaces out the two page requests of one sync, to be gentle with the site.
	requestPause = 2 * time.Second
)

type Client struct {
	httpClient  *http.Client
	scheduleURL string
	teacher     string
	pause       time.Duration
}

type Option func(*Client)

// WithScheduleURL overrides the schedule page, e.g. with a test server.
func WithScheduleURL(scheduleURL string) Option {
	return func(c *Client) { c.scheduleURL = scheduleURL }
}

func WithPause(pause time.Duration) Option {
	return func(c *Client) { c.pause = pause }
}

func NewClient(teacher string, opts ...Option) *Client {
	c := &Client{
		httpClient:  &http.Client{Timeout: requestTimeout},
		scheduleURL: DefaultScheduleURL,
		teacher:     teacher,
		pause:       requestPause,
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// FetchWeeks returns the current and the next week of the teacher's schedule, as the site sees them.
func (c *Client) FetchWeeks(ctx context.Context) ([]Week, error) {
	// Choosing the teacher opens the current week.
	current, err := c.fetchWeek(ctx, url.Values{"name": {c.teacher}})
	if err != nil {
		return nil, fmt.Errorf("current week: %w", err)
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(c.pause):
	}

	// The "next week" button sends Monday of the shown week and the teacher.
	next, err := c.fetchWeek(ctx, url.Values{"next": {current.Start.Format("2006-01-02") + "_" + c.teacher}})
	if err != nil {
		return nil, fmt.Errorf("next week: %w", err)
	}
	if !next.Start.Equal(current.Start.AddDate(0, 0, 7)) {
		return nil, fmt.Errorf("%w: next week starts %s after %s", ErrUnexpectedPage, next.Start.Format("02.01.2006"), current.Start.Format("02.01.2006"))
	}

	return []Week{*current, *next}, nil
}

func (c *Client) fetchWeek(ctx context.Context, form url.Values) (*Week, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.scheduleURL, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", userAgent)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("campus: request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("campus: unexpected status %d", resp.StatusCode)
	}
	page, err := io.ReadAll(io.LimitReader(resp.Body, maxPageSize))
	if err != nil {
		return nil, fmt.Errorf("campus: read page: %w", err)
	}

	return ParseWeek(string(page), c.teacher)
}
