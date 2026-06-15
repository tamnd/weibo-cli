// Package weibo is the library behind the weibo command: the HTTP client and
// the typed data models for Weibo public surfaces.
//
// Two hosts are used: weibo.com for the hot-rank board and search suggestions,
// and m.weibo.cn for post detail and comments. Each needs different headers.
// No cookies or authentication are required for any of these surfaces.
package weibo

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

// DefaultUserAgent is a desktop Chrome string for weibo.com endpoints.
const DefaultUserAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36"

// mobileUserAgent is a mobile Safari string for m.weibo.cn endpoints.
const mobileUserAgent = "Mozilla/5.0 (iPhone; CPU iPhone OS 17_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.0 Mobile/15E148 Safari/604.1"

// ErrWalled means the API returned ok:-100, meaning the surface requires a
// logged-in session. User profile and post timeline are examples.
var ErrWalled = errors.New("this surface requires a logged-in Weibo session")

// ErrNotFound means the post or resource does not exist.
var ErrNotFound = errors.New("not found")

// Config holds the tunable client settings.
type Config struct {
	BaseURL       string
	MobileBaseURL string
	UserAgent     string
	// Cookie is an optional "SUB=xxx; SUBP=yyy" string pasted from a logged-in
	// browser session. Without it, user and posts exit 4 (walled).
	Cookie  string
	Rate    time.Duration
	Retries int
	Timeout time.Duration
}

// DefaultConfig returns sensible defaults.
func DefaultConfig() Config {
	return Config{
		BaseURL:       "https://weibo.com",
		MobileBaseURL: "https://m.weibo.cn",
		UserAgent:     DefaultUserAgent,
		Rate:          500 * time.Millisecond,
		Retries:       3,
		Timeout:       30 * time.Second,
	}
}

// Client talks to Weibo's public JSON API.
type Client struct {
	cfg  Config
	http *http.Client
	mu   sync.Mutex
	last time.Time
}

// NewClient builds a Client from cfg.
func NewClient(cfg Config) *Client {
	if cfg.UserAgent == "" {
		cfg.UserAgent = DefaultUserAgent
	}
	if cfg.BaseURL == "" {
		cfg.BaseURL = "https://weibo.com"
	}
	if cfg.MobileBaseURL == "" {
		cfg.MobileBaseURL = "https://m.weibo.cn"
	}
	return &Client{
		cfg:  cfg,
		http: &http.Client{Timeout: cfg.Timeout},
	}
}

// getDesktop performs a GET to a weibo.com endpoint with desktop browser headers.
func (c *Client) getDesktop(ctx context.Context, rawURL string) ([]byte, error) {
	return c.get(ctx, rawURL, c.cfg.UserAgent, "https://weibo.com/", false)
}

// getMobile performs a GET to a m.weibo.cn endpoint with mobile browser headers.
func (c *Client) getMobile(ctx context.Context, rawURL string) ([]byte, error) {
	return c.get(ctx, rawURL, mobileUserAgent, "https://m.weibo.cn/", true)
}

func (c *Client) get(ctx context.Context, rawURL, ua, referer string, mobile bool) ([]byte, error) {
	var lastErr error
	for attempt := 0; attempt <= c.cfg.Retries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(backoff(attempt)):
			}
		}
		b, retry, err := c.do(ctx, rawURL, ua, referer, mobile)
		if err == nil {
			return b, nil
		}
		lastErr = err
		if !retry {
			return nil, err
		}
	}
	return nil, fmt.Errorf("get %s: %w", rawURL, lastErr)
}

func (c *Client) do(ctx context.Context, rawURL, ua, referer string, mobile bool) ([]byte, bool, error) {
	c.pace()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, false, err
	}
	req.Header.Set("User-Agent", ua)
	req.Header.Set("Referer", referer)
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")
	if mobile {
		req.Header.Set("MWeibo-Pwa", "1")
		req.Header.Set("X-Requested-With", "XMLHttpRequest")
	}
	if c.cfg.Cookie != "" {
		req.Header.Set("Cookie", c.cfg.Cookie)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, true, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500 {
		return nil, true, fmt.Errorf("http %d", resp.StatusCode)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, false, fmt.Errorf("http %d", resp.StatusCode)
	}

	b, err := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if err != nil {
		return nil, true, err
	}
	return b, false, nil
}

func (c *Client) pace() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.cfg.Rate <= 0 {
		return
	}
	if wait := c.cfg.Rate - time.Since(c.last); wait > 0 {
		time.Sleep(wait)
	}
	c.last = time.Now()
}

func backoff(attempt int) time.Duration {
	d := time.Duration(attempt) * 500 * time.Millisecond
	if d > 5*time.Second {
		return 5 * time.Second
	}
	return d
}
