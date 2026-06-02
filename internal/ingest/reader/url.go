package reader

import (
	"net/url"
	"strings"
)

// NormalizeURL trims and validates raw, returning a cleaned absolute http(s)
// URL. The bool is false when raw is empty or not a valid http(s) URL — callers
// should then treat the URL as absent (empty string) rather than drop the
// record, since the URL is non-content provenance metadata.
func NormalizeURL(raw string) (string, bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", false
	}
	u, err := url.Parse(raw)
	if err != nil || u.Host == "" || (u.Scheme != "http" && u.Scheme != "https") {
		return "", false
	}
	return u.String(), true
}
