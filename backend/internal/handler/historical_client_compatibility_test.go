//go:build unit

package handler

import (
	"encoding/json"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"strings"
	"testing"
)

func TestResponsesFailureEventIncludesSequence(t *testing.T) {
	c, recorder := newGinContextForEndpoint(t, "/v1/responses")
	if !writeResponsesFailedSSE(c, "upstream_error", "fixture failure") {
		t.Fatal("Responses failure was not written")
	}
	for _, line := range strings.Split(recorder.Body.String(), "\n") {
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		data := []byte(strings.TrimPrefix(line, "data: "))
		var object map[string]json.RawMessage
		if err := json.Unmarshal(data, &object); err != nil {
			t.Fatal(err)
		}
		if _, exists := object["sequence_number"]; !exists {
			t.Errorf("actual handler failure omits sequence_number: %s", data)
		}
		return
	}
	t.Fatal("probe produced no SSE data")
}

func TestResponsesFailureSequenceContinuesCommittedEvents(t *testing.T) {
	c, recorder := newGinContextForEndpoint(t, "/v1/responses")
	service.ObserveOpenAIResponsesSequence(c, 42)
	service.ObserveOpenAIResponsesSequence(c, 7)
	if !writeResponsesFailedSSE(c, "upstream_error", "fixture failure") {
		t.Fatal("failure not written")
	}
	for _, line := range strings.Split(recorder.Body.String(), "\n") {
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		var event struct {
			Sequence int64 `json:"sequence_number"`
		}
		if err := json.Unmarshal([]byte(strings.TrimPrefix(line, "data: ")), &event); err != nil {
			t.Fatal(err)
		}
		if event.Sequence != 43 {
			t.Fatalf("failure sequence=%d, want 43", event.Sequence)
		}
		return
	}
	t.Fatal("failure event missing")
}
