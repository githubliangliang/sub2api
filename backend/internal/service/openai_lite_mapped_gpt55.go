package service

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

// The account mapping changes the model, not the upstream capability header.
// Apply at the final HTTP request boundary, after account/model resolution.
// Never mutate the ingress headers/body: failover may select a native account.
func applyMappedGPT55LiteCompatibility(req *http.Request, account *Account, body []byte) error {
	if req == nil || !isMappedGPT55LiteTarget(account, gjson.GetBytes(body, "model").String()) {
		return nil
	}
	if !isOpenAIResponsesLiteHeader(req.Header.Get(responsesLiteHeader)) && !isOpenAIResponsesLiteWebSocketPayload(body) {
		return nil
	}
	// The Codex non-Lite endpoint accepts additional_tools, namespaces and
	// reasoning.context=all_turns. Preserve them and all history/tool results.
	if isOpenAIResponsesLiteWebSocketPayload(body) {
		var err error
		body, err = stripMappedGPT55LiteMetadata(body, account)
		if err != nil {
			return fmt.Errorf("remove mapped GPT-5.5 Lite metadata: %w", err)
		}
		req.Body = io.NopCloser(bytes.NewReader(body))
		req.ContentLength = int64(len(body))
		savedBody := append([]byte(nil), body...)
		req.GetBody = func() (io.ReadCloser, error) { return io.NopCloser(bytes.NewReader(savedBody)), nil }
	}
	req.Header.Del(responsesLiteHeader)
	return nil
}

func isMappedGPT55LiteTarget(account *Account, model string) bool {
	return account != nil && account.IsOpenAIOAuthLike() && strings.TrimSpace(model) == "gpt-5.5"
}

func stripMappedGPT55LiteMetadata(body []byte, account *Account) ([]byte, error) {
	if !isMappedGPT55LiteTarget(account, gjson.GetBytes(body, "model").String()) || !isOpenAIResponsesLiteWebSocketPayload(body) {
		return body, nil
	}
	return sjson.DeleteBytes(body, "client_metadata."+responsesLiteWSMetadataKey)
}
