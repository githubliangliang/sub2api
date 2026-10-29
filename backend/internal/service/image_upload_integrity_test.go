//go:build unit

package service

import (
	"bytes"
	"fmt"
	"mime"
	"mime/multipart"
	"net/http/httptest"
	"net/textproto"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func integrityMultipart(t *testing.T, uploads ...OpenAIImagesUpload) ([]byte, string) {
	t.Helper()
	var body bytes.Buffer
	w := multipart.NewWriter(&body)
	require.NoError(t, w.WriteField("model", "gpt-image-1"))
	require.NoError(t, w.WriteField("prompt", "keep complete images"))
	for _, upload := range uploads {
		h := textproto.MIMEHeader{}
		h.Set("Content-Disposition", mime.FormatMediaType("form-data", map[string]string{"name": upload.FieldName, "filename": upload.FileName}))
		h.Set("Content-Type", upload.ContentType)
		part, err := w.CreatePart(h)
		require.NoError(t, err)
		_, err = part.Write(upload.Data)
		require.NoError(t, err)
	}
	require.NoError(t, w.Close())
	return body.Bytes(), w.FormDataContentType()
}

func TestOpenAIImagesUploadPartBoundaryAndIntegrity(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, field := range []string{"image", "image[1]", "mask"} {
		for _, extra := range []int{0, 1} {
			t.Run(fmt.Sprintf("%s/extra=%d", field, extra), func(t *testing.T) {
				data := bytes.Repeat([]byte("x"), openAIImageMaxUploadPartSize+extra)
				data[len(data)-1] = 'z'
				first := OpenAIImagesUpload{FieldName: "image", FileName: "first.png", ContentType: "image/png", Data: []byte("first intact image")}
				last := OpenAIImagesUpload{FieldName: field, FileName: "last image.png", ContentType: "image/png", Data: data}
				body, contentType := integrityMultipart(t, first, last)
				c, _ := gin.CreateTestContext(httptest.NewRecorder())
				c.Request = httptest.NewRequest("POST", "/v1/images/edits", bytes.NewReader(body))
				c.Request.Header.Set("Content-Type", contentType)
				parsed, err := (&OpenAIGatewayService{}).ParseOpenAIImagesRequest(c, body)
				if extra > 0 {
					require.ErrorContains(t, err, "20 MB per-part limit")
					require.Nil(t, parsed, "an oversized part must not return a retryable partial request")
					return
				}
				require.NoError(t, err)
				require.Equal(t, first.Data, parsed.Uploads[0].Data)
				got := parsed.MaskUpload
				if field != "mask" {
					got = &parsed.Uploads[len(parsed.Uploads)-1]
				}
				require.NotNil(t, got)
				require.Equal(t, last.Data, got.Data)
				require.Equal(t, last.FileName, got.FileName)
				require.Equal(t, last.ContentType, got.ContentType)
			})
		}
	}
}

func TestGrokMediaUploadParsingRetainsFullBoundedBody(t *testing.T) {
	data := bytes.Repeat([]byte("a"), openAIImageMaxUploadPartSize+1)
	data[len(data)-1] = 'z'
	body, contentType := integrityMultipart(t,
		OpenAIImagesUpload{FieldName: "image[0]", FileName: "input.png", ContentType: "image/png", Data: data},
		OpenAIImagesUpload{FieldName: "mask", FileName: "mask.png", ContentType: "image/png", Data: []byte("mask intact")},
	)
	info := ParseGrokMediaRequest(contentType, body)
	require.Len(t, info.Uploads, 1)
	require.True(t, bytes.Equal(data, info.Uploads[0].Data), "review image length %d differs from accepted length %d", len(info.Uploads[0].Data), len(data))
	require.NotNil(t, info.MaskUpload)
	require.Equal(t, []byte("mask intact"), info.MaskUpload.Data)
	require.Equal(t, "input.png", info.Uploads[0].FileName)
}
