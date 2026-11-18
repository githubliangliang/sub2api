package service

import (
	"bytes"
	"strings"
)

// Escaped keys and event types must use the full parser. The remaining checks
// only skip decoding events that have no usage or mutable protocol metadata.
func sseEventNeedsJSONParse(eventName string, data []byte, useNoopDeltaKeepalive, needModelReplace bool) bool {
	if bytes.Contains(data, []byte(`\u`)) {
		return true
	}
	eventName = strings.ToLower(strings.TrimSpace(eventName))
	if eventName == "" || eventName == "message_start" || eventName == "message_delta" {
		return true
	}
	if useNoopDeltaKeepalive {
		switch eventName {
		case "content_block_start", "content_block_delta", "content_block_stop", "message_stop":
			return true
		}
	}
	if needModelReplace && bytes.Contains(data, []byte(`"message"`)) {
		return true
	}
	return bytes.Contains(data, []byte(`"usage"`))
}
