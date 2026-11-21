package securityaudit

import (
	"github.com/stretchr/testify/require"
	"testing"
	"unsafe"
)

func TestPrompt024ScalarOwnership(t *testing.T) {
	body := []byte(`{"input":"hello"}`)
	borrowed := unsafe.String(unsafe.SliceData(body), len(body))[:3]
	req := Request{Protocol: "openai_responses", Body: body, Model: borrowed, RequestID: borrowed, Username: borrowed, UserEmail: borrowed, APIKeyName: borrowed, GroupName: borrowed, Provider: borrowed, Endpoint: borrowed, Stage: borrowed}
	cloned := req.Clone()
	snapshot, err := ExtractPromptSnapshot(req)
	require.NoError(t, err)
	for i := range body {
		body[i] = 'x'
	}
	for _, value := range []string{cloned.Model, cloned.RequestID, cloned.Username, cloned.UserEmail, cloned.APIKeyName, cloned.GroupName, cloned.Provider, cloned.Endpoint, cloned.Stage, snapshot.Model, snapshot.RequestID, snapshot.UsernameSnapshot, snapshot.UserEmailSnapshot, snapshot.APIKeyNameSnapshot, snapshot.GroupName, snapshot.Provider, snapshot.Endpoint, snapshot.Stage} {
		require.Equal(t, `{"i`, value)
	}
	require.Equal(t, `{"input":"hello"}`, string(cloned.Body))
}
