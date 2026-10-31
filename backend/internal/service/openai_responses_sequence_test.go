//go:build unit

package service

import (
	"context"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestResponsesSequenceTracksActualNativeStreamWrites(t *testing.T) {
	for _, passthrough := range []bool{false, true} {
		t.Run(fmt.Sprintf("passthrough=%t", passthrough), func(t *testing.T) {
			payload := "data: {\"type\":\"response.created\",\"sequence_number\":10,\"response\":{\"id\":\"resp_sequence\",\"model\":\"gpt-5.4\"}}\n\ndata: {\"type\":\"response.output_text.delta\",\"sequence_number\":11,\"delta\":\"ok\"}\n\ndata: {\"type\":\"response.completed\",\"sequence_number\":12,\"response\":{\"id\":\"resp_sequence\",\"status\":\"completed\",\"output\":[],\"usage\":{\"input_tokens\":3,\"output_tokens\":2}}}\n\ndata: [DONE]\n\n"
			response := &http.Response{StatusCode: http.StatusOK, Header: http.Header{"Content-Type": []string{"text/event-stream"}}, Body: io.NopCloser(strings.NewReader(payload))}
			recorder := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(recorder)
			c.Request = httptest.NewRequest(http.MethodPost, "/v1/responses", nil)
			svc := &OpenAIGatewayService{cfg: rawChatCompletionsTestConfig()}
			account := &Account{ID: 1, Platform: PlatformOpenAI, Type: AccountTypeAPIKey}
			var err error
			if passthrough {
				_, err = svc.handleStreamingResponsePassthrough(context.Background(), response, c, account, time.Now(), "gpt-5.4", "gpt-5.4")
			} else {
				_, err = svc.handleStreamingResponse(context.Background(), response, c, account, time.Now(), "gpt-5.4", "gpt-5.4")
			}
			require.NoError(t, err)
			require.Contains(t, recorder.Body.String(), `"sequence_number":10`)
			require.Contains(t, recorder.Body.String(), `"sequence_number":12`)
			require.EqualValues(t, 13, NextOpenAIResponsesSequence(c))
		})
	}
}

func TestResponsesSequenceIntegerParsingAndConcurrentAllocation(t *testing.T) {
	for _, tc := range []struct {
		input string
		want  int64
	}{
		{": heartbeat\n\n", -1},
		{"data: {\"sequence_number\":0}\n\n", 0},
		{`{"sequence_number":41}`, 41},
		{"data: {\"sequence_number\":8}\n\ndata: {\"sequence_number\":11}\n\n", 11},
		{`{"sequence_number":-1}`, -1},
		{`{"sequence_number":"12"}`, -1},
		{`{"sequence_number":1.5}`, -1},
		{`{"sequence_number":9223372036854775808}`, -1},
	} {
		require.Equal(t, tc.want, openAIResponsesSequenceFromSSE(tc.input), tc.input)
	}
	state := newResponsesSequenceState()
	values := make(chan int64, 64)
	var workers sync.WaitGroup
	for range 64 {
		workers.Add(1)
		go func() { defer workers.Done(); values <- state.next() }()
	}
	workers.Wait()
	close(values)
	seen := map[int64]bool{}
	for value := range values {
		require.False(t, seen[value])
		seen[value] = true
	}
	require.Len(t, seen, 64)
	state.observe(math.MaxInt64)
	require.EqualValues(t, math.MaxInt64, state.next())
}

func TestDeferredFailureSequenceDoesNotReuseAnAlreadyWrittenSlot(t *testing.T) {
	frame := buildOpenAIResponseFailedSSE("response", "gpt-5.4", 1, []byte(`{"sequence_number":3,"error":{"message":"held"}}`), "", 10)
	require.Contains(t, frame, `"sequence_number":10`)
	frame = buildOpenAIResponseFailedSSE("response", "gpt-5.4", 1, []byte(`{"sequence_number":15,"error":{"message":"held"}}`), "", 10)
	require.Contains(t, frame, `"sequence_number":15`)
}
