package weibo

// HotItem is a single entry from the Weibo hot search list.
type HotItem struct {
	Rank     int    `json:"rank"`
	Word     string `json:"word"`
	Note     string `json:"note"`
	Heat     int    `json:"heat"`
	Category string `json:"category"`
	Label    string `json:"label"`
	URL      string `json:"url"`
}

// ─── wire types ──────────────────────────────────────────────────────────────

type hotSearchResponse struct {
	OK   int `json:"ok"`
	Data struct {
		Realtime []wireHotItem `json:"realtime"`
	} `json:"data"`
}

type wireHotItem struct {
	Rank      int    `json:"rank"`
	Word      string `json:"word"`
	Note      string `json:"note"`
	Num       int    `json:"num"`
	Category  string `json:"category"`
	LabelName string `json:"label_name"`
}

// wireToHotItem converts a wire item to HotItem.
// rank is the 1-based position derived from the slice index.
// The search URL wraps the topic word in #hashtag# notation.
func wireToHotItem(w wireHotItem, rank int, word string) HotItem {
	note := w.Note
	if note == "" {
		note = w.Word
	}
	return HotItem{
		Rank:     rank,
		Word:     w.Word,
		Note:     note,
		Heat:     w.Num,
		Category: w.Category,
		Label:    w.LabelName,
		URL:      topicURL(word),
	}
}
