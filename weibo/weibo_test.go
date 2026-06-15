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

// ─── user ────────────────────────────────────────────────────────────────────

const mockUserResponse = `{
  "ok": 1,
  "data": {
    "userInfo": {
      "id": 2656274875,
      "screen_name": "央视新闻",
      "description": "中央电视台新闻中心官方微博",
      "verified": true,
      "verified_reason": "中央电视台新闻中心官方账号",
      "gender": "m",
      "location": "北京 朝阳区",
      "followers_count": "1.8亿",
      "follow_count": 244,
      "statuses_count": 688481,
      "avatar_hd": "https://tvax3.sinaimg.cn/crop.0.0.300.300.1024/9e9cb5adly8h9k7xd99ddj208c08cwfe.jpg"
    }
  }
}`

const mockUserWalledResponse = `{"ok": -100, "data": {}}`

const mockPostsResponse = `{
  "ok": 1,
  "data": {
    "cards": [
      {
        "card_type": 9,
        "mblog": {
          "id": "5309997458393240",
          "bid": "R4c4VzdsQ",
          "text": "测试帖子",
          "created_at": "Mon Jun 15 09:05:12 +0800 2026",
          "source": "微博视频号",
          "reposts_count": 10,
          "comments_count": 5,
          "attitudes_count": 50,
          "isLongText": false,
          "pic_num": 0,
          "user": {"id": 2656274875, "screen_name": "央视新闻"}
        }
      },
      {
        "card_type": 11,
        "mblog": {"id": ""}
      }
    ]
  }
}`

const mockExtendResponse = `{
  "ok": 1,
  "data": {
    "longTextContent": "这是一篇超长微博的完整正文，包含了原来被省略的内容。"
  }
}`

func newTestClientWithCookie(ts *httptest.Server, cookie string) *weibo.Client {
	cfg := weibo.DefaultConfig()
	cfg.BaseURL = ts.URL
	cfg.MobileBaseURL = ts.URL
	cfg.Rate = 0
	cfg.Cookie = cookie
	return weibo.NewClient(cfg)
}

func TestUserWalledWithoutCookie(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(mockUserResponse))
	}))
	defer srv.Close()
	_, err := newTestClient(srv).UserByID(context.Background(), "2656274875")
	if err == nil {
		t.Fatal("expected ErrWalled without a cookie")
	}
}

func TestUserSendsCookie(t *testing.T) {
	var gotCookie string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotCookie = r.Header.Get("Cookie")
		_, _ = w.Write([]byte(mockUserResponse))
	}))
	defer srv.Close()
	_, err := newTestClientWithCookie(srv, "SUB=abc123").UserByID(context.Background(), "2656274875")
	if err != nil {
		t.Fatal(err)
	}
	if gotCookie != "SUB=abc123" {
		t.Errorf("server received Cookie %q, want SUB=abc123", gotCookie)
	}
}

func TestUserParses(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(mockUserResponse))
	}))
	defer srv.Close()

	u, err := newTestClientWithCookie(srv, "SUB=abc").UserByID(context.Background(), "2656274875")
	if err != nil {
		t.Fatal(err)
	}
	if u.ID != 2656274875 {
		t.Errorf("id = %d, want 2656274875", u.ID)
	}
	if u.ScreenName != "央视新闻" {
		t.Errorf("screen_name = %q", u.ScreenName)
	}
	if !u.Verified {
		t.Error("verified should be true")
	}
	if u.VerifiedFor != "中央电视台新闻中心官方账号" {
		t.Errorf("verified_for = %q", u.VerifiedFor)
	}
	if u.Followers != "1.8亿" {
		t.Errorf("followers = %q, want 1.8亿", u.Followers)
	}
	if u.Following != 244 {
		t.Errorf("following = %d, want 244", u.Following)
	}
	if u.Posts != 688481 {
		t.Errorf("posts = %d, want 688481", u.Posts)
	}
	if u.URL == "" {
		t.Error("url is empty")
	}
}

