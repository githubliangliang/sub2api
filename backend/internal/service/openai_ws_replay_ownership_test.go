package service

import (
	"encoding/json"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestOpenAIWSReplayConcurrentMetadataAndLineageOwnership(t *testing.T) {
	payload := []byte(`{"input":[{"type":"reasoning","id":"rs-1","encrypted_content":"stale"},{"type":"message","role":"user","content":"keep","internal_chat_message_metadata_passthrough":{"content_item_kinds":["text"],"other":true}}]}`)
	original := string(payload)
	store := NewOpenAIWSStateStore(nil)
	var wg sync.WaitGroup
	for worker := range 16 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range 20 {
				session := fmt.Sprintf("session-%d", worker%4)
				store.MarkSessionInvalidEncryptedContent(1, session, []string{openAIEncryptedContentDigest("stale")}, time.Minute)
				items, exists, err := openAIWSExtractNormalizedInputSequence(payload)
				if err != nil || !exists || len(items) != 2 {
					t.Errorf("extract: %v, exists=%v, items=%d", err, exists, len(items))
					return
				}
				invalid := store.GetSessionInvalidEncryptedContentDigests(1, session)
				clean, count := stripOpenAIInvalidEncryptedContentFromReplayItems(items, invalid)
				if count != 1 || len(clean) != 2 || string(items[0]) != `{"type":"reasoning","id":"rs-1","encrypted_content":"stale"}` {
					t.Error("lineage stripping mutated shared replay or removed valid input")
				}
				body, err := setOpenAIWSPayloadInputSequence(payload, clean, true)
				if err != nil {
					t.Error(err)
					return
				}
				body, err = stripOpenAIResponsesInputContentItemKinds(body)
				if err != nil || !json.Valid(body) || string(payload) != original {
					t.Errorf("metadata rewrite violated ownership: %v", err)
				}
			}
		}()
	}
	wg.Wait()
	require.Equal(t, original, string(payload))
}

func TestOpenAIWSLineageCleanupExpiresSessions(t *testing.T) {
	store := NewOpenAIWSStateStore(nil).(*defaultOpenAIWSStateStore)
	store.MarkSessionInvalidEncryptedContent(1, "expired", []string{"digest"}, time.Minute)
	store.sessionInvalidEncryptedMu.Lock()
	binding := store.sessionInvalidEncrypted["1:expired"]
	binding.expiresAt = time.Now().Add(-time.Second)
	store.sessionInvalidEncrypted["1:expired"] = binding
	store.sessionInvalidEncryptedMu.Unlock()
	store.lastCleanupUnixNano.Store(time.Now().Add(-2 * time.Minute).UnixNano())
	require.Empty(t, store.GetSessionInvalidEncryptedContentDigests(1, "expired"))
	require.False(t, store.HasAnySessionInvalidEncryptedContent())
	store.MarkSessionInvalidEncryptedContent(1, "expired", []string{"fresh"}, time.Minute)
	require.Equal(t, map[string]struct{}{"fresh": {}}, store.GetSessionInvalidEncryptedContentDigests(1, "expired"))
}
