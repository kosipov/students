package campus

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

// DefaultStudentsURL is the form the campus site itself uses to fill the student list of a group.
const DefaultStudentsURL = "https://campus.syktsu.ru/php_files/get_student.php"

// ErrUnexpectedRoster means the answer doesn't look like a group's roster, e.g. the site changed.
var ErrUnexpectedRoster = errors.New("campus: unexpected students answer")

var optionRe = regexp.MustCompile(`<option\s+value="[^"]*">([^<]*)</option>`)

// studentsResponse is how get_student.php answers: options of a <select> inside a JSON field.
type studentsResponse struct {
	Student string `json:"student"`
}

// FetchStudents returns the full names of the group's students as the campus lists them.
// An unknown group is not an error: the answer just has no students.
func (c *Client) FetchStudents(ctx context.Context, group string) ([]string, error) {
	form := url.Values{"group": {group}, "reg": {"group"}}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.studentsURL, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", userAgent)
	// The relay needs to tell this form from the schedule one; the campus itself ignores the header.
	req.Header.Set("X-Relay-Target", "students")
	for name, value := range c.headers {
		req.Header.Set(name, value)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("campus: request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("campus: unexpected status %d", resp.StatusCode)
	}
	answer, err := io.ReadAll(io.LimitReader(resp.Body, maxPageSize))
	if err != nil {
		return nil, fmt.Errorf("campus: read answer: %w", err)
	}

	return ParseStudents(string(answer))
}

// ParseStudents extracts student names from the campus answer.
func ParseStudents(answer string) ([]string, error) {
	var response studentsResponse
	if err := json.Unmarshal([]byte(answer), &response); err != nil {
		return nil, fmt.Errorf("%w: not a JSON answer", ErrUnexpectedRoster)
	}

	matches := optionRe.FindAllStringSubmatch(response.Student, -1)
	names := make([]string, 0, len(matches))
	for _, match := range matches {
		if name := strings.TrimSpace(html.UnescapeString(match[1])); name != "" {
			names = append(names, name)
		}
	}
	return names, nil
}
