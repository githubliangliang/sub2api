package apicompat

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestResponsesToChatCompletionsRequest_AgentMessageBecomesUserMessage(t *testing.T) {
	req := &ResponsesRequest{
		Model: "glm-5.3",
		Input: json.RawMessage(`[
			{"type":"message","role":"user","content":[{"type":"input_text","text":"<environment_context>cwd</environment_context>"}]},
			{"type":"agent_message","id":"amsg_1","content":[
				{"type":"input_text","text":"Message Type: NEW_TASK\nTask name: /root/alpha_task\nSender: /root\nPayload:\n"},
				{"type":"encrypted_content","encrypted_content":"Reply with the single word: ALPHA"}
			]},
			{"type":"reasoning","summary":[{"type":"summary_text","text":"just answer"}]},
			{"type":"message","role":"assistant","content":[{"type":"output_text","text":"ALPHA"}]},
			{"type":"agent_message","id":"amsg_2","content":[{"type":"input_text","text":"Message Type: FINAL_ANSWER\nTask name: /root\nSender: /root/alpha_task\nPayload:\nALPHA"}]}
		]`),
	}

	out, err := ResponsesToChatCompletionsRequest(req)
	require.NoError(t, err)
	require.Len(t, out.Messages, 4)
	require.Equal(t, "user", out.Messages[1].Role)
	require.JSONEq(t, `"Message Type: NEW_TASK\nTask name: /root/alpha_task\nSender: /root\nPayload:\nReply with the single word: ALPHA"`, string(out.Messages[1].Content))
	require.Equal(t, "assistant", out.Messages[2].Role)
	require.Equal(t, "user", out.Messages[3].Role)
}

func TestResponsesToChatCompletionsRequest_AgentMessageWithoutTextIsSkipped(t *testing.T) {
	req := &ResponsesRequest{
		Model: "glm-5.3",
		Input: json.RawMessage(`[
			{"type":"message","role":"user","content":[{"type":"input_text","text":"hi"}]},
			{"type":"agent_message","content":[]},
			{"type":"agent_message"},
			{"type":"agent_message","content":[{"type":"input_image","image_url":"data:image/png;base64,AA=="}]}
		]`),
	}

	out, err := ResponsesToChatCompletionsRequest(req)
	require.NoError(t, err)
	require.Len(t, out.Messages, 1)
}
