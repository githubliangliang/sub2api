package service

import (
	"net/url"
	"strings"
)

const geminiOfficialAPIHost = "generativelanguage.googleapis.com"

// Used only to classify missing-cache errors in the existing generation path.
const GeminiCachedContentNotFoundResponse = `{"error":{"code":403,"message":"CachedContent not found (or permission denied)","status":"PERMISSION_DENIED"}}`

func isOfficialGeminiAPIBaseURL(raw string) bool {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return true
	}
	u, err := url.Parse(raw)
	if err != nil {
		return false
	}
	path := strings.TrimRight(u.Path, "/")
	return strings.EqualFold(u.Scheme, "https") &&
		strings.EqualFold(u.Hostname(), geminiOfficialAPIHost) &&
		u.Port() == "" && path == "" && u.RawQuery == ""
}
