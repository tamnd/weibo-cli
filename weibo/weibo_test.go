package weibo_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/tamnd/weibo-cli/weibo"
)

// Hot search fixture mirrors the real wire shape from weibo.com/ajax/side/hotSearch.
const mockHotResponse = `{
  "ok": 1,
  "data": {
    "realtime": [
      {
        "rank": 0, "realpos": 1,
        "word": "山姆被约谈", "num": 2453012,
        "label_name": "热", "word_scheme": "#山姆被约谈#",
        "flag_desc": "", "is_ad": 0
      },
      {
        "rank": 1, "realpos": 2,
        "word": "严浩翔贺峻霖", "num": 322924,
        "label_name": "新", "word_scheme": "#严浩翔贺峻霖#",
        "flag_desc": "综艺", "is_ad": 0
      },
      {
        "rank": 2, "realpos": 2,
        "word": "品牌广告", "num": 815180,
        "label_name": "商", "word_scheme": "#品牌广告#",
        "flag_desc": "", "is_ad": 1,
        "id": 342321
      }
    ]
  }
}`

// Status fixture mirrors m.weibo.cn/statuses/show response.
const mockStatusResponse = `{
  "ok": 1,
  "data": {
    "id": "5309997458393240",
    "bid": "R4c4VzdsQ",
    "text": "【时政微视频 | <a href=\"...\"><span class=\"surl-text\">#共产党员习近平#</span></a>】入党52年。<br />为人民服务。",
    "created_at": "Mon Jun 15 09:05:12 +0800 2026",
    "source": "微博视频号",
    "reposts_count": 334,
    "comments_count": 263,
    "attitudes_count": 1552,
    "isLongText": true,
    "pic_num": 0,
    "user": {
      "id": 2656274875,
      "screen_name": "央视新闻"
    }
  }
}`

// Comments fixture mirrors m.weibo.cn/comments/hotflow response.
const mockCommentsResponse = `{
  "ok": 1,
  "data": {
    "data": [
      {
        "id": "5309998831766407",
        "floor_number": 19,
        "text": "为人民服务",
        "created_at": "Mon Jun 15 09:10:39 +0800 2026",
        "source": "来自安徽",
        "like_count": 15,
        "user": { "id": 5274035875, "screen_name": "love梦妍" }
      }
    ],
    "max_id": 0,
    "total_number": 263
  }
}`

// Suggest fixture mirrors weibo.com/ajax/side/search response.
const mockSuggestResponse = `{
  "ok": 1,
  "data": {
    "hotquery": [
      { "suggestion": "山姆被约谈", "count": 2453012, "top_flag": 2 },
      { "suggestion": "山姆回应", "count": 224813, "top_flag": 0 }
    ],
    "history": [], "real_hot": [], "query_relates": [], "user": [], "users": []
  }
}`

func newTestClient(ts *httptest.Server) *weibo.Client {
	cfg := weibo.DefaultConfig()
	cfg.BaseURL = ts.URL
	cfg.MobileBaseURL = ts.URL
	cfg.Rate = 0
	return weibo.NewClient(cfg)
}

// ─── hot ─────────────────────────────────────────────────────────────────────

func TestHotSendsUserAgent(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("User-Agent") == "" {
			t.Error("request carried no User-Agent")
		}
		_, _ = w.Write([]byte(mockHotResponse))
	}))
	defer srv.Close()
	_, err := newTestClient(srv).HotSearch(context.Background(), 0)
	if err != nil {
		t.Fatal(err)
	}
}

func TestHotSendsReferer(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Referer") == "" {
			t.Error("request carried no Referer")
		}
		_, _ = w.Write([]byte(mockHotResponse))
	}))
	defer srv.Close()
	_, err := newTestClient(srv).HotSearch(context.Background(), 0)
	if err != nil {
		t.Fatal(err)
	}
}

func TestHotParses(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(mockHotResponse))
	}))
	defer srv.Close()

	items, err := newTestClient(srv).HotSearch(context.Background(), 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 3 {
		t.Fatalf("got %d items, want 3", len(items))
	}

	first := items[0]
	if first.Rank != 1 {
		t.Errorf("rank = %d, want 1", first.Rank)
	}
	if first.Word != "山姆被约谈" {
		t.Errorf("word = %q", first.Word)
	}
	if first.Heat != 2453012 {
		t.Errorf("heat = %d, want 2453012", first.Heat)
	}
	if first.Label != "热" {
		t.Errorf("label = %q, want 热", first.Label)
	}
	if first.IsAd {
		t.Error("first item should not be an ad")
	}
	if first.URL == "" {
		t.Error("url is empty")
	}

	second := items[1]
	if second.Category != "综艺" {
		t.Errorf("category = %q, want 综艺", second.Category)
	}

	ad := items[2]
	if !ad.IsAd {
		t.Error("third item should be an ad")
	}
}

func TestHotLimit(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(mockHotResponse))
	}))
	defer srv.Close()

	items, err := newTestClient(srv).HotSearch(context.Background(), 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 {
		t.Fatalf("got %d items with limit=1, want 1", len(items))
	}
}

