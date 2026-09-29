package service

import (
	"github.com/stretchr/testify/require"
	"net/http"
	"testing"
)

func TestOpenAIStreamSemanticStatusHonorsErrorStatusAlias(t *testing.T) {
	tests := []struct {
		name         string
		body         string
		status       int
		wantFailover bool
	}{
		{
			name:         "response.error.status unauthorized",
			body:         `{"type":"response.failed","response":{"error":{"code":"server_error","status":401,"message":"upstream rejected the key"}}}`,
			status:       http.StatusUnauthorized,
			wantFailover: true,
		},
		{
			name:         "error.status rate limited",
			body:         `{"type":"error","error":{"code":"server_error","status":429,"message":"upstream is busy"}}`,
			status:       http.StatusTooManyRequests,
			wantFailover: true,
		},
		{
			name:         "error.status overloaded",
			body:         `{"type":"error","error":{"code":"server_error","status":529,"message":"upstream is overloaded"}}`,
			status:       529,
			wantFailover: true,
		},
		{
			name:   "response.error.status forbidden without account signal",
			body:   `{"type":"response.failed","response":{"error":{"code":"server_error","status":403,"message":"request was rejected"}}}`,
			status: http.StatusForbidden,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			payload := []byte(tt.body)
			message := extractOpenAISSEErrorMessage(payload)
			require.Equal(t, tt.status, openAIStreamFailureStatus(payload, message))
			require.Equal(t, tt.wantFailover, openAIStreamErrorEventShouldFailover(payload, message))
			require.Equal(t, tt.wantFailover, openAIStreamFailedEventShouldFailover(payload, message))
		})
	}
}
