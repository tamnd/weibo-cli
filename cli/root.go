// Package cli assembles the weibo command tree on top of the weibo library and
// the any-cli/kit framework. Record-stream commands are kit operations the weibo
// package declares once and exposes as CLI, HTTP, and MCP.
package cli

import (
	"github.com/tamnd/any-cli/kit"
	"github.com/tamnd/weibo-cli/weibo"
)

// Build metadata, set via -ldflags at release time.
var (
	Version = "dev"
	Commit  = "none"
	Date    = "unknown"
)

// NewApp builds the kit application.
func NewApp() *kit.App {
	id := weibo.BaseIdentity()
	id.Version = Version

	app := kit.New(id, kit.WithDefaults(weibo.Defaults))
	weibo.Register(app)

	var userAgent string
	app.GlobalFlags(func(f *kit.FlagSet) {
		f.StringVar(&userAgent, "user-agent", "", "override the User-Agent sent with each request")
	})
	app.Finalize(func(c *kit.Config) {
		if userAgent != "" {
			c.UserAgent = userAgent
		}
	})

	app.AddCommand(newVersionCmd())
	return app
}
