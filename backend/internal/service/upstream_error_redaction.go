package service

import (
	"net/url"
	"regexp"
	"sort"
	"strings"
)

// sanitizeUpstreamErrorMessageForAccount is for client-visible error fields;
// internal diagnostics retain their existing upstream message and body.
func sanitizeUpstreamErrorMessageForAccount(message string, account *Account) string {
	message = sanitizeUpstreamErrorMessage(strings.TrimSpace(message))
	if message == "" || account == nil {
		return message
	}
	secrets := []string{
		strings.TrimSpace(account.GetCredential("api_key")),
		strings.TrimSpace(account.GetCredential("access_token")),
		strings.TrimSpace(account.GetCredential("refresh_token")),
	}
	// Remove longer credentials first when configured values share a prefix.
	sort.Slice(secrets, func(i, j int) bool { return len(secrets[i]) > len(secrets[j]) })
	for _, secret := range secrets {
		if secret != "" {
			message = strings.ReplaceAll(message, secret, "[redacted]")
		}
	}
	baseURL := strings.TrimSpace(account.GetCredential("base_url"))
	if baseURL == "" {
		return message
	}
	parsed, err := url.Parse(baseURL)
	if err != nil {
		return ""
	}
	host := parsed.Hostname()
	if host == "" {
		return message
	}
	return replaceCaseInsensitive(message, host, "[upstream]")
}

func replaceCaseInsensitive(input, target, replacement string) string {
	if input == "" || target == "" {
		return input
	}
	// Match the original bytes: case folding can change UTF-8 byte lengths.
	pattern, err := regexp.Compile("(?i)" + regexp.QuoteMeta(target))
	if err != nil {
		return "" // A malformed target must not return unredacted client text.
	}
	return pattern.ReplaceAllLiteralString(input, replacement)
}
