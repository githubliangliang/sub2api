package service

import (
	"bytes"
	"fmt"

	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

// HTTP Responses rejects this nested metadata field when replaying messages.
// Strip it from every input item in one pass, preserving other metadata and raw
// JSON values so long histories do not require repeated upstream retries.
func stripOpenAIResponsesInputContentItemKinds(body []byte) ([]byte, error) {
	if !bytes.Contains(body, []byte("content_item_kinds")) && !bytes.Contains(body, []byte(`\u`)) {
		return body, nil
	}
	input := gjson.GetBytes(body, "input")
	if !input.IsArray() {
		return body, nil
	}

	const metadataPath = "internal_chat_message_metadata_passthrough.content_item_kinds"
	rebuilt := make([]byte, 0, len(input.Raw))
	rebuilt = append(rebuilt, '[')
	changed := false
	var stripErr error
	input.ForEach(func(_, item gjson.Result) bool {
		if len(rebuilt) > 1 {
			rebuilt = append(rebuilt, ',')
		}
		if item.IsObject() && item.Get(metadataPath).Exists() {
			var stripped []byte
			stripped, stripErr = sjson.DeleteBytes([]byte(item.Raw), metadataPath)
			if stripErr != nil {
				return false
			}
			rebuilt = append(rebuilt, stripped...)
			changed = true
		} else {
			rebuilt = append(rebuilt, item.Raw...)
		}
		return true
	})
	if stripErr != nil {
		return body, fmt.Errorf("delete OpenAI input content_item_kinds: %w", stripErr)
	}
	if !changed {
		return body, nil
	}
	rebuilt = append(rebuilt, ']')
	stripped, err := sjson.SetRawBytes(body, "input", rebuilt)
	if err != nil {
		return body, fmt.Errorf("replace OpenAI input after metadata deletion: %w", err)
	}
	return stripped, nil
}
