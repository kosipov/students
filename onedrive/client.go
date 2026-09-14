// Package onedrive downloads markdown files from public OneDrive sharing links.
//
// Personal OneDrive no longer serves shared files to anonymous API requests. The client
// does what the OneDrive web viewer does for a visitor who is not signed in: it obtains
// a guest ("badger") token and reads the file through the shares API. This API is
// undocumented, so it may change without notice. Callers are expected to keep the last
// downloaded copy and treat errors as temporary.
package onedrive

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/kosipov/students/educational"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"
	"sync"
	"time"
)

const (
	DefaultTokenURL = "https://api-badgerp.svc.ms/v1.0/token"
	DefaultAPIURL   = "https://api.onedrive.com/v1.0"
	DefaultMaxSize  = 1 << 20

	// badgerAppID is the application id of the OneDrive web app, used to request guest tokens.
	badgerAppID = "5cbed6ac-a083-4e14-b191-b4ba07653de2"
	// tokenRefreshMargin renews a token before it expires, so it doesn't expire mid-request.
	tokenRefreshMargin = 5 * time.Minute
	// defaultTokenTTL is used when the token endpoint doesn't return a parsable expiry time.
	defaultTokenTTL = time.Hour
	requestTimeout  = 10 * time.Second
)

var (
	ErrNotFound = errors.New("onedrive: file not found or link is not shared")
	ErrTooLarge = errors.New("onedrive: file is too large")
)

type Client struct {
	httpClient *http.Client
	tokenURL   string
	apiURL     string
	maxSize    int64
	now        func() time.Time

	mu          sync.Mutex
	token       string
	tokenExpiry time.Time
}

type Option func(*Client)

func WithHTTPClient(httpClient *http.Client) Option {
	return func(c *Client) { c.httpClient = httpClient }
}

// WithEndpoints overrides OneDrive endpoints, e.g. with a test server.
func WithEndpoints(tokenURL, apiURL string) Option {
	return func(c *Client) {
		c.tokenURL = tokenURL
		c.apiURL = strings.TrimRight(apiURL, "/")
	}
}

func WithMaxSize(maxSize int64) Option {
	return func(c *Client) { c.maxSize = maxSize }
}

func NewClient(opts ...Option) *Client {
	c := &Client{
		httpClient: &http.Client{Timeout: requestTimeout},
		tokenURL:   DefaultTokenURL,
		apiURL:     DefaultAPIURL,
		maxSize:    DefaultMaxSize,
		now:        time.Now,
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

func (c *Client) Supports(href string) bool {
	u, err := url.Parse(strings.TrimSpace(href))
	if err != nil || (u.Scheme != "https" && u.Scheme != "http") {
		return false
	}

	switch strings.ToLower(u.Hostname()) {
	case "1drv.ms", "onedrive.live.com":
		return true
	default:
		return false
	}
}

type driveItem struct {
	Name string `json:"name"`
	Size int64  `json:"size"`
	ETag string `json:"eTag"`
	File *struct {
		MimeType string `json:"mimeType"`
	} `json:"file"`
}

func (c *Client) Fetch(ctx context.Context, href string, etag string) (*educational.Content, error) {
	itemPath := "/shares/" + encodeSharingURL(strings.TrimSpace(href)) + "/driveitem"

	var item driveItem
	if err := c.getJSON(ctx, itemPath+"?$select=name,size,eTag,file", &item); err != nil {
		return nil, err
	}

	if !isMarkdown(item) {
		return nil, educational.ErrContentUnsupported
	}
	if etag != "" && item.ETag == etag {
		return nil, educational.ErrContentNotModified
	}
	if item.Size > c.maxSize {
		return nil, fmt.Errorf("%w: %d bytes", ErrTooLarge, item.Size)
	}

	body, err := c.download(ctx, itemPath+"/content")
	if err != nil {
		return nil, err
	}

	return &educational.Content{Body: body, ETag: item.ETag}, nil
}

func isMarkdown(item driveItem) bool {
	if item.File == nil {
		// Folders and other non-file items.
		return false
	}
	if strings.EqualFold(item.File.MimeType, "text/markdown") {
		return true
	}

	switch strings.ToLower(path.Ext(item.Name)) {
	case ".md", ".markdown":
		return true
	default:
		return false
	}
}

// encodeSharingURL converts a sharing link into a token accepted by the shares API.
func encodeSharingURL(sharingURL string) string {
	return "u!" + base64.RawURLEncoding.EncodeToString([]byte(sharingURL))
}

func (c *Client) getJSON(ctx context.Context, apiPath string, out interface{}) error {
	resp, err := c.get(ctx, apiPath)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(out); err != nil {
		return fmt.Errorf("onedrive: decode response: %w", err)
	}
	return nil
}

func (c *Client) download(ctx context.Context, apiPath string) ([]byte, error) {
	resp, err := c.get(ctx, apiPath)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, c.maxSize+1))
	if err != nil {
		return nil, fmt.Errorf("onedrive: read file: %w", err)
	}
	if int64(len(body)) > c.maxSize {
		return nil, ErrTooLarge
	}
	return body, nil
}

// get performs an authorized request. On 401 the token is renewed and the request is retried once.
// The caller must close the body of the returned response, which always has status 200.
func (c *Client) get(ctx context.Context, apiPath string) (*http.Response, error) {
	for attempt := 1; ; attempt++ {
		token, err := c.accessToken(ctx)
		if err != nil {
			return nil, err
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.apiURL+apiPath, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Authorization", "Badger "+token)
		// Lets a guest open the link without redeeming it in a browser first.
		req.Header.Set("Prefer", "autoredeem")

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("onedrive: request: %w", err)
		}

		switch {
		case resp.StatusCode == http.StatusOK:
			return resp, nil
		case resp.StatusCode == http.StatusUnauthorized && attempt == 1:
			drainAndClose(resp)
			c.resetToken()
			continue
		case resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusForbidden:
			drainAndClose(resp)
			return nil, ErrNotFound
		default:
			drainAndClose(resp)
			return nil, fmt.Errorf("onedrive: unexpected status %d", resp.StatusCode)
		}
	}
}

func (c *Client) accessToken(ctx context.Context) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.token != "" && c.now().Add(tokenRefreshMargin).Before(c.tokenExpiry) {
		return c.token, nil
	}

	payload, err := json.Marshal(map[string]string{"appId": badgerAppID})
	if err != nil {
		return "", err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.tokenURL, bytes.NewReader(payload))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("onedrive: token request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("onedrive: token request: unexpected status %d", resp.StatusCode)
	}

	var body struct {
		Token         string `json:"token"`
		ExpiryTimeUTC string `json:"expiryTimeUtc"`
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&body); err != nil {
		return "", fmt.Errorf("onedrive: decode token: %w", err)
	}
	if body.Token == "" {
		return "", errors.New("onedrive: empty token")
	}

	expiry, err := time.Parse(time.RFC3339Nano, body.ExpiryTimeUTC)
	if err != nil {
		expiry = c.now().Add(defaultTokenTTL)
	}

	c.token = body.Token
	c.tokenExpiry = expiry
	return c.token, nil
}

func (c *Client) resetToken() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.token = ""
}

func drainAndClose(resp *http.Response) {
	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 1<<16))
	resp.Body.Close()
}
