//go:build embed

package web

import "net/http"

// Settings-injected HTML varies with settings and the request's CSP nonce.
func applyHTMLNoCacheHeaders(header http.Header) {
	header.Set("Cache-Control", "no-cache, no-store, must-revalidate")
	header.Set("Pragma", "no-cache")
	header.Set("Expires", "0")
	header.Set("CDN-Cache-Control", "no-store")
	header.Set("Surrogate-Control", "no-store")
}