func TestHotRetriesOn503(t *testing.T) {
	var hits int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		if hits < 3 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		_, _ = w.Write([]byte(mockHotResponse))
	}))
	defer srv.Close()

	cfg := weibo.DefaultConfig()
	cfg.BaseURL = srv.URL
	cfg.Rate = 0
	cfg.Retries = 5
	c := weibo.NewClient(cfg)

	start := time.Now()
	_, err := c.HotSearch(context.Background(), 0)
	if err != nil {
		t.Fatal(err)
	}
	if hits != 3 {
		t.Errorf("server saw %d hits, want 3", hits)
	}
	if time.Since(start) < 500*time.Millisecond {
		t.Error("retries did not back off")
	}
}

func TestHotHTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer srv.Close()
	_, err := newTestClient(srv).HotSearch(context.Background(), 0)
	if err == nil {
		t.Fatal("expected error on 403")
	}
}

// ─── status ──────────────────────────────────────────────────────────────────

func TestStatusParses(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(mockStatusResponse))
	}))
	defer srv.Close()

	s, err := newTestClient(srv).StatusByID(context.Background(), "5309997458393240")
	if err != nil {
		t.Fatal(err)
	}
	if s.ID != "5309997458393240" {
		t.Errorf("id = %q", s.ID)
	}
	if s.Bid != "R4c4VzdsQ" {
		t.Errorf("bid = %q", s.Bid)
	}
	if s.Likes != 1552 {
		t.Errorf("likes = %d, want 1552", s.Likes)
	}
	if s.Username != "央视新闻" {
		t.Errorf("username = %q", s.Username)
	}
	// HTML should be stripped from text
	if s.Text == "" {
		t.Error("text is empty after strip")
	}
	if contains(s.Text, "<a ") || contains(s.Text, "<span") {
		t.Errorf("text still contains HTML tags: %q", s.Text)
	}
	// Date should be reformatted
	if s.CreatedAt != "2026-06-15 01:05:12" {
		t.Errorf("created_at = %q, want 2026-06-15 01:05:12 (UTC)", s.CreatedAt)
	}
	if s.URL == "" {
		t.Error("url is empty")
	}
}

func TestStatusSendsMobileHeaders(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("MWeibo-Pwa") != "1" {
			t.Errorf("missing MWeibo-Pwa header")
		}
		if r.Header.Get("X-Requested-With") != "XMLHttpRequest" {
			t.Errorf("missing X-Requested-With header")
		}
		_, _ = w.Write([]byte(mockStatusResponse))
	}))
	defer srv.Close()
	_, err := newTestClient(srv).StatusByID(context.Background(), "5309997458393240")
	if err != nil {
		t.Fatal(err)
	}
}

// ─── comments ────────────────────────────────────────────────────────────────

func TestCommentsParses(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(mockCommentsResponse))
	}))
	defer srv.Close()

	comments, err := newTestClient(srv).Comments(context.Background(), "5309997458393240", 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(comments) != 1 {
		t.Fatalf("got %d comments, want 1", len(comments))
	}
	c := comments[0]
	if c.ID != "5309998831766407" {
		t.Errorf("id = %q", c.ID)
	}
	if c.Floor != 19 {
		t.Errorf("floor = %d, want 19", c.Floor)
	}
	if c.Text != "为人民服务" {
		t.Errorf("text = %q", c.Text)
	}
	if c.Likes != 15 {
		t.Errorf("likes = %d, want 15", c.Likes)
	}
	if c.Source != "来自安徽" {
		t.Errorf("source = %q", c.Source)
	}
	if c.Username != "love梦妍" {
		t.Errorf("username = %q", c.Username)
	}
}

// ─── suggest ─────────────────────────────────────────────────────────────────

func TestSuggestParses(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(mockSuggestResponse))
	}))
	defer srv.Close()

	sugs, err := newTestClient(srv).Suggest(context.Background(), "%E5%B1%B1%E5%A7%86", 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(sugs) != 2 {
		t.Fatalf("got %d suggestions, want 2", len(sugs))
	}
	if sugs[0].Word != "山姆被约谈" {
		t.Errorf("word = %q", sugs[0].Word)
	}
	if sugs[0].Count != 2453012 {
		t.Errorf("count = %d, want 2453012", sugs[0].Count)
	}
	if !sugs[0].IsHot {
		t.Error("first suggestion should be is_hot (top_flag==2)")
	}
	if sugs[1].IsHot {
		t.Error("second suggestion should not be is_hot")
	}
}

// ─── parseStatusID ───────────────────────────────────────────────────────────

func TestParseStatusID(t *testing.T) {
	cases := []struct {
		input string
		want  string
		err   bool
	}{
		{"5309997458393240", "5309997458393240", false},
		{"https://m.weibo.cn/detail/5309997458393240", "5309997458393240", false},
		{"https://m.weibo.cn/status/5309997458393240", "5309997458393240", false},
		{"", "", true},
		{"notanumber", "", true},
	}
	for _, tc := range cases {
		got, err := weibo.ParseStatusID(tc.input)
		if tc.err {
			if err == nil {
				t.Errorf("ParseStatusID(%q) expected error", tc.input)
			}
			continue
		}
		if err != nil {
			t.Errorf("ParseStatusID(%q) error: %v", tc.input, err)
			continue
		}
		if got != tc.want {
			t.Errorf("ParseStatusID(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(sub) == 0 || indexOf(s, sub) >= 0)
}

func indexOf(s, sub string) int {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
