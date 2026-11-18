//go:build unit

package handler

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/binary"
	"hash/crc32"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/repository"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func tier223BedrockFrame(payload string) []byte {
	var headers bytes.Buffer
	for _, pair := range [][2]string{{":event-type", "chunk"}, {":message-type", "event"}} {
		_ = headers.WriteByte(byte(len(pair[0])))
		_, _ = headers.WriteString(pair[0])
		_ = headers.WriteByte(7)
		_ = binary.Write(&headers, binary.BigEndian, uint16(len(pair[1])))
		_, _ = headers.WriteString(pair[1])
	}
	body := []byte(`{"bytes":"` + base64.StdEncoding.EncodeToString([]byte(payload)) + `"}`)
	frame := make([]byte, 12+headers.Len()+len(body)+4)
	binary.BigEndian.PutUint32(frame, uint32(len(frame)))
	binary.BigEndian.PutUint32(frame[4:], uint32(headers.Len()))
	binary.BigEndian.PutUint32(frame[8:], crc32.ChecksumIEEE(frame[:8]))
	copy(frame[12:], headers.Bytes())
	copy(frame[12+headers.Len():], body)
	binary.BigEndian.PutUint32(frame[len(frame)-4:], crc32.ChecksumIEEE(frame[:len(frame)-4]))
	return frame
}

func TestBedrockTruncatedHTTPUsageSettlesOnceSQLite(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, client := tier223SQLiteLedger(t, tier223SQLiteTemplate(t), service.PlatformAnthropic)
	var stream []byte
	for _, payload := range []string{
		`{"type":"message_start","message":{"id":"msg_br","type":"message","role":"assistant","model":"claude-sonnet-4-6","content":[],"usage":{"input_tokens":500,"cache_read_input_tokens":200,"cache_creation_input_tokens":300,"cache_creation":{"ephemeral_5m_input_tokens":100,"ephemeral_1h_input_tokens":100}}}}`,
		`{"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":"partial"}}`,
		`{"type":"message_delta","usage":{"output_tokens":10}}`,
	} {
		stream = append(stream, tier223BedrockFrame(payload)...)
	}
	stream = append(stream, 1, 2, 3, 4, 5) // A truncated binary prelude follows already observed usage.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.Copy(io.Discard, r.Body)
		if !strings.HasSuffix(r.URL.Path, "/invoke-with-response-stream") {
			t.Errorf("wrong Bedrock endpoint: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/vnd.amazon.eventstream")
		w.Header().Set("X-Amzn-Requestid", "bedrock-partial")
		_, _ = w.Write(stream)
	}))
	t.Cleanup(server.Close)
	target, err := url.Parse(server.URL)
	require.NoError(t, err)
	cfg := &config.Config{Default: config.DefaultConfig{RateMultiplier: 1}}
	pricing := tier223Pricing(t, cfg, false)
	repo := repository.NewAccountRepository(client, db, nil)
	svc := tier223ClaudeGateway(repo, db, cfg, tier223LocalHTTP{server.Client(), target}, nil, pricing)
	rate := 0.5
	account := &service.Account{ID: 3, Platform: service.PlatformAnthropic, Type: service.AccountTypeBedrock, Status: service.StatusActive, Schedulable: true, RateMultiplier: &rate, Credentials: map[string]any{"auth_mode": "apikey", "api_key": "bedrock-fixture-key", "aws_region": "us-east-1", "model_mapping": map[string]any{"claude-sonnet-4-6": "us.anthropic.claude-sonnet-4-6-v1"}}}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/messages", nil)
	result, err := tier223ForwardClaude(context.Background(), c, svc, account, "messages", true)
	require.Error(t, err)
	require.NotNil(t, result)
	var failover *service.UpstreamFailoverError
	require.NotErrorAs(t, err, &failover)
	require.Contains(t, w.Body.String(), "partial")
	require.Equal(t, 10, result.Usage.OutputTokens)
	groupID := int64(4)
	user := &service.User{ID: 1, Balance: 10}
	key := &service.APIKey{ID: 2, UserID: 1, User: user, Quota: 1, GroupID: &groupID, Group: &service.Group{ID: 4, Platform: service.PlatformAnthropic, Hydrated: true, RateMultiplier: 2}}
	input := &service.RecordUsageInput{Result: result, User: user, APIKey: key, Account: account, APIKeyService: grokAuditQuotaUpdater{}}
	require.NoError(t, svc.RecordUsage(context.Background(), input))
	require.NoError(t, svc.RecordUsage(context.Background(), input))
	tier223AssertLedger(t, db, 0.00283, 500, 300, 100, 100)
}
