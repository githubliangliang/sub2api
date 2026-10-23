package service

import (
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/tidwall/gjson"
)

// Codex may recover after a bare error. Delay non-retryable errors until a
// terminal response settles the outcome; keep existing pre-output failover.
// This state belongs to one stream and is accessed only by its consumer.
type openAIBareErrorStream struct {
	enabled       bool
	pending       []byte
	eventName     string
	suppressFrame bool
}

func (s *openAIBareErrorStream) skipLine(line string, outputStarted bool) bool {
	if !s.enabled {
		return false
	}
	if name, ok := extractOpenAISSEEventLine(line); ok {
		s.eventName = strings.TrimSpace(name)
		s.suppressFrame = s.eventName == "error"
		return s.suppressFrame
	}
	if line == "" {
		skip := s.suppressFrame
		s.suppressFrame = false
		s.eventName = ""
		return skip
	}
	data, ok := extractOpenAISSEDataLine(line)
	if !ok {
		return s.suppressFrame
	}
	if strings.TrimSpace(data) == "[DONE]" && len(s.pending) > 0 {
		s.suppressFrame = true
		return true
	}
	payload := []byte(data)
	kind := effectiveOpenAISSEEventType(payload, s.eventName)
	if kind == "error" {
		message := extractOpenAISSEErrorMessage(payload)
		retryable := openAIStreamErrorEventShouldFailover(payload, message) || isOpenAIUpstreamCapacityShedEvent(payload)
		if outputStarted || !retryable {
			s.pending = append(s.pending[:0], payload...)
			s.suppressFrame = true
			return true
		}
		s.suppressFrame = false
	}
	if openAIStreamEventTypeIsTerminal(kind) {
		s.pending = nil
		s.suppressFrame = false
	}
	return s.suppressFrame
}

func buildOpenAIResponseFailedSSE(responseID, model string, createdAt int64, source []byte, fallbackMessage string) string {
	responseID = strings.TrimSpace(responseID)
	if responseID == "" {
		responseID = "resp_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	}
	errorType := strings.TrimSpace(gjson.GetBytes(source, "error.type").String())
	if errorType == "" {
		errorType = strings.TrimSpace(gjson.GetBytes(source, "response.error.type").String())
	}
	code := strings.TrimSpace(gjson.GetBytes(source, "error.code").String())
	if code == "" {
		code = strings.TrimSpace(gjson.GetBytes(source, "response.error.code").String())
	}
	if code == "" {
		code = "upstream_error"
	}
	message := extractOpenAISSEErrorMessage(source)
	if message == "" {
		message = strings.TrimSpace(fallbackMessage)
	}
	if message == "" {
		message = "Upstream response failed"
	}
	errorBody := gin.H{"code": code, "message": message}
	if errorType != "" {
		errorBody["type"] = errorType
	}
	if createdAt <= 0 {
		createdAt = time.Now().Unix()
	}
	sequenceNumber := openAIResponsesFailedSequenceNumber(source)
	response := gin.H{
		"id":         responseID,
		"object":     "response",
		"created_at": createdAt,
		"status":     "failed",
		"output":     []any{},
		"error":      errorBody,
	}
	if model = strings.TrimSpace(model); model != "" {
		response["model"] = model
	}
	payload, err := marshalOpenAIUpstreamJSON(gin.H{
		"type":            "response.failed",
		"sequence_number": sequenceNumber,
		"response":        response,
	})
	if err != nil {
		// All values above are JSON primitives, so this is only a defensive fallback.
		payload = []byte(`{"type":"response.failed","sequence_number":` + strconv.FormatInt(sequenceNumber, 10) +
			`,"response":{"object":"response","created_at":` + strconv.FormatInt(createdAt, 10) +
			`,"status":"failed","output":[],"error":{"code":"upstream_error","message":"Upstream response failed"}}}`)
	}
	return "event: response.failed\ndata: " + string(payload) + "\n\n"
}
