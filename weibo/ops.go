package weibo

import (
	"context"
	"net/url"

	"github.com/tamnd/any-cli/kit"
	"github.com/tamnd/any-cli/kit/errs"
)

func registerOps(app *kit.App) {
	registerHot(app)
	registerStatus(app)
	registerComments(app)
	registerSuggest(app)
	registerUser(app)
	registerPosts(app)
}

func effectiveLimit(n, def int) int {
	if n > 0 {
		return n
	}
	return def
}

// --- hot ---

type hotIn struct {
	Limit int      `kit:"flag,inherit" help:"max records"`
	Sess  *Session `kit:"inject"`
}

func registerHot(app *kit.App) {
	kit.Handle(app, kit.OpMeta{
		Name: "hot", Group: "read",
		Summary: "Weibo trending topics (微博热搜榜)",
	}, func(ctx context.Context, in hotIn, emit func(HotItem) error) error {
		limit := effectiveLimit(in.Limit, 30)
		in.Sess.Progressf("fetching hot search board")
		items, err := in.Sess.Client.HotSearch(ctx, limit)
		if err != nil {
			return MapErr(err)
		}
		return emitAll(items, emit)
	})
}

// --- status ---

type statusIn struct {
	ID   string   `kit:"arg" help:"post id or post URL"`
	Sess *Session `kit:"inject"`
}

func registerStatus(app *kit.App) {
	kit.Handle(app, kit.OpMeta{
		Name: "status", Group: "read", Single: true,
		Summary: "One Weibo post",
		Args:    []kit.Arg{{Name: "id-or-url", Help: "post id or post URL"}},
	}, func(ctx context.Context, in statusIn, emit func(Status) error) error {
		id, err := parseStatusID(in.ID)
		if err != nil {
			return errs.Usage("%s", err.Error())
		}
		in.Sess.Progressf("fetching post %s", id)
		s, err := in.Sess.Client.StatusByID(ctx, id)
		if err != nil {
			return MapErr(err)
		}
		return emit(s)
	})
}

// --- comments ---

type commentsIn struct {
	ID    string   `kit:"arg" help:"post id or post URL"`
	Limit int      `kit:"flag,inherit" help:"max records"`
	Sess  *Session `kit:"inject"`
}

func registerComments(app *kit.App) {
	kit.Handle(app, kit.OpMeta{
		Name: "comments", Group: "read", List: true,
		Summary: "Comments under a Weibo post",
		Args:    []kit.Arg{{Name: "id-or-url", Help: "post id or post URL"}},
	}, func(ctx context.Context, in commentsIn, emit func(Comment) error) error {
		id, err := parseStatusID(in.ID)
		if err != nil {
			return errs.Usage("%s", err.Error())
		}
		limit := effectiveLimit(in.Limit, 20)
		in.Sess.Progressf("fetching comments for post %s", id)
		comments, err := in.Sess.Client.Comments(ctx, id, limit)
		if err != nil {
			return MapErr(err)
		}
		return emitAll(comments, emit)
	})
}

// --- suggest ---

type suggestIn struct {
	Query []string `kit:"arg,variadic" help:"search terms"`
	Limit int      `kit:"flag,inherit" help:"max records"`
	Sess  *Session `kit:"inject"`
}

func registerSuggest(app *kit.App) {
	kit.Handle(app, kit.OpMeta{
		Name: "suggest", Group: "read",
		Summary: "Search autocomplete suggestions",
		Args:    []kit.Arg{{Name: "query", Help: "search terms", Variadic: true}},
	}, func(ctx context.Context, in suggestIn, emit func(Suggestion) error) error {
		if len(in.Query) == 0 {
			return errs.Usage("suggest requires a query")
		}
		q := url.QueryEscape(joinQuery(in.Query))
		limit := effectiveLimit(in.Limit, 10)
		in.Sess.Progressf("fetching suggestions for %q", joinQuery(in.Query))
		sugs, err := in.Sess.Client.Suggest(ctx, q, limit)
		if err != nil {
			return MapErr(err)
		}
		return emitAll(sugs, emit)
	})
}

// --- user ---

type userIn struct {
	UID  string   `kit:"arg" help:"numeric user id or profile URL"`
	Sess *Session `kit:"inject"`
}

func registerUser(app *kit.App) {
	kit.Handle(app, kit.OpMeta{
		Name: "user", Group: "read", Single: true,
		Summary: "Weibo user profile (requires --cookie)",
		Args:    []kit.Arg{{Name: "uid-or-url", Help: "numeric user id or profile URL"}},
	}, func(ctx context.Context, in userIn, emit func(User) error) error {
		uid, err := parseUID(in.UID)
		if err != nil {
			return errs.Usage("%s", err.Error())
		}
		in.Sess.Progressf("fetching profile for uid %s", uid)
		u, err := in.Sess.Client.UserByID(ctx, uid)
		if err != nil {
			return MapErr(err)
		}
		return emit(u)
	})
}

// --- posts ---

type postsIn struct {
	UID   string   `kit:"arg" help:"numeric user id or profile URL"`
	Page  int      `kit:"flag" help:"page number (1-based)"`
	Limit int      `kit:"flag,inherit" help:"max records"`
	Sess  *Session `kit:"inject"`
}

func registerPosts(app *kit.App) {
	kit.Handle(app, kit.OpMeta{
		Name: "posts", Group: "read",
		Summary: "Posts from a user's timeline (requires --cookie)",
		Args:    []kit.Arg{{Name: "uid-or-url", Help: "numeric user id or profile URL"}},
	}, func(ctx context.Context, in postsIn, emit func(Post) error) error {
		uid, err := parseUID(in.UID)
		if err != nil {
			return errs.Usage("%s", err.Error())
		}
		page := in.Page
		if page < 1 {
			page = 1
		}
		limit := effectiveLimit(in.Limit, 10)
		in.Sess.Progressf("fetching posts for uid %s page %d", uid, page)
		posts, err := in.Sess.Client.PostsByUID(ctx, uid, page, limit)
		if err != nil {
			return MapErr(err)
		}
		return emitAll(posts, emit)
	})
}

func joinQuery(parts []string) string {
	out := ""
	for i, p := range parts {
		if i > 0 {
			out += " "
		}
		out += p
	}
	return out
}

func emitAll[T any](items []T, emit func(T) error) error {
	for _, it := range items {
		if err := emit(it); err != nil {
			return err
		}
	}
	return nil
}
