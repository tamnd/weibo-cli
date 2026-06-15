package weibo

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/tamnd/any-cli/kit"
	"github.com/tamnd/any-cli/kit/errs"
)

func init() { kit.Register(Domain{}) }

// Domain is the Weibo driver for the any-cli/kit framework.
type Domain struct{}

// Info describes the scheme and hostnames a pasted link is matched against.
func (Domain) Info() kit.DomainInfo {
	return kit.DomainInfo{
		Scheme:   "weibo",
		Hosts:    []string{"weibo.com", "www.weibo.com", "m.weibo.cn", "s.weibo.com"},
		Identity: BaseIdentity(),
	}
}

// BaseIdentity is the help and version identity shared by the binary and any host.
func BaseIdentity() kit.Identity {
	return kit.Identity{
		Binary: "weibo",
		Short:  "A command line for Weibo (微博).",
		Long: `weibo reads public Weibo (微博) data and prints clean, pipeable records.

It reads hot trending topics, individual posts, comment threads, and search
suggestions through Weibo's public JSON API. No API key or login is required.

Records come out as table, JSON, JSONL, CSV, TSV, url, or raw.

weibo is an independent tool and is not affiliated with Weibo or Sina.`,
		Site: "https://weibo.com",
		Repo: "https://github.com/tamnd/weibo-cli",
	}
}

// Defaults seeds the framework baseline from the weibo defaults.
func Defaults(c *kit.Config) {
	d := DefaultConfig()
	c.Rate = d.Rate
	c.Timeout = d.Timeout
	c.Retries = d.Retries
	c.UserAgent = d.UserAgent
}

// Register installs the client factory and every Weibo operation onto app.
func (Domain) Register(app *kit.App) {
	app.SetClient(newClient)
	registerOps(app)
}

// Register is a convenience so callers don't need to name the zero-value Domain.
func Register(app *kit.App) { Domain{}.Register(app) }

// Session is the per-run client kit injects into every operation.
type Session struct {
	Client *Client
	Quiet  bool
}

// Progressf prints a one-line progress note to stderr unless quiet.
func (s *Session) Progressf(format string, args ...any) {
	if s == nil || s.Quiet {
		return
	}
	_, _ = fmt.Fprintf(os.Stderr, format+"\n", args...)
}

func newClient(_ context.Context, c kit.Config) (any, error) {
	cfg := DefaultConfig()
	if c.UserAgent != "" {
		cfg.UserAgent = c.UserAgent
	}
	if c.Rate > 0 {
		cfg.Rate = c.Rate
	}
	if c.Timeout > 0 {
		cfg.Timeout = c.Timeout
	}
	if c.Retries > 0 {
		cfg.Retries = c.Retries
	}
	if v, ok := c.Extra["cookie"]; ok && v != "" {
		cfg.Cookie = v
	}
	return &Session{Client: NewClient(cfg), Quiet: c.Quiet}, nil
}

// MapErr converts a library error into the kit error kind with the right exit code.
func MapErr(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, ErrWalled):
		return errs.NeedAuth("%s", err.Error())
	case errors.Is(err, ErrNotFound):
		return errs.NotFound("%s", err.Error())
	default:
		return err
	}
}
