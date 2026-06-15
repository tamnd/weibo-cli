package weibo

// HotItem is a single entry from the Weibo hot search board (微博热搜榜).
type HotItem struct {
	Rank     int    `json:"rank"     table:",right"`
	Word     string `json:"word"`
	Scheme   string `json:"scheme"`
	Heat     int    `json:"heat"     table:",right"`
	Category string `json:"category"`
	Label    string `json:"label"`
	IsAd     bool   `json:"is_ad"`
	URL      string `json:"url"      kit:"url" table:",truncate"`
}

// Status is a single Weibo post fetched by its numeric id.
type Status struct {
	ID        string `json:"id"`
	Bid       string `json:"bid"`
	Text      string `json:"text"      kit:"body"          table:",truncate"`
	CreatedAt string `json:"created_at"`
	Source    string `json:"source"`
	Reposts   int    `json:"reposts"   table:",right"`
	Comments  int    `json:"comments"  table:",right"`
	Likes     int    `json:"likes"     table:",right"`
	IsLong    bool   `json:"is_long"`
	PicNum    int    `json:"pic_num"   table:",right"`
	Username  string `json:"username"`
	UserID    int64  `json:"user_id"`
	URL       string `json:"url"       kit:"url"           table:",truncate"`
}

// Comment is one comment under a Weibo post.
type Comment struct {
	ID        string `json:"id"`
	Floor     int    `json:"floor"     table:",right"`
	Text      string `json:"text"      kit:"body"          table:",truncate"`
	CreatedAt string `json:"created_at"`
	Source    string `json:"source"`
	Likes     int    `json:"likes"     table:",right"`
	Username  string `json:"username"`
	UserID    int64  `json:"user_id"`
}

// Suggestion is one autocomplete entry from the Weibo search sidebar.
type Suggestion struct {
	Word  string `json:"word"`
	Count int    `json:"count"     table:",right"`
	IsHot bool   `json:"is_hot"`
}

// User is a public Weibo profile. Requires a session cookie.
type User struct {
	ID          int64  `json:"id"`
	ScreenName  string `json:"screen_name"`
	Description string `json:"description"    table:",truncate"`
	Verified    bool   `json:"verified"`
	VerifiedFor string `json:"verified_for"`
	Gender      string `json:"gender"`
	Location    string `json:"location"`
	Followers   string `json:"followers"`
	Following   int    `json:"following"      table:",right"`
	Posts       int    `json:"posts"          table:",right"`
	Avatar      string `json:"avatar"         table:",truncate"`
	URL         string `json:"url"            kit:"url" table:",truncate"`
}

// Post is one item from a user's post timeline. Requires a session cookie.
type Post struct {
	ID        string `json:"id"`
	Bid       string `json:"bid"`
	Text      string `json:"text"      kit:"body"          table:",truncate"`
	CreatedAt string `json:"created_at"`
	Source    string `json:"source"`
	Reposts   int    `json:"reposts"   table:",right"`
	Comments  int    `json:"comments"  table:",right"`
	Likes     int    `json:"likes"     table:",right"`
	IsLong    bool   `json:"is_long"`
	PicNum    int    `json:"pic_num"   table:",right"`
	URL       string `json:"url"       kit:"url"           table:",truncate"`
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
	RealPos   int    `json:"realpos"`
	Word      string `json:"word"`
	Num       int    `json:"num"`
	LabelName string `json:"label_name"`
	WordSch   string `json:"word_scheme"`
	FlagDesc  string `json:"flag_desc"`
	IsAd      int    `json:"is_ad"`
}

type statusResponse struct {
	OK   int        `json:"ok"`
	Data wireStatus `json:"data"`
}

type wireStatus struct {
	ID        string   `json:"id"`
	Bid       string   `json:"bid"`
	Text      string   `json:"text"`
	CreatedAt string   `json:"created_at"`
	Source    string   `json:"source"`
	Reposts   int      `json:"reposts_count"`
	Comments  int      `json:"comments_count"`
	Likes     int      `json:"attitudes_count"`
	IsLong    bool     `json:"isLongText"`
	PicNum    int      `json:"pic_num"`
	User      wireUser `json:"user"`
}

type wireUser struct {
	ID         int64  `json:"id"`
	ScreenName string `json:"screen_name"`
}

type commentsResponse struct {
	OK   int `json:"ok"`
	Data struct {
		Data  []wireComment `json:"data"`
		MaxID int64         `json:"max_id"`
		Total int           `json:"total_number"`
	} `json:"data"`
}

type wireComment struct {
	ID        string   `json:"id"`
	Floor     int      `json:"floor_number"`
	Text      string   `json:"text"`
	CreatedAt string   `json:"created_at"`
	Source    string   `json:"source"`
	Likes     int      `json:"like_count"`
	User      wireUser `json:"user"`
}

type suggestResponse struct {
	OK   int `json:"ok"`
	Data struct {
		HotQuery []wireSuggestion `json:"hotquery"`
	} `json:"data"`
}

type wireSuggestion struct {
	Suggestion string `json:"suggestion"`
	Count      int    `json:"count"`
	TopFlag    int    `json:"top_flag"`
}

type userProfileResponse struct {
	OK   int `json:"ok"`
	Data struct {
		UserInfo wireProfileUser `json:"userInfo"`
	} `json:"data"`
}

type wireProfileUser struct {
	ID             int64  `json:"id"`
	ScreenName     string `json:"screen_name"`
	Description    string `json:"description"`
	Verified       bool   `json:"verified"`
	VerifiedReason string `json:"verified_reason"`
	Gender         string `json:"gender"`
	Location       string `json:"location"`
	FollowersCount string `json:"followers_count"`
	FollowCount    int    `json:"follow_count"`
	StatusesCount  int    `json:"statuses_count"`
	AvatarHD       string `json:"avatar_hd"`
}

type userTimelineResponse struct {
	OK   int `json:"ok"`
	Data struct {
		Cards []wireTimelineCard `json:"cards"`
	} `json:"data"`
}

type wireTimelineCard struct {
	CardType int        `json:"card_type"`
	MBlog    wireStatus `json:"mblog"`
}

type extendResponse struct {
	OK   int `json:"ok"`
	Data struct {
		LongTextContent string `json:"longTextContent"`
	} `json:"data"`
}
