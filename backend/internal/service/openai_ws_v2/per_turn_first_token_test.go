package openai_ws_v2

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestEveryTurnFirstTokenWithoutResponseID(t *testing.T) {
	state := &relayState{}
	start := time.Date(2026, 11, 10, 0, 0, 0, 0, time.UTC)
	for i, id := range []string{"first", "second"} {
		turnAt := start.Add(time.Duration(i) * time.Minute)
		now := turnAt
		clock := func() time.Time { return now }
		observeUpstreamMessage(state, []byte(`{"type":"response.created","response":{"id":"`+id+`"}}`), start, clock, nil)
		now = turnAt.Add(700 * time.Millisecond)
		observeUpstreamMessage(state, []byte(`{"type":"response.output_text.delta","delta":"hello"}`), start, clock, nil)
		now = turnAt.Add(time.Second)
		result := observeUpstreamMessage(state, []byte(`{"type":"response.completed","response":{"id":"`+id+`","usage":{"input_tokens":1,"output_tokens":1}}}`), start, clock, nil)
		require.NotNil(t, result.firstToken, id)
		require.Equal(t, 700, *result.firstToken, id)
	}
}

func TestPerTurnFirstTokenNegativeClockAndDuplicateTerminal(t *testing.T) {
	state := &relayState{}
	start := time.Now()
	now := start
	clock := func() time.Time { return now }
	observeUpstreamMessage(state, []byte(`{"type":"response.created","response":{"id":"turn"}}`), start, clock, nil)
	now = start.Add(-time.Millisecond)
	observeUpstreamMessage(state, []byte(`{"type":"response.output_text.delta","response_id":"turn","delta":"early"}`), start, clock, nil)
	require.Nil(t, state.activeTurn.firstTokenMs)
	now = start.Add(10 * time.Millisecond)
	observeUpstreamMessage(state, []byte(`{"type":"response.output_text.delta","response_id":"turn","delta":"valid"}`), start, clock, nil)
	terminal := []byte(`{"type":"response.completed","response":{"id":"turn","service_tier":"ultrafast","model":"gpt-6-astra","usage":{"input_tokens":2,"output_tokens":3}}}`)
	got := observeUpstreamMessage(state, terminal, start, clock, nil)
	require.NotNil(t, got.firstToken)
	require.Equal(t, 10, *got.firstToken)
	require.Equal(t, "ultrafast", got.responseServiceTier)
	require.False(t, observeUpstreamMessage(state, terminal, start, clock, nil).terminal)
}
