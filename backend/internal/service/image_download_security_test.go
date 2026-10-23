package service

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestImageDownloadRejectsPrivateNetworkBeforeRequest(t *testing.T) {
	var hits atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		hits.Add(1)
		w.Header().Set("Content-Type", "image/png")
		_, _ = w.Write(pngBytes)
	}))
	defer server.Close()
	uploader := NewImageResultUploader(nil, "", 0, nil)
	_, _, err := uploader.download(context.Background(), server.URL)
	require.Error(t, err)
	require.Zero(t, hits.Load(), "the rejected destination must not be contacted")
}

type imageDownloadTrackedBody struct {
	io.Reader
	closed bool
}

func (b *imageDownloadTrackedBody) Close() error { b.closed = true; return nil }

func TestImageDownloadContentAndSizeValidationClosesBody(t *testing.T) {
	for _, tc := range []struct {
		name, contentType string
		data              []byte
		limit             int64
		status            int
		wantError         bool
	}{
		{"json", "application/json", []byte(`{"private":"data"}`), 0, http.StatusOK, true},
		{"html", "text/html", []byte("<html>private</html>"), 0, http.StatusOK, true},
		{"image", "application/octet-stream", pngBytes, 0, http.StatusOK, false},
		{"oversized", "image/png", pngBytes, 2, http.StatusOK, true},
		{"upstream error", "image/png", pngBytes, 0, http.StatusForbidden, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			body := &imageDownloadTrackedBody{Reader: bytes.NewReader(tc.data)}
			client := &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
				return &http.Response{StatusCode: tc.status, Header: http.Header{"Content-Type": []string{tc.contentType}}, Body: body, Request: req}, nil
			})}
			uploader := NewImageResultUploader(nil, "", tc.limit, client)
			_, _, err := uploader.download(context.Background(), "https://images.example.test/test")
			if tc.wantError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			require.True(t, body.closed, "response body must be closed on success and failure")
		})
	}
}
