package weibo_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/tamnd/weibo-cli/weibo"
)

const mockHotSearchResponse = `{
  "ok": 1,
  "data": {
    "realtime": [
      {
        "rank": 0,
        "word": "巴西醒醒 这是世界杯",
        "note": "巴西醒醒 这是世界杯",
        "num": 11551545,
        "category": "",
        "label_name": "爆"
      },
      {
        "rank": 1,
        "word": "中国女篮",
        "note": "中国女篮",
        "num": 8234567,
        "category": "体育",
        "label_name": "热"
      },
      {
        "rank": 2,
        "word": "新能源车型发布",
        "note": "新能源车型发布",
        "num": 5100000,
        "category": "科技",
        "label_name": "新"
      }
    ]
  }
}`

func newTestClient(ts *httptest.Server) *weibo.Client {
	cfg := weibo.DefaultConfig()
	cfg.BaseURL = ts.URL
	cfg.Rate = 0
	return weibo.NewClient(cfg)
}

func TestHotSearchSendsUserAgent(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("User-Agent") == "" {
			t.Error("request carried no User-Agent")
		}
		_, _ = w.Write([]byte(mockHotSearchResponse))
	}))
	defer srv.Close()

	c := newTestClient(srv)
	_, err := c.HotSearch(context.Background(), 0)
	if err != nil {
		t.Fatal(err)
	}
}

func TestHotSearchSendsReferer(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Referer") == "" {
			t.Error("request carried no Referer header")
		}
		_, _ = w.Write([]byte(mockHotSearchResponse))
	}))
	defer srv.Close()

	c := newTestClient(srv)
	_, err := c.HotSearch(context.Background(), 0)
	if err != nil {
		t.Fatal(err)
	}
}

func TestHotSearchParsesResults(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(mockHotSearchResponse))
	}))
	defer srv.Close()

	c := newTestClient(srv)
	items, err := c.HotSearch(context.Background(), 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 3 {
		t.Fatalf("got %d items, want 3", len(items))
	}

	it := items[0]
	if it.Rank != 1 {
		t.Errorf("rank = %d, want 1 (1-based)", it.Rank)
	}
	if it.Word != "巴西醒醒 这是世界杯" {
		t.Errorf("word = %q", it.Word)
	}
	if it.Heat != 11551545 {
		t.Errorf("heat = %d, want 11551545", it.Heat)
	}
	if it.Label != "爆" {
		t.Errorf("label = %q, want 爆", it.Label)
	}
	if it.URL == "" {
		t.Error("url is empty")
	}

	// Second item: category present
	it2 := items[1]
	if it2.Rank != 2 {
		t.Errorf("rank = %d, want 2", it2.Rank)
	}
	if it2.Category != "体育" {
		t.Errorf("category = %q, want 体育", it2.Category)
	}
}

func TestHotSearchLimit(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(mockHotSearchResponse))
	}))
	defer srv.Close()

	c := newTestClient(srv)
	items, err := c.HotSearch(context.Background(), 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 {
		t.Fatalf("got %d items with limit=1, want 1", len(items))
	}
}

func TestHotSearchRetriesOn503(t *testing.T) {
	var hits int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		if hits < 3 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		_, _ = w.Write([]byte(mockHotSearchResponse))
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

func TestHotSearchHTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer srv.Close()

	c := newTestClient(srv)
	_, err := c.HotSearch(context.Background(), 0)
	if err == nil {
		t.Fatal("expected error on 403, got nil")
	}
}
