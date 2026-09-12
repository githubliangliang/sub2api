package service

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

func TestOpenAIAPIKeyForwardPreservesNamespaceToolCalls(t *testing.T) {
	for _, passthrough := range []bool{false, true} {
		name := "responses"
		if passthrough {
			name = "passthrough"
		}
		t.Run(name, func(t *testing.T) {
			body := []byte(codexNamespaceRequestBody)
			account := newOpenAIRejectedFieldTestAccount()
			account.Extra["openai_passthrough"] = passthrough
			upstream := &httpUpstreamRecorder{responses: []*http.Response{
				newOpenAIRejectedFieldTestResponse(http.StatusOK, namespaceForwardOKResponse),
			}}

			result, err := newOpenAIRejectedFieldTestService(upstream).Forward(
				context.Background(), newOpenAIRejectedFieldTestContext(body), account, body,
			)

			require.NoError(t, err)
			require.NotNil(t, result)
			require.Len(t, upstream.bodies, 1)
			forwarded := upstream.bodies[0]
			require.Equal(t, "namespace", gjson.GetBytes(forwarded, "tools.0.type").String())
			require.Equal(t, "collaboration", gjson.GetBytes(forwarded, "input.0.namespace").String(),
				"a declared namespaced spawn_agent call must not become a default-namespace call")
			require.Equal(t, "spawn_agent", gjson.GetBytes(forwarded, "input.0.name").String())
			require.False(t, gjson.GetBytes(forwarded, "input.1.namespace").Exists())
			require.Equal(t, "collaboration", gjson.GetBytes(body, "input.0.namespace").String(), "caller history must remain immutable")
		})
	}
}

func TestOpenAIAPIKeyForwardPreservesHistoricalNamespaceContext(t *testing.T) {
	for name, field := range map[string]struct{ path, value string }{
		"previous_response":   {"previous_response_id", `"resp_previous"`},
		"conversation":        {"conversation", `"conv_previous"`},
		"conversation_object": {"conversation", `{"id":"conv_previous"}`},
		"additional_tools":    {"input.-1", `{"type":"additional_tools","tools":[{"type":"namespace","name":"collaboration","tools":[{"type":"function","name":"spawn_agent","parameters":{"type":"object"}}]}]}`},
		"discovered_tools":    {"input.-1", `{"type":"tool_search_output","tools":[{"type":"namespace","name":"collaboration","tools":[{"type":"function","name":"spawn_agent","parameters":{"type":"object"}}]}]}`},
	} {
		t.Run(name, func(t *testing.T) {
			body := []byte(`{"model":"gpt-5.5","stream":false,"input":[{"type":"function_call","name":"spawn_agent","namespace":"collaboration","call_id":"call_1","arguments":"{}","sequence":9007199254740993}]}`)
			body, err := sjson.SetRawBytes(body, field.path, []byte(field.value))
			require.NoError(t, err)
			upstream := &httpUpstreamRecorder{responses: []*http.Response{
				newOpenAIRejectedFieldTestResponse(http.StatusOK, namespaceForwardOKResponse),
			}}
			_, err = newOpenAIRejectedFieldTestService(upstream).Forward(
				context.Background(), newOpenAIRejectedFieldTestContext(body), newOpenAIRejectedFieldTestAccount(), body,
			)
			require.NoError(t, err)
			require.Len(t, upstream.bodies, 1)
			require.Equal(t, "collaboration", gjson.GetBytes(upstream.bodies[0], "input.0.namespace").String())
			require.Equal(t, "9007199254740993", gjson.GetBytes(upstream.bodies[0], "input.0.sequence").Raw)
		})
	}
}

func TestOpenAIAPIKeyForwardRetriesExplicitNamespaceRejection(t *testing.T) {
	body := []byte(codexNamespaceRequestBody)
	upstream := &httpUpstreamRecorder{responses: []*http.Response{
		newOpenAIRejectedFieldTestResponse(http.StatusBadRequest, `{"error":{"code":"unknown_parameter","message":"Unknown parameter: 'input[0].namespace'.","param":"input[0].namespace"}}`),
		newOpenAIRejectedFieldTestResponse(http.StatusOK, namespaceForwardOKResponse),
	}}
	_, err := newOpenAIRejectedFieldTestService(upstream).Forward(
		context.Background(), newOpenAIRejectedFieldTestContext(body), newOpenAIRejectedFieldTestAccount(), body,
	)
	require.NoError(t, err)
	require.Len(t, upstream.bodies, 2)
	require.Equal(t, "collaboration", gjson.GetBytes(upstream.bodies[0], "input.0.namespace").String())
	require.False(t, gjson.GetBytes(upstream.bodies[1], "input.0.namespace").Exists())
	require.Equal(t, gjson.GetBytes(upstream.bodies[0], "tools").Raw, gjson.GetBytes(upstream.bodies[1], "tools").Raw)
	require.Equal(t, "collaboration", gjson.GetBytes(body, "input.0.namespace").String())
}

func TestOpenAIAPIKeyCompactStillStripsToolCallNamespaces(t *testing.T) {
	body := []byte(codexNamespaceRequestBody)
	c := newOpenAIRejectedFieldTestContext(body)
	c.Request.URL.Path = "/v1/responses/compact"
	upstream := &httpUpstreamRecorder{responses: []*http.Response{
		newOpenAIRejectedFieldTestResponse(http.StatusOK, namespaceForwardOKResponse),
	}}
	_, err := newOpenAIRejectedFieldTestService(upstream).Forward(context.Background(), c, newOpenAIRejectedFieldTestAccount(), body)
	require.NoError(t, err)
	require.Len(t, upstream.bodies, 1)
	require.False(t, gjson.GetBytes(upstream.bodies[0], "input.0.namespace").Exists())
}
