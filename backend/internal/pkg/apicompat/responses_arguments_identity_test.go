package apicompat

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBufferedResponseAccumulator_SupplementUsesCallIdentityBeforePosition(t *testing.T) {
	acc := NewBufferedResponseAccumulator()
	for index, callID := range []string{"call_a", "call_b"} {
		acc.ProcessEvent(&ResponsesStreamEvent{
			Type:        "response.output_item.added",
			OutputIndex: index,
			Item:        &ResponsesOutput{Type: "function_call", CallID: callID, Name: "read_file"},
		})
		acc.ProcessEvent(&ResponsesStreamEvent{
			Type:        "response.function_call_arguments.done",
			OutputIndex: index,
			Arguments:   []string{`{"path":"a.txt"}`, `{"path":"b.txt"}`}[index],
		})
	}

	response := &ResponsesResponse{Output: []ResponsesOutput{
		{Type: "function_call", CallID: "call_b", Name: "read_file"},
		{Type: "function_call", CallID: "call_a", Name: "read_file"},
	}}
	acc.SupplementResponseOutput(response)

	require.Equal(t, `{"path":"b.txt"}`, response.Output[0].Arguments)
	require.Equal(t, `{"path":"a.txt"}`, response.Output[1].Arguments)
}

func TestBufferedResponseAccumulator_SupplementRejectsConflictingCallIdentity(t *testing.T) {
	acc := NewBufferedResponseAccumulator()
	acc.ProcessEvent(&ResponsesStreamEvent{
		Type: "response.output_item.added",
		Item: &ResponsesOutput{Type: "function_call", CallID: "call_old", Name: "read_file"},
	})
	acc.ProcessEvent(&ResponsesStreamEvent{
		Type:      "response.function_call_arguments.done",
		Arguments: `{"path":"old.txt"}`,
	})

	response := &ResponsesResponse{Output: []ResponsesOutput{
		{Type: "function_call", CallID: "call_new", Name: "read_file"},
	}}
	acc.SupplementResponseOutput(response)

	require.Empty(t, response.Output[0].Arguments)
}

func TestBufferedResponseAccumulator_SupplementFallsBackToPositionWithoutIdentity(t *testing.T) {
	acc := NewBufferedResponseAccumulator()
	acc.ProcessEvent(&ResponsesStreamEvent{
		Type: "response.output_item.added",
		Item: &ResponsesOutput{Type: "function_call", CallID: "call_a", Name: "read_file"},
	})
	acc.ProcessEvent(&ResponsesStreamEvent{
		Type:      "response.function_call_arguments.done",
		Arguments: `{"path":"a.txt"}`,
	})
	response := &ResponsesResponse{Output: []ResponsesOutput{
		{Type: "function_call", Name: "read_file"},
	}}
	acc.SupplementResponseOutput(response)
	require.Equal(t, `{"path":"a.txt"}`, response.Output[0].Arguments)
}
