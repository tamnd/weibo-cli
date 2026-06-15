package weibo

import (
	"context"
	"encoding/json"
	"fmt"
)

// HotSearch returns the Weibo hot search board in rank order.
// If limit > 0, only the first limit items are returned.
func (c *Client) HotSearch(ctx context.Context, limit int) ([]HotItem, error) {
	b, err := c.getDesktop(ctx, c.cfg.BaseURL+"/ajax/side/hotSearch")
	if err != nil {
		return nil, err
	}
	var resp hotSearchResponse
	if err := json.Unmarshal(b, &resp); err != nil {
		return nil, fmt.Errorf("decode hot search: %w", err)
	}
	if resp.OK != 1 {
		return nil, ErrWalled
	}
	items := resp.Data.Realtime
	if limit > 0 && limit < len(items) {
		items = items[:limit]
	}
	out := make([]HotItem, 0, len(items))
	for i, w := range items {
		out = append(out, hotFrom(w, i+1))
	}
	return out, nil
}

// StatusByID returns one Weibo post by its numeric id string.
func (c *Client) StatusByID(ctx context.Context, id string) (Status, error) {
	rawURL := c.cfg.MobileBaseURL + "/statuses/show?id=" + id
	b, err := c.getMobile(ctx, rawURL)
	if err != nil {
		return Status{}, err
	}
	var resp statusResponse
	if err := json.Unmarshal(b, &resp); err != nil {
		return Status{}, fmt.Errorf("decode status: %w", err)
	}
	switch resp.OK {
	case 1:
		return statusFrom(resp.Data), nil
	case -100:
		return Status{}, ErrWalled
	default:
		return Status{}, ErrNotFound
	}
}

// Comments returns up to limit comments under the post with the given id.
// Weibo paginates via max_id; this fetches a single page (up to ~20 items).
// Pass limit=0 for no cap beyond the single page.
func (c *Client) Comments(ctx context.Context, id string, limit int) ([]Comment, error) {
	rawURL := c.cfg.MobileBaseURL + "/comments/hotflow?id=" + id + "&mid=" + id + "&max_id=0"
	b, err := c.getMobile(ctx, rawURL)
	if err != nil {
		return nil, err
	}
	var resp commentsResponse
	if err := json.Unmarshal(b, &resp); err != nil {
		return nil, fmt.Errorf("decode comments: %w", err)
	}
	if resp.OK != 1 {
		return nil, ErrWalled
	}
	wires := resp.Data.Data
	if limit > 0 && limit < len(wires) {
		wires = wires[:limit]
	}
	out := make([]Comment, 0, len(wires))
	for _, w := range wires {
		out = append(out, commentFrom(w))
	}
	return out, nil
}

// UserByID returns the public profile for a Weibo user by numeric uid.
// Requires a session cookie (SUB=xxx) — exits ErrWalled without one.
func (c *Client) UserByID(ctx context.Context, uid string) (User, error) {
	if c.cfg.Cookie == "" {
		return User{}, ErrWalled
	}
	rawURL := c.cfg.MobileBaseURL + "/api/container/getIndex?containerid=100505" + uid
	b, err := c.getMobile(ctx, rawURL)
	if err != nil {
		return User{}, err
	}
	var resp userProfileResponse
	if err := json.Unmarshal(b, &resp); err != nil {
		return User{}, fmt.Errorf("decode user: %w", err)
	}
	switch resp.OK {
	case 1:
		u := resp.Data.UserInfo
		if u.ID == 0 {
			return User{}, ErrNotFound
		}
		return userFrom(u), nil
	case -100:
		return User{}, ErrWalled
	default:
		return User{}, ErrNotFound
	}
}

// PostsByUID returns posts from a user's timeline (one page of ~10).
// Requires a session cookie (SUB=xxx) — exits ErrWalled without one.
func (c *Client) PostsByUID(ctx context.Context, uid string, page, limit int) ([]Post, error) {
	if c.cfg.Cookie == "" {
		return nil, ErrWalled
	}
	rawURL := fmt.Sprintf("%s/api/container/getIndex?containerid=107603%s&page=%d",
		c.cfg.MobileBaseURL, uid, page)
	b, err := c.getMobile(ctx, rawURL)
	if err != nil {
		return nil, err
	}
	var resp userTimelineResponse
	if err := json.Unmarshal(b, &resp); err != nil {
		return nil, fmt.Errorf("decode posts: %w", err)
	}
	if resp.OK != 1 {
		return nil, ErrWalled
	}
	var out []Post
	for _, card := range resp.Data.Cards {
		if card.CardType != 9 || card.MBlog.ID == "" {
			continue
		}
		out = append(out, postFrom(card.MBlog))
		if limit > 0 && len(out) >= limit {
			break
		}
	}
	return out, nil
}

// ExtendStatus fetches the full longTextContent for a post with isLongText=true.
// Works anonymously without a session cookie.
func (c *Client) ExtendStatus(ctx context.Context, id string) (string, error) {
	rawURL := c.cfg.MobileBaseURL + "/statuses/extend?id=" + id
	b, err := c.getMobile(ctx, rawURL)
	if err != nil {
		return "", err
	}
	var resp extendResponse
	if err := json.Unmarshal(b, &resp); err != nil {
		return "", fmt.Errorf("decode extend: %w", err)
	}
	if resp.OK != 1 {
		return "", ErrNotFound
	}
	return resp.Data.LongTextContent, nil
}

// Suggest returns search autocomplete suggestions for the given query.
func (c *Client) Suggest(ctx context.Context, query string, limit int) ([]Suggestion, error) {
	rawURL := c.cfg.BaseURL + "/ajax/side/search?q=" + query
	b, err := c.getDesktop(ctx, rawURL)
	if err != nil {
		return nil, err
	}
	var resp suggestResponse
	if err := json.Unmarshal(b, &resp); err != nil {
		return nil, fmt.Errorf("decode suggest: %w", err)
	}
	if resp.OK != 1 {
		return nil, ErrWalled
	}
	wires := resp.Data.HotQuery
	if limit > 0 && limit < len(wires) {
		wires = wires[:limit]
	}
	out := make([]Suggestion, 0, len(wires))
	for _, w := range wires {
		out = append(out, suggestionFrom(w))
	}
	return out, nil
}
