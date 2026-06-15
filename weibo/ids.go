package weibo

import (
	"fmt"
	"net/url"
	"strings"
)

// ParseStatusID extracts the numeric post id from a full URL, a mobile URL, or
// a bare id string. Accepted forms:
//
//	5309997458393240
//	https://m.weibo.cn/detail/5309997458393240
//	https://m.weibo.cn/status/5309997458393240
//	https://weibo.com/2656274875/R4c4VzdsQ  (bid form — rejected, needs fetch)
// ParseStatusID is the exported form used by tests and ops.go.
func ParseStatusID(input string) (string, error) { return parseStatusID(input) }

func parseStatusID(input string) (string, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return "", fmt.Errorf("post id is required")
	}
	if !strings.Contains(input, "/") {
		// bare id
		if !isNumeric(input) {
			return "", fmt.Errorf("post id must be a numeric id or a weibo post URL; got %q", input)
		}
		return input, nil
	}
	u, err := url.Parse(input)
	if err != nil {
		return "", fmt.Errorf("invalid URL %q: %w", input, err)
	}
	// /detail/ID or /status/ID
	seg := strings.Split(strings.TrimPrefix(u.Path, "/"), "/")
	for i, s := range seg {
		if (s == "detail" || s == "status") && i+1 < len(seg) && isNumeric(seg[i+1]) {
			return seg[i+1], nil
		}
	}
	// last path segment if numeric
	last := seg[len(seg)-1]
	if isNumeric(last) {
		return last, nil
	}
	return "", fmt.Errorf("cannot extract a numeric post id from %q", input)
}

func isNumeric(s string) bool {
	if s == "" {
		return false
	}
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}
