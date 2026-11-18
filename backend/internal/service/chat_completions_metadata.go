package service

import (
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

// Carry only a recognized metadata.user_id, preserving its original JSON token.
func withChatCompletionsMetadataUserID(chatBody, anthropicBody []byte) []byte {
	userID := gjson.GetBytes(chatBody, "metadata.user_id")
	if userID.Type != gjson.String {
		return anthropicBody
	}
	if parsed := ParseMetadataUserID(userID.String()); parsed == nil || parsed.SessionID == "" {
		return anthropicBody
	}
	out, err := sjson.SetRawBytes(anthropicBody, "metadata", []byte(`{"user_id":`+userID.Raw+`}`))
	if err != nil {
		return anthropicBody
	}
	return out
}
