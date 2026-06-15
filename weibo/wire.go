package weibo

import (
	"net/url"
	"regexp"
	"strings"
	"time"
)

// topicURL returns the Weibo search URL for a hot-search topic word.
func topicURL(word string) string {
	return "https://s.weibo.com/weibo?q=%23" + url.QueryEscape(word) + "%23"
}

// statusURL returns the canonical mobile detail URL for a post.
func statusURL(id string) string {
	return "https://m.weibo.cn/detail/" + id
}

var tagRE = regexp.MustCompile(`<[^>]+>`)

// stripHTML removes HTML tags from Weibo text. It replaces <br/> variants with
// a space first so joined sentences stay readable on a single line.
func stripHTML(s string) string {
	s = strings.ReplaceAll(s, "<br />", " ")
	s = strings.ReplaceAll(s, "<br/>", " ")
	s = strings.ReplaceAll(s, "<br>", " ")
	s = tagRE.ReplaceAllString(s, "")
	s = strings.TrimSpace(s)
	// collapse consecutive whitespace that arises from removed tags
	return strings.Join(strings.Fields(s), " ")
}

// weiboTime parses the Weibo timestamp format "Mon Jun 15 09:05:12 +0800 2026"
// into "2026-06-15 09:05:12". Returns the original string on parse failure.
func weiboTime(s string) string {
	t, err := time.Parse("Mon Jan 02 15:04:05 -0700 2006", s)
	if err != nil {
		return s
	}
	return t.UTC().Format("2006-01-02 15:04:05")
}

func hotFrom(w wireHotItem, rank int) HotItem {
	return HotItem{
		Rank:     rank,
		Word:     w.Word,
		Scheme:   w.WordSch,
		Heat:     w.Num,
		Category: w.FlagDesc,
		Label:    w.LabelName,
		IsAd:     w.IsAd == 1,
		URL:      topicURL(w.Word),
	}
}

func statusFrom(w wireStatus) Status {
	return Status{
		ID:        w.ID,
		Bid:       w.Bid,
		Text:      stripHTML(w.Text),
		CreatedAt: weiboTime(w.CreatedAt),
		Source:    w.Source,
		Reposts:   w.Reposts,
		Comments:  w.Comments,
		Likes:     w.Likes,
		IsLong:    w.IsLong,
		PicNum:    w.PicNum,
		Username:  w.User.ScreenName,
		UserID:    w.User.ID,
		URL:       statusURL(w.ID),
	}
}

func commentFrom(w wireComment) Comment {
	return Comment{
		ID:        w.ID,
		Floor:     w.Floor,
		Text:      stripHTML(w.Text),
		CreatedAt: weiboTime(w.CreatedAt),
		Source:    w.Source,
		Likes:     w.Likes,
		Username:  w.User.ScreenName,
		UserID:    w.User.ID,
	}
}

func suggestionFrom(w wireSuggestion) Suggestion {
	return Suggestion{
		Word:  w.Suggestion,
		Count: w.Count,
		IsHot: w.TopFlag == 2,
	}
}
