package service

import (
	"math"
	"strconv"
	"strings"
	"sync/atomic"

	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"
)

const responsesSequenceContextKey = "openai_responses_sequence"

type responsesSequenceState struct {
	last atomic.Int64
}

func newResponsesSequenceState() *responsesSequenceState {
	state := &responsesSequenceState{}
	state.last.Store(-1)
	return state
}

func (s *responsesSequenceState) observe(sequence int64) {
	for sequence >= 0 {
		last := s.last.Load()
		if sequence <= last || s.last.CompareAndSwap(last, sequence) {
			return
		}
	}
}

func (s *responsesSequenceState) next() int64 {
	for {
		last := s.last.Load()
		if last == math.MaxInt64 {
			return last
		}
		if s.last.CompareAndSwap(last, last+1) {
			return last + 1
		}
	}
}

func (s *responsesSequenceState) peek() int64 {
	last := s.last.Load()
	if last == math.MaxInt64 {
		return last
	}
	return last + 1
}

// The stream's consumer initializes this context state before writing events.
func responsesSequenceFromContext(c *gin.Context) *responsesSequenceState {
	if value, ok := c.Get(responsesSequenceContextKey); ok {
		if state, ok := value.(*responsesSequenceState); ok {
			return state
		}
	}
	state := newResponsesSequenceState()
	c.Set(responsesSequenceContextKey, state)
	return state
}

// ObserveOpenAIResponsesSequence records committed events without rewriting them.
func ObserveOpenAIResponsesSequence(c *gin.Context, sequence int64) {
	if c != nil && sequence >= 0 {
		responsesSequenceFromContext(c).observe(sequence)
	}
}

// NextOpenAIResponsesSequence reserves the next local event slot; heartbeat
// comments never consume slots. It is also used by the handler's final failure.
func NextOpenAIResponsesSequence(c *gin.Context) int64 {
	if c == nil {
		return 0
	}
	return responsesSequenceFromContext(c).next()
}

// SSE writers already own complete data lines. Read only their integer field,
// retaining no payload or additional stream buffer. Raw WS messages also work.
func openAIResponsesSequenceFromSSE(value string) int64 {
	maximum := int64(-1)
	for value != "" {
		var line string
		line, value, _ = strings.Cut(value, "\n")
		payload, dataLine := extractOpenAISSEDataLine(line)
		if !dataLine {
			payload = strings.TrimSpace(line)
			if !strings.HasPrefix(payload, "{") {
				continue
			}
		}
		field := gjson.Get(payload, "sequence_number")
		if field.Type != gjson.Number {
			continue
		}
		sequence, err := strconv.ParseInt(field.Raw, 10, 64)
		if err == nil && sequence > maximum {
			maximum = sequence
		}
	}
	return maximum
}