func TestUserWalledResponse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(mockUserWalledResponse))
	}))
	defer srv.Close()
	_, err := newTestClientWithCookie(srv, "SUB=expired").UserByID(context.Background(), "2656274875")
	if err == nil {
		t.Fatal("expected error for ok:-100")
	}
}

// ─── posts ───────────────────────────────────────────────────────────────────

func TestPostsWalledWithoutCookie(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(mockPostsResponse))
	}))
	defer srv.Close()
	_, err := newTestClient(srv).PostsByUID(context.Background(), "2656274875", 1, 0)
	if err == nil {
		t.Fatal("expected ErrWalled without a cookie")
	}
}

func TestPostsFiltersCardType(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(mockPostsResponse))
	}))
	defer srv.Close()

	posts, err := newTestClientWithCookie(srv, "SUB=abc").PostsByUID(context.Background(), "2656274875", 1, 0)
	if err != nil {
		t.Fatal(err)
	}
	// card_type=11 and mblog.id="" should be filtered out; only 1 card_type=9
	if len(posts) != 1 {
		t.Fatalf("got %d posts, want 1 (card_type=11 should be filtered)", len(posts))
	}
	p := posts[0]
	if p.ID != "5309997458393240" {
		t.Errorf("id = %q", p.ID)
	}
	if p.Text != "测试帖子" {
		t.Errorf("text = %q, want 测试帖子", p.Text)
	}
}

func TestPostsLimitRespected(t *testing.T) {
	body := `{
	  "ok": 1,
	  "data": {
	    "cards": [
	      {"card_type": 9, "mblog": {"id": "1", "bid": "a", "text": "A", "created_at": "Mon Jun 15 09:00:00 +0800 2026", "source": "s", "user": {"id": 1, "screen_name": "u"}}},
	      {"card_type": 9, "mblog": {"id": "2", "bid": "b", "text": "B", "created_at": "Mon Jun 15 09:01:00 +0800 2026", "source": "s", "user": {"id": 1, "screen_name": "u"}}},
	      {"card_type": 9, "mblog": {"id": "3", "bid": "c", "text": "C", "created_at": "Mon Jun 15 09:02:00 +0800 2026", "source": "s", "user": {"id": 1, "screen_name": "u"}}}
	    ]
	  }
	}`
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(body))
	}))
	defer srv.Close()

	posts, err := newTestClientWithCookie(srv, "SUB=abc").PostsByUID(context.Background(), "1", 1, 2)
	if err != nil {
		t.Fatal(err)
	}
	if len(posts) != 2 {
		t.Fatalf("got %d posts with limit=2, want 2", len(posts))
	}
}

// ─── extend ──────────────────────────────────────────────────────────────────

func TestExtendParses(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(mockExtendResponse))
	}))
	defer srv.Close()

	text, err := newTestClient(srv).ExtendStatus(context.Background(), "5309997458393240")
	if err != nil {
		t.Fatal(err)
	}
	if text != "这是一篇超长微博的完整正文，包含了原来被省略的内容。" {
		t.Errorf("longTextContent = %q", text)
	}
}

// ─── parseUID ────────────────────────────────────────────────────────────────

func TestParseUID(t *testing.T) {
	cases := []struct {
		input string
		want  string
		err   bool
	}{
		{"2656274875", "2656274875", false},
		{"https://weibo.com/u/2656274875", "2656274875", false},
		{"https://m.weibo.cn/u/2656274875", "2656274875", false},
		{"https://m.weibo.cn/profile/2656274875", "2656274875", false},
		{"", "", true},
		{"notanumber", "", true},
		{"https://weibo.com/cctv_news", "", true},
	}
	for _, tc := range cases {
		got, err := weibo.ParseUID(tc.input)
		if tc.err {
			if err == nil {
				t.Errorf("ParseUID(%q) expected error", tc.input)
			}
			continue
		}
		if err != nil {
			t.Errorf("ParseUID(%q) error: %v", tc.input, err)
			continue
		}
		if got != tc.want {
			t.Errorf("ParseUID(%q) = %q, want %q", tc.input, got, tc.want)
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
