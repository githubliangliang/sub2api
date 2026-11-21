//go:build unit

package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func TestUpstream024GrokEditInputValidation(t *testing.T) {
	for _, oauth := range []bool{false, true} {
		for _, async := range []bool{false, true} {
			for _, format := range []string{"json", "multipart"} {
				for _, count := range []int{3, 4} {
					t.Run(fmt.Sprintf("oauth=%t/async=%t/%s/%d", oauth, async, format, count), func(t *testing.T) {
						h, slots, _, upstream := newGrokMediaSlotHandler(t, oauth, false)
						h.grokMediaEligibilityProber = grokMediaSlotProber(func(context.Context, int64) (bool, string, error) { return true, "eligible", nil })
						upstream.call = func(*http.Request, int64) (*http.Response, error) {
							return &http.Response{StatusCode: http.StatusBadRequest, Header: http.Header{"Content-Type": {"application/json"}}, Body: io.NopCloser(strings.NewReader(`{"code":"invalid_image","error":"upstream image rejection"}`))}, nil
						}
						body, contentType := upstream024GrokEditBody(t, format, count)
						c, rec := grokMediaSlotContext(context.Background(), true)
						path := "/v1/images/edits"
						if async {
							path += "/async"
						}
						c.Request = httptest.NewRequest(http.MethodPost, path, bytes.NewReader(body))
						c.Request.Header.Set("Content-Type", contentType)
						var response []byte
						if async {
							store := &asyncImageMemoryStore{tasks: make(map[string]*service.ImageTaskRecord)}
							tasks := service.NewImageTaskServiceWithUploader(store, nil, time.Hour, time.Minute)
							asyncHandler := NewAsyncImageHandler(tasks, h)
							asyncHandler.Submit(c)
							require.Equal(t, http.StatusAccepted, rec.Code, rec.Body.String())
							taskID := gjson.GetBytes(rec.Body.Bytes(), "task_id").String()
							owner := service.ImageTaskOwner{UserID: 10, APIKeyID: 20}
							require.Eventually(t, func() bool {
								task, err := tasks.Get(context.Background(), owner, taskID)
								return err == nil && task.Status != service.ImageTaskStatusProcessing
							}, time.Second, 10*time.Millisecond)
							poll, pollRec := grokMediaSlotContext(context.Background(), false)
							poll.Params[0].Key, poll.Params[0].Value = "task_id", taskID
							asyncHandler.Get(poll)
							response = pollRec.Body.Bytes()
							require.Equal(t, int64(http.StatusBadRequest), gjson.GetBytes(response, "http_status").Int())
						} else {
							h.GrokImages(c)
							require.Equal(t, http.StatusBadRequest, rec.Code, rec.Body.String())
							response = rec.Body.Bytes()
						}
						require.Equal(t, "invalid_request_error", gjson.GetBytes(response, "error.type").String())
						if count > 3 {
							require.Contains(t, gjson.GetBytes(response, "error.message").String(), "a maximum of 3")
							require.Zero(t, upstream.calls)
						} else {
							require.Equal(t, "upstream image rejection", gjson.GetBytes(response, "error.message").String())
							require.Equal(t, 1, upstream.calls)
						}
						slots.assertReleased(t)
					})
				}
			}
		}
	}
}

func upstream024GrokEditBody(t *testing.T, format string, count int) ([]byte, string) {
	t.Helper()
	if format == "json" {
		images := make([]string, count)
		for i := range images {
			images[i] = fmt.Sprintf("https://example.test/%d.png", i)
		}
		body, err := json.Marshal(map[string]any{"model": "grok-imagine-image", "prompt": "test", "images": images})
		require.NoError(t, err)
		return body, "application/json"
	}
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	require.NoError(t, writer.WriteField("model", "grok-imagine-image"))
	require.NoError(t, writer.WriteField("prompt", "test"))
	for i := 0; i < count; i++ {
		require.NoError(t, writer.WriteField("image", fmt.Sprintf("https://example.test/%d.png", i)))
	}
	require.NoError(t, writer.Close())
	return body.Bytes(), writer.FormDataContentType()
}
